package application

import (
	"context"
	"fmt"
	"time"

	"github.com/tikiclone/tiki/platforms/live-commerce/internal/cache"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/domain"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/engagement"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/fanout"
	ch "github.com/tikiclone/tiki/platforms/live-commerce/internal/infrastructure/clickhouse"
	redi "github.com/tikiclone/tiki/platforms/live-commerce/internal/infrastructure/redis"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/metrics"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/moderation"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/recommendations"
	"github.com/tikiclone/tiki/platforms/live-commerce/internal/replay"
	"go.opentelemetry.io/otel"
)

type LiveCommerceService struct {
	livestreamRepo domain.LivestreamRepository
	messageRepo    domain.ChatMessageRepository
	reactionRepo   domain.ReactionRepository
	giftRepo       domain.GiftRepository
	pinnedRepo     domain.PinnedProductRepository
	moderationRepo domain.ModerationRepository
	publisher      EventPublisher
	redis          *redi.Store
	cache          *cache.Store
	fanout         *fanout.Broadcaster
	counters       *engagement.Counters
	modFilter      *moderation.Filter
	modQueue       *moderation.Queue
	replay         *replay.EventBuffer
	recEngine      *recommendations.Engine
	clickhouse     *ch.Conn
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
}

func NewLiveCommerceService(
	livestreamRepo domain.LivestreamRepository,
	messageRepo domain.ChatMessageRepository,
	reactionRepo domain.ReactionRepository,
	giftRepo domain.GiftRepository,
	pinnedRepo domain.PinnedProductRepository,
	moderationRepo domain.ModerationRepository,
	pub EventPublisher,
	redis *redi.Store,
	cache *cache.Store,
	fanout *fanout.Broadcaster,
	counters *engagement.Counters,
	modFilter *moderation.Filter,
	modQueue *moderation.Queue,
	replay *replay.EventBuffer,
	recEngine *recommendations.Engine,
	clickhouse *ch.Conn,
) *LiveCommerceService {
	return &LiveCommerceService{
		livestreamRepo: livestreamRepo,
		messageRepo:    messageRepo,
		reactionRepo:   reactionRepo,
		giftRepo:       giftRepo,
		pinnedRepo:     pinnedRepo,
		moderationRepo: moderationRepo,
		publisher:      pub,
		redis:          redis,
		cache:          cache,
		fanout:         fanout,
		counters:       counters,
		modFilter:      modFilter,
		modQueue:       modQueue,
		replay:         replay,
		recEngine:      recEngine,
		clickhouse:     clickhouse,
	}
}

func (s *LiveCommerceService) CreateLivestream(ctx context.Context, sellerID, title, description, coverURL, category string, tags []string, scheduledAt *time.Time) (*domain.Livestream, error) {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.CreateLivestream")
	defer span.End()

	ls := domain.NewLivestream(sellerID, title, description, coverURL, category, tags, scheduledAt)
	if err := s.livestreamRepo.Create(ctx, ls); err != nil {
		return nil, fmt.Errorf("create livestream: %w", err)
	}
	if s.cache != nil {
		s.cache.InvalidateLivestream(ctx, ls.ID)
	}
	metrics.LivestreamsCreated.Inc()
	s.publishEvent(ctx, domain.EventLivestreamCreated, &domain.LivestreamEventPayload{
		LivestreamID: ls.ID, SellerID: sellerID, Title: title, Timestamp: time.Now().UnixMilli(),
	})
	return ls, nil
}

func (s *LiveCommerceService) GetLivestream(ctx context.Context, id string) (*domain.Livestream, error) {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.GetLivestream")
	defer span.End()

	if s.cache != nil {
		if cached, _ := s.cache.GetLivestream(ctx, id); cached != nil {
			return cached, nil
		}
	}
	ls, err := s.livestreamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.SetLivestream(ctx, ls)
	}
	return ls, nil
}

func (s *LiveCommerceService) StartLivestream(ctx context.Context, livestreamID string) error {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.StartLivestream")
	defer span.End()

	ls, err := s.livestreamRepo.GetByID(ctx, livestreamID)
	if err != nil {
		return err
	}
	if err := ls.Start(); err != nil {
		return err
	}
	if err := s.livestreamRepo.Update(ctx, ls); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.InvalidateLivestream(ctx, livestreamID)
	}
	if s.redis != nil {
		s.redis.SetRoomStatus(ctx, livestreamID, domain.LiveStatusLive)
	}
	metrics.LivestreamsStarted.Inc()
	s.publishEvent(ctx, domain.EventLivestreamStarted, &domain.LivestreamEventPayload{
		LivestreamID: livestreamID, SellerID: ls.SellerID, Timestamp: time.Now().UnixMilli(),
	})
	if s.fanout != nil {
		s.fanout.Broadcast(ctx, livestreamID, "livestream.started", map[string]string{"id": livestreamID}, "")
	}
	return nil
}

func (s *LiveCommerceService) EndLivestream(ctx context.Context, livestreamID string) error {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.EndLivestream")
	defer span.End()

	ls, err := s.livestreamRepo.GetByID(ctx, livestreamID)
	if err != nil {
		return err
	}
	if err := ls.End(); err != nil {
		return err
	}
	if err := s.livestreamRepo.Update(ctx, ls); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.InvalidateLivestream(ctx, livestreamID)
	}
	if s.redis != nil {
		s.redis.SetRoomStatus(ctx, livestreamID, domain.LiveStatusEnded)
	}
	metrics.LivestreamsEnded.Inc()
	s.publishEvent(ctx, domain.EventLivestreamEnded, &domain.LivestreamEventPayload{
		LivestreamID: livestreamID, SellerID: ls.SellerID, Timestamp: time.Now().UnixMilli(),
	})
	if s.fanout != nil {
		s.fanout.Broadcast(ctx, livestreamID, "livestream.ended", map[string]string{"id": livestreamID}, "")
	}
	return nil
}

func (s *LiveCommerceService) ListActiveLivestreams(ctx context.Context, offset, limit int) ([]*domain.Livestream, int64, error) {
	return s.livestreamRepo.ListActive(ctx, offset, limit)
}

func (s *LiveCommerceService) ListSellerLivestreams(ctx context.Context, sellerID string, offset, limit int) ([]*domain.Livestream, int64, error) {
	return s.livestreamRepo.ListBySeller(ctx, sellerID, offset, limit)
}

func (s *LiveCommerceService) SendChatMessage(ctx context.Context, roomID, userID, username, content string) (*domain.ChatMessage, error) {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.SendChatMessage")
	defer span.End()

	if s.redis != nil {
		banned, err := s.redis.IsUserBanned(ctx, roomID, userID)
		if err == nil && banned {
			return nil, domain.ErrUserBanned
		}
		muted, err := s.redis.IsUserMuted(ctx, roomID, userID)
		if err == nil && muted {
			return nil, domain.ErrUserMuted
		}
	}

	valid, reason := s.modFilter.ValidateContent(content)
	if !valid {
		return nil, fmt.Errorf("content rejected: %s", reason)
	}

	seq, _ := s.messageRepo.GetLastSequence(ctx, roomID)
	seq++

	msg := domain.NewChatMessage(roomID, userID, username, content, domain.MsgTypeText)
	msg.Sequence = seq

	if err := s.messageRepo.Save(ctx, msg); err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	metrics.ChatMessagesTotal.Inc()
	s.publishEvent(ctx, domain.EventChatMessageSent, &domain.ChatMessageSentPayload{
		MessageID: msg.ID, RoomID: roomID, UserID: userID, Username: username,
		Content: content, Timestamp: time.Now().UnixMilli(),
	})

	if s.replay != nil {
		s.replay.Append(roomID, seq, "chat", map[string]interface{}{
			"id": msg.ID, "user_id": userID, "username": username, "content": content,
		})
	}

	if s.fanout != nil {
		s.fanout.Broadcast(ctx, roomID, "chat", map[string]interface{}{
			"id": msg.ID, "user_id": userID, "username": username,
			"content": content, "timestamp": msg.Timestamp.UnixMilli(),
		}, "")
	}

	if s.clickhouse != nil {
		s.clickhouse.InsertEngagementEvent(ctx, roomID, userID, "chat", 1, time.Now())
	}

	return msg, nil
}

func (s *LiveCommerceService) SendReaction(ctx context.Context, roomID, userID, reactionType string) error {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.SendReaction")
	defer span.End()

	validTypes := map[string]bool{"like": true, "love": true, "wow": true, "laugh": true, "sad": true, "angry": true}
	if !validTypes[reactionType] {
		return domain.ErrInvalidReaction
	}

	reaction := domain.NewReaction(roomID, userID, reactionType)
	if err := s.reactionRepo.Save(ctx, reaction); err != nil {
		return fmt.Errorf("save reaction: %w", err)
	}

	metrics.ReactionsTotal.WithLabelValues(reactionType).Inc()
	if s.counters != nil {
		s.counters.AddReaction(ctx, roomID, reactionType)
	}
	s.publishEvent(ctx, domain.EventReactionAdded, &domain.ReactionAddedPayload{
		ReactionID: reaction.ID, RoomID: roomID, UserID: userID, Type: reactionType,
	})

	if s.clickhouse != nil {
		s.clickhouse.InsertEngagementEvent(ctx, roomID, userID, "reaction_"+reactionType, 1, time.Now())
	}

	return nil
}

func (s *LiveCommerceService) SendGift(ctx context.Context, roomID, userID, username, giftType string, amount int64, currency string) error {
	_, span := otel.Tracer("tiki-live-commerce").Start(ctx, "service.SendGift")
	defer span.End()

	gift := &domain.Gift{
		ID:        fmt.Sprintf("gift_%d", time.Now().UnixNano()),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		GiftType:  giftType,
		Amount:    amount,
		Currency:  currency,
		Timestamp: time.Now(),
	}
	if err := s.giftRepo.Save(ctx, gift); err != nil {
		return fmt.Errorf("save gift: %w", err)
	}

	metrics.GiftsTotal.WithLabelValues(giftType).Inc()
	if s.counters != nil {
		s.counters.AddGift(ctx, roomID, amount)
	}
	if s.redis != nil {
		s.redis.AddToGiftLeaderboard(ctx, roomID, userID, username, amount)
	}
	s.publishEvent(ctx, domain.EventGiftSent, &domain.GiftSentPayload{
		GiftID: gift.ID, RoomID: roomID, UserID: userID, GiftType: giftType, Amount: amount,
	})

	if s.fanout != nil {
		s.fanout.Broadcast(ctx, roomID, "gift", map[string]interface{}{
			"user_id": userID, "username": username, "gift_type": giftType,
			"amount": amount, "timestamp": time.Now().UnixMilli(),
		}, "")
	}

	if s.clickhouse != nil {
		s.clickhouse.InsertEngagementEvent(ctx, roomID, userID, "gift", amount, time.Now())
	}

	return nil
}

func (s *LiveCommerceService) PinProduct(ctx context.Context, livestreamID, productID, productName string, price int64, imageURL string) (*domain.PinnedProduct, error) {
	pp := &domain.PinnedProduct{
		ID:           fmt.Sprintf("pin_%d", time.Now().UnixNano()),
		LivestreamID: livestreamID,
		ProductID:    productID,
		ProductName:  productName,
		Price:        price,
		ImageURL:     imageURL,
		IsActive:     true,
		PinnedAt:     time.Now(),
	}
	if err := s.pinnedRepo.Pin(ctx, pp); err != nil {
		return nil, fmt.Errorf("pin product: %w", err)
	}
	s.publishEvent(ctx, domain.EventProductPinned, map[string]string{
		"livestream_id": livestreamID, "product_id": productID,
	})
	if s.fanout != nil {
		s.fanout.Broadcast(ctx, livestreamID, "product.pinned", pp, "")
	}
	return pp, nil
}

func (s *LiveCommerceService) UnpinProduct(ctx context.Context, livestreamID, productID string) error {
	if err := s.pinnedRepo.Unpin(ctx, livestreamID, productID); err != nil {
		return err
	}
	s.publishEvent(ctx, domain.EventProductUnpinned, map[string]string{
		"livestream_id": livestreamID, "product_id": productID,
	})
	if s.fanout != nil {
		s.fanout.Broadcast(ctx, livestreamID, "product.unpinned", map[string]string{"product_id": productID}, "")
	}
	return nil
}

func (s *LiveCommerceService) GetPinnedProducts(ctx context.Context, livestreamID string) ([]*domain.PinnedProduct, error) {
	return s.pinnedRepo.GetActiveByLivestream(ctx, livestreamID)
}

func (s *LiveCommerceService) GetViewerCount(ctx context.Context, livestreamID string) (int64, error) {
	if s.counters != nil {
		return s.counters.GetViewerCount(ctx, livestreamID), nil
	}
	return 0, nil
}

func (s *LiveCommerceService) GetReactionSummary(ctx context.Context, roomID string) (map[string]int64, error) {
	if s.counters != nil {
		return s.counters.GetReactionCounts(ctx, roomID), nil
	}
	return make(map[string]int64), nil
}

func (s *LiveCommerceService) GetGiftLeaderboard(ctx context.Context, roomID string, limit int) ([]*domain.GiftLeaderboardEntry, error) {
	return s.giftRepo.GetLeaderboardByRoom(ctx, roomID, limit)
}

func (s *LiveCommerceService) GetChatHistory(ctx context.Context, roomID string, offset, limit int) ([]*domain.ChatMessage, int64, error) {
	return s.messageRepo.GetByRoom(ctx, roomID, offset, limit)
}

func (s *LiveCommerceService) ModerateAction(ctx context.Context, roomID, userID, action, reason, moderatedBy string, durationSec int64) error {
	modAction := &domain.ModerationAction{
		ID:          fmt.Sprintf("mod_%d", time.Now().UnixNano()),
		RoomID:      roomID,
		UserID:      userID,
		Action:      action,
		Reason:      reason,
		ModeratedBy: moderatedBy,
		DurationSec: durationSec,
		CreatedAt:   time.Now(),
	}
	if err := s.moderationRepo.SaveAction(ctx, modAction); err != nil {
		return err
	}

	switch action {
	case domain.ModActionMute:
		if s.redis != nil {
			s.redis.SetUserMuted(ctx, roomID, userID, time.Duration(durationSec)*time.Second)
		}
	case domain.ModActionBan:
		if s.redis != nil {
			s.redis.SetUserBanned(ctx, roomID, userID)
		}
	case domain.ModActionRemove:
		s.messageRepo.MarkModerated(ctx, reason)
	}

	s.publishEvent(ctx, domain.EventModerationAction, &domain.ModerationActionPayload{
		ActionID: modAction.ID, RoomID: roomID, UserID: userID,
		Action: action, Reason: reason, ModeratedBy: moderatedBy,
	})

	if s.fanout != nil {
		s.fanout.Broadcast(ctx, roomID, "moderation", modAction, "")
	}
	return nil
}

func (s *LiveCommerceService) GetTrendingLivestreams(ctx context.Context, limit int) []*recommendations.TrendingScore {
	if s.recEngine != nil {
		return s.recEngine.GetTrending(ctx, limit)
	}
	return nil
}

func (s *LiveCommerceService) GetReplayEvents(ctx context.Context, roomID string, sinceSeq int64) []*replay.ReplayEvent {
	if s.replay != nil {
		return s.replay.GetSince(roomID, sinceSeq)
	}
	return nil
}

func (s *LiveCommerceService) HandleViewerJoined(ctx context.Context, roomID, userID string) {
	if s.counters != nil {
		s.counters.AddViewer(ctx, roomID, userID)
	}
	if s.clickhouse != nil {
		s.clickhouse.InsertViewerEvent(ctx, roomID, userID, "join", time.Now())
	}
	s.publishEvent(ctx, domain.EventViewerJoined, map[string]string{"room_id": roomID, "user_id": userID})
}

func (s *LiveCommerceService) HandleViewerLeft(ctx context.Context, roomID, userID string) {
	if s.counters != nil {
		s.counters.RemoveViewer(ctx, roomID, userID)
	}
	if s.clickhouse != nil {
		s.clickhouse.InsertViewerEvent(ctx, roomID, userID, "leave", time.Now())
	}
	s.publishEvent(ctx, domain.EventViewerLeft, map[string]string{"room_id": roomID, "user_id": userID})
}

func (s *LiveCommerceService) UpdateTrending(ctx context.Context) {
	streams, _, _ := s.livestreamRepo.ListActive(ctx, 0, 100)
	if s.recEngine != nil {
		s.recEngine.UpdateTrending(ctx, streams)
	}
}

func (s *LiveCommerceService) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if s.publisher != nil {
		s.publisher.Publish(ctx, eventType, payload)
	}
}

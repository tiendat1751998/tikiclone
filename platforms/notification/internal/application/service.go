package application
import ("context"; "fmt"; "time"; "github.com/tikiclone/tiki/platforms/notification/internal/domain"; "github.com/tikiclone/tiki/platforms/notification/internal/infrastructure/redis"; "github.com/tikiclone/tiki/platforms/notification/internal/metrics"; "go.opentelemetry.io/otel"; "go.opentelemetry.io/otel/attribute")

type NotificationService struct { notifyRepo domain.NotificationRepository; templateRepo domain.TemplateRepository; prefRepo domain.PreferenceRepository; deliveryRepo domain.DeliveryLogRepository; redis *redis.Store; publisher EventPublisher }
type EventPublisher interface { Publish(ctx context.Context, eventType string, payload interface{}) error }

func NewNotificationService(nr domain.NotificationRepository, tr domain.TemplateRepository, pr domain.PreferenceRepository, dr domain.DeliveryLogRepository, rs *redis.Store, pub EventPublisher) *NotificationService {
	return &NotificationService{notifyRepo: nr, templateRepo: tr, prefRepo: pr, deliveryRepo: dr, redis: rs, publisher: pub}
}

func (s *NotificationService) SendNotification(ctx context.Context, userID, nType, title, body, channel string, data map[string]interface{}) (*domain.Notification, error) {
	ctx, span := otel.Tracer("tiki-notification").Start(ctx, "notify.send"); defer span.End()
	span.SetAttributes(attribute.String("user_id", userID), attribute.String("channel", channel))

	// Check rate limit
	if ok, _ := s.redis.CheckRateLimit(ctx, fmt.Sprintf("%s:%s", userID, channel), 10); !ok {
		metrics.RateLimitHits.Inc()
		return nil, fmt.Errorf("rate limit exceeded for user %s channel %s", userID, channel)
	}

	n := &domain.Notification{
		ID: fmt.Sprintf("notif_%d", time.Now().UnixNano()), UserID: userID, Type: nType,
		Title: title, Body: body, Channel: channel, Status: domain.NotifyStatusPending,
		Priority: 0, CreatedAt: time.Now(),
	}

	if err := s.notifyRepo.Create(ctx, n); err != nil { return nil, err }
	s.redis.IncrementUnread(ctx, userID)
	metrics.NotificationsSent.Inc()

	if s.publisher != nil { s.publisher.Publish(ctx, "notification.sent", n) }
	return n, nil
}

func (s *NotificationService) GetInbox(ctx context.Context, userID string, offset, limit int) ([]*domain.Notification, int64, error) {
	return s.notifyRepo.FindByUserID(ctx, userID, offset, limit)
}

func (s *NotificationService) MarkRead(ctx context.Context, notificationID, userID string) error {
	n, err := s.notifyRepo.FindByID(ctx, notificationID)
	if err != nil { return err }
	if n == nil { return domain.ErrNotificationNotFound }
	if err := s.notifyRepo.MarkRead(ctx, notificationID); err != nil { return err }
	s.redis.DecrementUnread(ctx, userID)
	return nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	count, err := s.redis.GetUnreadCount(ctx, userID)
	if err != nil { return 0, err }
	return count, nil
}

func (s *NotificationService) UpdatePreference(ctx context.Context, userID, channel string, enabled bool) error {
	return s.prefRepo.UpdatePreference(ctx, &domain.UserPreference{UserID: userID, Channel: channel, Enabled: enabled})
}

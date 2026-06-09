package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tikiclone/tiki/services/notification-campaign/internal/domain"
)

type Repository interface {
	CreateCampaign(ctx context.Context, c *domain.NotificationCampaign) error
	GetCampaign(ctx context.Context, id string) (*domain.NotificationCampaign, error)
	ListCampaigns(ctx context.Context, status, campaignType string, limit, offset int) ([]domain.NotificationCampaign, error)
	UpdateCampaign(ctx context.Context, c *domain.NotificationCampaign) error
	DeleteCampaign(ctx context.Context, id string) error
	UpdateCampaignStatus(ctx context.Context, id string, status domain.CampaignStatus) error
	UpdateCampaignStats(ctx context.Context, id string, sent, delivered, opened, clicked, bounced, unsubscribed int) error

	CreateTemplate(ctx context.Context, t *domain.NotificationTemplate) error
	GetTemplate(ctx context.Context, id string) (*domain.NotificationTemplate, error)
	ListTemplates(ctx context.Context, campaignID string, channel string) ([]domain.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, t *domain.NotificationTemplate) error

	AddRecipient(ctx context.Context, r *domain.CampaignRecipient) error
	ListRecipients(ctx context.Context, campaignID string, status string, limit, offset int) ([]domain.CampaignRecipient, error)
	UpdateRecipientStatus(ctx context.Context, id string, status domain.RecipientStatus, msg *string) error

	GetPreference(ctx context.Context, userID string) (*domain.NotificationPreference, error)
	UpsertPreference(ctx context.Context, p *domain.NotificationPreference) error

	CreateDeliveryLog(ctx context.Context, l *domain.NotificationDeliveryLog) error
	GetDeliveryLog(ctx context.Context, id string) (*domain.NotificationDeliveryLog, error)
	ListDeliveryLogs(ctx context.Context, campaignID, userID string, limit, offset int) ([]domain.NotificationDeliveryLog, error)
}

type NotificationService struct {
	repo Repository
}

func NewNotificationService(repo Repository) *NotificationService {
	return &NotificationService{repo: repo}
}

type CreateCampaignRequest struct {
	Name            string            `json:"name" binding:"required"`
	Description     *string           `json:"description"`
	CampaignType    domain.CampaignType `json:"campaign_type" binding:"required"`
	Channel         domain.Channel    `json:"channel" binding:"required"`
	Priority        domain.Priority   `json:"priority"`
	TargetAudience  *string           `json:"target_audience"`
	SegmentCriteria *string           `json:"segment_criteria"`
	ScheduledAt     *time.Time        `json:"scheduled_at"`
	CreatedBy       *string           `json:"created_by"`
	Metadata        *string           `json:"metadata"`
}

func (s *NotificationService) CreateCampaign(ctx context.Context, req *CreateCampaignRequest) (*domain.NotificationCampaign, error) {
	now := time.Now().UTC()
	p := req.Priority
	if p == "" {
		p = domain.PriorityNormal
	}
	c := &domain.NotificationCampaign{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		CampaignType:    req.CampaignType,
		Channel:         req.Channel,
		Status:          domain.CampaignDraft,
		Priority:        p,
		TargetAudience:  req.TargetAudience,
		SegmentCriteria: req.SegmentCriteria,
		ScheduledAt:     req.ScheduledAt,
		CreatedBy:       req.CreatedBy,
		Metadata:        req.Metadata,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateCampaign(ctx, c); err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}
	return c, nil
}

func (s *NotificationService) GetCampaign(ctx context.Context, id string) (*domain.NotificationCampaign, error) {
	return s.repo.GetCampaign(ctx, id)
}

func (s *NotificationService) ListCampaigns(ctx context.Context, status, campaignType string, limit, offset int) ([]domain.NotificationCampaign, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListCampaigns(ctx, status, campaignType, limit, offset)
}

type UpdateCampaignRequest struct {
	Name            *string            `json:"name"`
	Description     *string            `json:"description"`
	CampaignType    *domain.CampaignType `json:"campaign_type"`
	Channel         *domain.Channel    `json:"channel"`
	Priority        *domain.Priority   `json:"priority"`
	TargetAudience  *string            `json:"target_audience"`
	SegmentCriteria *string            `json:"segment_criteria"`
	ScheduledAt     *time.Time         `json:"scheduled_at"`
	Metadata        *string            `json:"metadata"`
}

func (s *NotificationService) UpdateCampaign(ctx context.Context, id string, req *UpdateCampaignRequest) (*domain.NotificationCampaign, error) {
	c, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.CampaignDraft {
		return nil, domain.ErrCampaignNotDraft
	}
	if req.Name != nil { c.Name = *req.Name }
	if req.Description != nil { c.Description = req.Description }
	if req.CampaignType != nil { c.CampaignType = *req.CampaignType }
	if req.Channel != nil { c.Channel = *req.Channel }
	if req.Priority != nil { c.Priority = *req.Priority }
	if req.TargetAudience != nil { c.TargetAudience = req.TargetAudience }
	if req.SegmentCriteria != nil { c.SegmentCriteria = req.SegmentCriteria }
	if req.ScheduledAt != nil { c.ScheduledAt = req.ScheduledAt }
	if req.Metadata != nil { c.Metadata = req.Metadata }
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateCampaign(ctx, c); err != nil {
		return nil, fmt.Errorf("update campaign: %w", err)
	}
	return c, nil
}

func (s *NotificationService) DeleteCampaign(ctx context.Context, id string) error {
	return s.repo.DeleteCampaign(ctx, id)
}

func (s *NotificationService) LaunchCampaign(ctx context.Context, id string) (*domain.NotificationCampaign, error) {
	c, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.CampaignDraft && c.Status != domain.CampaignPaused {
		return nil, domain.ErrInvalidStatus
	}
	now := time.Now().UTC()
	c.Status = domain.CampaignActive
	c.StartedAt = &now
	c.UpdatedAt = now
	if err := s.repo.UpdateCampaignStatus(ctx, id, domain.CampaignActive); err != nil {
		return nil, fmt.Errorf("launch campaign: %w", err)
	}
	return c, nil
}

func (s *NotificationService) PauseCampaign(ctx context.Context, id string) (*domain.NotificationCampaign, error) {
	c, err := s.repo.GetCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.CampaignActive {
		return nil, domain.ErrCampaignNotActive
	}
	c.Status = domain.CampaignPaused
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateCampaignStatus(ctx, id, domain.CampaignPaused); err != nil {
		return nil, fmt.Errorf("pause campaign: %w", err)
	}
	return c, nil
}

type CreateTemplateRequest struct {
	CampaignID     *string                `json:"campaign_id"`
	Name           string                 `json:"name" binding:"required"`
	Channel        domain.TemplateChannel `json:"channel" binding:"required"`
	Subject        *string                `json:"subject"`
	BodyHTML       *string                `json:"body_html"`
	BodyText       *string                `json:"body_text"`
	BodyPush       *string                `json:"body_push"`
	VariablesSchema *string               `json:"variables_schema"`
	Locale         string                 `json:"locale"`
}

func (s *NotificationService) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*domain.NotificationTemplate, error) {
	now := time.Now().UTC()
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}
	t := &domain.NotificationTemplate{
		ID:              uuid.New().String(),
		CampaignID:      req.CampaignID,
		Name:            req.Name,
		Channel:         req.Channel,
		Subject:         req.Subject,
		BodyHTML:        req.BodyHTML,
		BodyText:        req.BodyText,
		BodyPush:        req.BodyPush,
		VariablesSchema: req.VariablesSchema,
		Locale:          locale,
		Version:         1,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return t, nil
}

func (s *NotificationService) GetTemplate(ctx context.Context, id string) (*domain.NotificationTemplate, error) {
	return s.repo.GetTemplate(ctx, id)
}

func (s *NotificationService) ListTemplates(ctx context.Context, campaignID, channel string) ([]domain.NotificationTemplate, error) {
	return s.repo.ListTemplates(ctx, campaignID, channel)
}

type UpdateTemplateRequest struct {
	Name           *string                `json:"name"`
	Channel        *domain.TemplateChannel `json:"channel"`
	Subject        *string                `json:"subject"`
	BodyHTML       *string                `json:"body_html"`
	BodyText       *string                `json:"body_text"`
	BodyPush       *string                `json:"body_push"`
	VariablesSchema *string               `json:"variables_schema"`
	Locale         *string                `json:"locale"`
	IsActive       *bool                  `json:"is_active"`
}

func (s *NotificationService) UpdateTemplate(ctx context.Context, id string, req *UpdateTemplateRequest) (*domain.NotificationTemplate, error) {
	t, err := s.repo.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil { t.Name = *req.Name }
	if req.Channel != nil { t.Channel = *req.Channel }
	if req.Subject != nil { t.Subject = req.Subject }
	if req.BodyHTML != nil { t.BodyHTML = req.BodyHTML }
	if req.BodyText != nil { t.BodyText = req.BodyText }
	if req.BodyPush != nil { t.BodyPush = req.BodyPush }
	if req.VariablesSchema != nil { t.VariablesSchema = req.VariablesSchema }
	if req.Locale != nil { t.Locale = *req.Locale }
	if req.IsActive != nil { t.IsActive = *req.IsActive }
	t.Version++
	t.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}
	return t, nil
}

type AddRecipientRequest struct {
	UserID  string         `json:"user_id" binding:"required"`
	Channel domain.Channel `json:"channel" binding:"required"`
}

func (s *NotificationService) AddRecipient(ctx context.Context, campaignID string, req *AddRecipientRequest) (*domain.CampaignRecipient, error) {
	now := time.Now().UTC()
	r := &domain.CampaignRecipient{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		UserID:     req.UserID,
		Channel:    req.Channel,
		Status:     domain.RecipientPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.AddRecipient(ctx, r); err != nil {
		return nil, fmt.Errorf("add recipient: %w", err)
	}
	return r, nil
}

func (s *NotificationService) ListRecipients(ctx context.Context, campaignID, status string, limit, offset int) ([]domain.CampaignRecipient, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListRecipients(ctx, campaignID, status, limit, offset)
}

type UpdatePreferenceRequest struct {
	EmailEnabled      *bool   `json:"email_enabled"`
	SMSEnabled        *bool   `json:"sms_enabled"`
	PushEnabled       *bool   `json:"push_enabled"`
	InAppEnabled      *bool   `json:"in_app_enabled"`
	WhatsAppEnabled   *bool   `json:"whatsapp_enabled"`
	PromotionalEmail  *bool   `json:"promotional_email"`
	TransactionalEmail *bool  `json:"transactional_email"`
	OrderUpdatesSMS   *bool   `json:"order_updates_sms"`
	PriceAlertsPush   *bool   `json:"price_alerts_push"`
	QuietHoursStart   *string `json:"quiet_hours_start"`
	QuietHoursEnd     *string `json:"quiet_hours_end"`
	Timezone          *string `json:"timezone"`
}

func (s *NotificationService) GetPreference(ctx context.Context, userID string) (*domain.NotificationPreference, error) {
	p, err := s.repo.GetPreference(ctx, userID)
	if err != nil {
		return s.repo.UpsertPreference(ctx, &domain.NotificationPreference{
			ID:                uuid.New().String(),
			UserID:            userID,
			EmailEnabled:      true,
			SMSEnabled:        true,
			PushEnabled:       true,
			InAppEnabled:      true,
			WhatsAppEnabled:   false,
			PromotionalEmail:  true,
			TransactionalEmail: true,
			OrderUpdatesSMS:   true,
			PriceAlertsPush:   true,
			Timezone:          "UTC",
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		})
	}
	return p, nil
}

func (s *NotificationService) UpdatePreference(ctx context.Context, userID string, req *UpdatePreferenceRequest) (*domain.NotificationPreference, error) {
	p, err := s.repo.GetPreference(ctx, userID)
	if err != nil {
		now := time.Now().UTC()
		p = &domain.NotificationPreference{
			ID:   uuid.New().String(),
			UserID: userID,
			EmailEnabled:      true,
			SMSEnabled:        true,
			PushEnabled:       true,
			InAppEnabled:      true,
			WhatsAppEnabled:   false,
			PromotionalEmail:  true,
			TransactionalEmail: true,
			OrderUpdatesSMS:   true,
			PriceAlertsPush:   true,
			Timezone:          "UTC",
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	}
	if req.EmailEnabled != nil { p.EmailEnabled = *req.EmailEnabled }
	if req.SMSEnabled != nil { p.SMSEnabled = *req.SMSEnabled }
	if req.PushEnabled != nil { p.PushEnabled = *req.PushEnabled }
	if req.InAppEnabled != nil { p.InAppEnabled = *req.InAppEnabled }
	if req.WhatsAppEnabled != nil { p.WhatsAppEnabled = *req.WhatsAppEnabled }
	if req.PromotionalEmail != nil { p.PromotionalEmail = *req.PromotionalEmail }
	if req.TransactionalEmail != nil { p.TransactionalEmail = *req.TransactionalEmail }
	if req.OrderUpdatesSMS != nil { p.OrderUpdatesSMS = *req.OrderUpdatesSMS }
	if req.PriceAlertsPush != nil { p.PriceAlertsPush = *req.PriceAlertsPush }
	if req.QuietHoursStart != nil { p.QuietHoursStart = req.QuietHoursStart }
	if req.QuietHoursEnd != nil { p.QuietHoursEnd = req.QuietHoursEnd }
	if req.Timezone != nil { p.Timezone = *req.Timezone }
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpsertPreference(ctx, p); err != nil {
		return nil, fmt.Errorf("update preference: %w", err)
	}
	return p, nil
}

func (s *NotificationService) GetDeliveryLog(ctx context.Context, id string) (*domain.NotificationDeliveryLog, error) {
	return s.repo.GetDeliveryLog(ctx, id)
}

func (s *NotificationService) ListDeliveryLogs(ctx context.Context, campaignID, userID string, limit, offset int) ([]domain.NotificationDeliveryLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListDeliveryLogs(ctx, campaignID, userID, limit, offset)
}

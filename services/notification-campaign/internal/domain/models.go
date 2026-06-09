package domain

import "time"

type CampaignType string

const (
	CampaignPromotional    CampaignType = "promotional"
	CampaignTransactional  CampaignType = "transactional"
	CampaignSystem         CampaignType = "system"
	CampaignRetargeting    CampaignType = "retargeting"
	CampaignAbandonedCart  CampaignType = "abandoned_cart"
	CampaignPriceDrop      CampaignType = "price_drop"
	CampaignBackInStock    CampaignType = "back_in_stock"
)

type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelPush     Channel = "push"
	ChannelInApp    Channel = "in_app"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelAll      Channel = "all"
)

type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "draft"
	CampaignScheduled CampaignStatus = "scheduled"
	CampaignActive    CampaignStatus = "active"
	CampaignPaused    CampaignStatus = "paused"
	CampaignCompleted CampaignStatus = "completed"
	CampaignCancelled CampaignStatus = "cancelled"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type NotificationCampaign struct {
	ID               string         `db:"id" json:"id"`
	Name             string         `db:"name" json:"name"`
	Description      *string        `db:"description" json:"description,omitempty"`
	CampaignType     CampaignType   `db:"campaign_type" json:"campaign_type"`
	Channel          Channel        `db:"channel" json:"channel"`
	Status           CampaignStatus `db:"status" json:"status"`
	Priority         Priority       `db:"priority" json:"priority"`
	TargetAudience   *string        `db:"target_audience" json:"target_audience,omitempty"`
	SegmentCriteria  *string        `db:"segment_criteria" json:"segment_criteria,omitempty"`
	ScheduledAt      *time.Time     `db:"scheduled_at" json:"scheduled_at,omitempty"`
	StartedAt        *time.Time     `db:"started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
	TotalRecipients  int            `db:"total_recipients" json:"total_recipients"`
	SentCount        int            `db:"sent_count" json:"sent_count"`
	DeliveredCount   int            `db:"delivered_count" json:"delivered_count"`
	OpenedCount      int            `db:"opened_count" json:"opened_count"`
	ClickedCount     int            `db:"clicked_count" json:"clicked_count"`
	BouncedCount     int            `db:"bounced_count" json:"bounced_count"`
	UnsubscribedCount int           `db:"unsubscribed_count" json:"unsubscribed_count"`
	CreatedBy        *string        `db:"created_by" json:"created_by,omitempty"`
	ApprovedBy       *string        `db:"approved_by" json:"approved_by,omitempty"`
	Metadata         *string        `db:"metadata" json:"metadata,omitempty"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt        *time.Time     `db:"deleted_at" json:"deleted_at,omitempty"`
}

type TemplateChannel string

const (
	TemplateChannelEmail    TemplateChannel = "email"
	TemplateChannelSMS      TemplateChannel = "sms"
	TemplateChannelPush     TemplateChannel = "push"
	TemplateChannelInApp    TemplateChannel = "in_app"
	TemplateChannelWhatsApp TemplateChannel = "whatsapp"
)

type NotificationTemplate struct {
	ID             string          `db:"id" json:"id"`
	CampaignID     *string         `db:"campaign_id" json:"campaign_id,omitempty"`
	Name           string          `db:"name" json:"name"`
	Channel        TemplateChannel `db:"channel" json:"channel"`
	Subject        *string         `db:"subject" json:"subject,omitempty"`
	BodyHTML       *string         `db:"body_html" json:"body_html,omitempty"`
	BodyText       *string         `db:"body_text" json:"body_text,omitempty"`
	BodyPush       *string         `db:"body_push" json:"body_push,omitempty"`
	VariablesSchema *string        `db:"variables_schema" json:"variables_schema,omitempty"`
	Locale         string          `db:"locale" json:"locale"`
	Version        int             `db:"version" json:"version"`
	IsActive       bool            `db:"is_active" json:"is_active"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}

type RecipientStatus string

const (
	RecipientPending     RecipientStatus = "pending"
	RecipientSent        RecipientStatus = "sent"
	RecipientDelivered   RecipientStatus = "delivered"
	RecipientOpened      RecipientStatus = "opened"
	RecipientClicked     RecipientStatus = "clicked"
	RecipientBounced     RecipientStatus = "bounced"
	RecipientFailed      RecipientStatus = "failed"
	RecipientUnsubscribed RecipientStatus = "unsubscribed"
)

type CampaignRecipient struct {
	ID           string          `db:"id" json:"id"`
	CampaignID   string          `db:"campaign_id" json:"campaign_id"`
	UserID       string          `db:"user_id" json:"user_id"`
	Channel      Channel         `db:"channel" json:"channel"`
	Status       RecipientStatus `db:"status" json:"status"`
	SentAt       *time.Time      `db:"sent_at" json:"sent_at,omitempty"`
	DeliveredAt  *time.Time      `db:"delivered_at" json:"delivered_at,omitempty"`
	OpenedAt     *time.Time      `db:"opened_at" json:"opened_at,omitempty"`
	ClickedAt    *time.Time      `db:"clicked_at" json:"clicked_at,omitempty"`
	ErrorMessage *string         `db:"error_message" json:"error_message,omitempty"`
	Metadata     *string         `db:"metadata" json:"metadata,omitempty"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
}

type NotificationPreference struct {
	ID                string     `db:"id" json:"id"`
	UserID            string     `db:"user_id" json:"user_id"`
	EmailEnabled      bool       `db:"email_enabled" json:"email_enabled"`
	SMSEnabled        bool       `db:"sms_enabled" json:"sms_enabled"`
	PushEnabled       bool       `db:"push_enabled" json:"push_enabled"`
	InAppEnabled      bool       `db:"in_app_enabled" json:"in_app_enabled"`
	WhatsAppEnabled   bool       `db:"whatsapp_enabled" json:"whatsapp_enabled"`
	PromotionalEmail  bool       `db:"promotional_email" json:"promotional_email"`
	TransactionalEmail bool      `db:"transactional_email" json:"transactional_email"`
	OrderUpdatesSMS   bool       `db:"order_updates_sms" json:"order_updates_sms"`
	PriceAlertsPush   bool       `db:"price_alerts_push" json:"price_alerts_push"`
	QuietHoursStart   *string    `db:"quiet_hours_start" json:"quiet_hours_start,omitempty"`
	QuietHoursEnd     *string    `db:"quiet_hours_end" json:"quiet_hours_end,omitempty"`
	Timezone          string     `db:"timezone" json:"timezone"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

type DeliveryStatus string

const (
	DeliveryQueued     DeliveryStatus = "queued"
	DeliverySent       DeliveryStatus = "sent"
	DeliveryDelivered  DeliveryStatus = "delivered"
	DeliveryOpened     DeliveryStatus = "opened"
	DeliveryClicked    DeliveryStatus = "clicked"
	DeliveryBounced    DeliveryStatus = "bounced"
	DeliveryFailed     DeliveryStatus = "failed"
	DeliverySuppressed DeliveryStatus = "suppressed"
)

type NotificationDeliveryLog struct {
	ID               string         `db:"id" json:"id"`
	CampaignID       *string        `db:"campaign_id" json:"campaign_id,omitempty"`
	TemplateID       *string        `db:"template_id" json:"template_id,omitempty"`
	UserID           string         `db:"user_id" json:"user_id"`
	Channel          Channel        `db:"channel" json:"channel"`
	Status           DeliveryStatus `db:"status" json:"status"`
	Provider         *string        `db:"provider" json:"provider,omitempty"`
	ProviderMessageID *string       `db:"provider_message_id" json:"provider_message_id,omitempty"`
	ErrorCode        *string        `db:"error_code" json:"error_code,omitempty"`
	ErrorMessage     *string        `db:"error_message" json:"error_message,omitempty"`
	RetryCount       int            `db:"retry_count" json:"retry_count"`
	Metadata         *string        `db:"metadata" json:"metadata,omitempty"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
}

var (
	ErrNotFound             = NewError("not found")
	ErrCampaignNotFound     = NewError("campaign not found")
	ErrTemplateNotFound     = NewError("template not found")
	ErrPreferenceNotFound   = NewError("preference not found")
	ErrDeliveryLogNotFound  = NewError("delivery log not found")
	ErrInvalidStatus        = NewError("invalid status transition")
	ErrCampaignNotDraft     = NewError("campaign is not in draft status")
	ErrCampaignNotActive    = NewError("campaign is not active")
	ErrUnauthorized         = NewError("unauthorized")
	ErrInvalidInput         = NewError("invalid input")
)

type Error struct {
	msg string
}

func NewError(msg string) error {
	return &Error{msg: msg}
}

func (e *Error) Error() string {
	return e.msg
}

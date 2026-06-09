package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tikiclone/tiki/services/notification-campaign/internal/config"
	"github.com/tikiclone/tiki/services/notification-campaign/internal/domain"
)

type Repository struct {
	db *sqlx.DB
}

func NewDB(cfg config.MySQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

var campaignCols = "id, name, description, campaign_type, channel, status, priority, target_audience, segment_criteria, scheduled_at, started_at, completed_at, total_recipients, sent_count, delivered_count, opened_count, clicked_count, bounced_count, unsubscribed_count, created_by, approved_by, metadata, created_at, updated_at, deleted_at"

func (r *Repository) CreateCampaign(ctx context.Context, c *domain.NotificationCampaign) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO notification_campaigns
		(id, name, description, campaign_type, channel, status, priority, target_audience, segment_criteria, scheduled_at, started_at, completed_at, total_recipients, sent_count, delivered_count, opened_count, clicked_count, bounced_count, unsubscribed_count, created_by, approved_by, metadata, created_at, updated_at, deleted_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.Name, c.Description, c.CampaignType, c.Channel, c.Status, c.Priority, c.TargetAudience, c.SegmentCriteria, c.ScheduledAt, c.StartedAt, c.CompletedAt, c.TotalRecipients, c.SentCount, c.DeliveredCount, c.OpenedCount, c.ClickedCount, c.BouncedCount, c.UnsubscribedCount, c.CreatedBy, c.ApprovedBy, c.Metadata, c.CreatedAt, c.UpdatedAt, c.DeletedAt)
	return err
}

func (r *Repository) GetCampaign(ctx context.Context, id string) (*domain.NotificationCampaign, error) {
	c := &domain.NotificationCampaign{}
	err := r.db.GetContext(ctx, c, "SELECT "+campaignCols+" FROM notification_campaigns WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return c, nil
}

func (r *Repository) ListCampaigns(ctx context.Context, status, campaignType string, limit, offset int) ([]domain.NotificationCampaign, error) {
	var items []domain.NotificationCampaign
	q := "SELECT " + campaignCols + " FROM notification_campaigns WHERE deleted_at IS NULL"
	args := []interface{}{}
	if status != "" {
		q += " AND status = ?"
		args = append(args, status)
	}
	if campaignType != "" {
		q += " AND campaign_type = ?"
		args = append(args, campaignType)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func (r *Repository) UpdateCampaign(ctx context.Context, c *domain.NotificationCampaign) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notification_campaigns SET
		name=?, description=?, campaign_type=?, channel=?, priority=?, target_audience=?, segment_criteria=?, scheduled_at=?, metadata=?, updated_at=?
		WHERE id=?`,
		c.Name, c.Description, c.CampaignType, c.Channel, c.Priority, c.TargetAudience, c.SegmentCriteria, c.ScheduledAt, c.Metadata, c.UpdatedAt, c.ID)
	return err
}

func (r *Repository) DeleteCampaign(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, "UPDATE notification_campaigns SET deleted_at = ? WHERE id = ?", now, id)
	return err
}

func (r *Repository) UpdateCampaignStatus(ctx context.Context, id string, status domain.CampaignStatus) error {
	_, err := r.db.ExecContext(ctx, "UPDATE notification_campaigns SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *Repository) UpdateCampaignStats(ctx context.Context, id string, sent, delivered, opened, clicked, bounced, unsubscribed int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notification_campaigns SET
		sent_count = sent_count + ?, delivered_count = delivered_count + ?, opened_count = opened_count + ?,
		clicked_count = clicked_count + ?, bounced_count = bounced_count + ?, unsubscribed_count = unsubscribed_count + ?
		WHERE id = ?`, sent, delivered, opened, clicked, bounced, unsubscribed, id)
	return err
}

var templateCols = "id, campaign_id, name, channel, subject, body_html, body_text, body_push, variables_schema, locale, version, is_active, created_at, updated_at"

func (r *Repository) CreateTemplate(ctx context.Context, t *domain.NotificationTemplate) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO notification_templates
		(id, campaign_id, name, channel, subject, body_html, body_text, body_push, variables_schema, locale, version, is_active, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.CampaignID, t.Name, t.Channel, t.Subject, t.BodyHTML, t.BodyText, t.BodyPush, t.VariablesSchema, t.Locale, t.Version, t.IsActive, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Repository) GetTemplate(ctx context.Context, id string) (*domain.NotificationTemplate, error) {
	t := &domain.NotificationTemplate{}
	err := r.db.GetContext(ctx, t, "SELECT "+templateCols+" FROM notification_templates WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return t, nil
}

func (r *Repository) ListTemplates(ctx context.Context, campaignID, channel string) ([]domain.NotificationTemplate, error) {
	var items []domain.NotificationTemplate
	q := "SELECT " + templateCols + " FROM notification_templates WHERE 1=1"
	args := []interface{}{}
	if campaignID != "" {
		q += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if channel != "" {
		q += " AND channel = ?"
		args = append(args, channel)
	}
	q += " ORDER BY version DESC"
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func (r *Repository) UpdateTemplate(ctx context.Context, t *domain.NotificationTemplate) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notification_templates SET
		name=?, channel=?, subject=?, body_html=?, body_text=?, body_push=?, variables_schema=?, locale=?, version=?, is_active=?, updated_at=?
		WHERE id=?`,
		t.Name, t.Channel, t.Subject, t.BodyHTML, t.BodyText, t.BodyPush, t.VariablesSchema, t.Locale, t.Version, t.IsActive, t.UpdatedAt, t.ID)
	return err
}

var recipientCols = "id, campaign_id, user_id, channel, status, sent_at, delivered_at, opened_at, clicked_at, error_message, metadata, created_at, updated_at"

func (r *Repository) AddRecipient(ctx context.Context, rec *domain.CampaignRecipient) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO campaign_recipients
		(id, campaign_id, user_id, channel, status, sent_at, delivered_at, opened_at, clicked_at, error_message, metadata, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		rec.ID, rec.CampaignID, rec.UserID, rec.Channel, rec.Status, rec.SentAt, rec.DeliveredAt, rec.OpenedAt, rec.ClickedAt, rec.ErrorMessage, rec.Metadata, rec.CreatedAt, rec.UpdatedAt)
	return err
}

func (r *Repository) ListRecipients(ctx context.Context, campaignID, status string, limit, offset int) ([]domain.CampaignRecipient, error) {
	var items []domain.CampaignRecipient
	q := "SELECT " + recipientCols + " FROM campaign_recipients WHERE campaign_id = ?"
	args := []interface{}{campaignID}
	if status != "" {
		q += " AND status = ?"
		args = append(args, status)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func (r *Repository) UpdateRecipientStatus(ctx context.Context, id string, status domain.RecipientStatus, msg *string) error {
	now := time.Now().UTC()
	var err error
	switch status {
	case domain.RecipientSent:
		_, err = r.db.ExecContext(ctx, "UPDATE campaign_recipients SET status = ?, sent_at = ? WHERE id = ?", status, now, id)
	case domain.RecipientDelivered:
		_, err = r.db.ExecContext(ctx, "UPDATE campaign_recipients SET status = ?, delivered_at = ? WHERE id = ?", status, now, id)
	case domain.RecipientOpened:
		_, err = r.db.ExecContext(ctx, "UPDATE campaign_recipients SET status = ?, opened_at = ? WHERE id = ?", status, now, id)
	case domain.RecipientClicked:
		_, err = r.db.ExecContext(ctx, "UPDATE campaign_recipients SET status = ?, clicked_at = ? WHERE id = ?", status, now, id)
	default:
		_, err = r.db.ExecContext(ctx, "UPDATE campaign_recipients SET status = ?, error_message = ? WHERE id = ?", status, msg, id)
	}
	return err
}

var prefCols = "id, user_id, email_enabled, sms_enabled, push_enabled, in_app_enabled, whatsapp_enabled, promotional_email, transactional_email, order_updates_sms, price_alerts_push, quiet_hours_start, quiet_hours_end, timezone, created_at, updated_at"

func (r *Repository) GetPreference(ctx context.Context, userID string) (*domain.NotificationPreference, error) {
	p := &domain.NotificationPreference{}
	err := r.db.GetContext(ctx, p, "SELECT "+prefCols+" FROM notification_preferences WHERE user_id = ?", userID)
	if err != nil {
		return nil, mapError(err)
	}
	return p, nil
}

func (r *Repository) UpsertPreference(ctx context.Context, p *domain.NotificationPreference) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO notification_preferences
		(id, user_id, email_enabled, sms_enabled, push_enabled, in_app_enabled, whatsapp_enabled, promotional_email, transactional_email, order_updates_sms, price_alerts_push, quiet_hours_start, quiet_hours_end, timezone, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
		email_enabled=VALUES(email_enabled), sms_enabled=VALUES(sms_enabled), push_enabled=VALUES(push_enabled),
		in_app_enabled=VALUES(in_app_enabled), whatsapp_enabled=VALUES(whatsapp_enabled),
		promotional_email=VALUES(promotional_email), transactional_email=VALUES(transactional_email),
		order_updates_sms=VALUES(order_updates_sms), price_alerts_push=VALUES(price_alerts_push),
		quiet_hours_start=VALUES(quiet_hours_start), quiet_hours_end=VALUES(quiet_hours_end),
		timezone=VALUES(timezone), updated_at=VALUES(updated_at)`,
		p.ID, p.UserID, p.EmailEnabled, p.SMSEnabled, p.PushEnabled, p.InAppEnabled, p.WhatsAppEnabled,
		p.PromotionalEmail, p.TransactionalEmail, p.OrderUpdatesSMS, p.PriceAlertsPush,
		p.QuietHoursStart, p.QuietHoursEnd, p.Timezone, p.CreatedAt, p.UpdatedAt)
	return err
}

var deliveryCols = "id, campaign_id, template_id, user_id, channel, status, provider, provider_message_id, error_code, error_message, retry_count, metadata, created_at"

func (r *Repository) CreateDeliveryLog(ctx context.Context, l *domain.NotificationDeliveryLog) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO notification_delivery_logs
		(id, campaign_id, template_id, user_id, channel, status, provider, provider_message_id, error_code, error_message, retry_count, metadata, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		l.ID, l.CampaignID, l.TemplateID, l.UserID, l.Channel, l.Status, l.Provider, l.ProviderMessageID, l.ErrorCode, l.ErrorMessage, l.RetryCount, l.Metadata, l.CreatedAt)
	return err
}

func (r *Repository) GetDeliveryLog(ctx context.Context, id string) (*domain.NotificationDeliveryLog, error) {
	l := &domain.NotificationDeliveryLog{}
	err := r.db.GetContext(ctx, l, "SELECT "+deliveryCols+" FROM notification_delivery_logs WHERE id = ?", id)
	if err != nil {
		return nil, mapError(err)
	}
	return l, nil
}

func (r *Repository) ListDeliveryLogs(ctx context.Context, campaignID, userID string, limit, offset int) ([]domain.NotificationDeliveryLog, error) {
	var items []domain.NotificationDeliveryLog
	q := "SELECT " + deliveryCols + " FROM notification_delivery_logs WHERE 1=1"
	args := []interface{}{}
	if campaignID != "" {
		q += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if userID != "" {
		q += " AND user_id = ?"
		args = append(args, userID)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &items, q, args...)
	return items, err
}

func mapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

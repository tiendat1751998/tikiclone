package mysql
import ("context"; "database/sql"; "github.com/jmoiron/sqlx"; "github.com/tikiclone/tiki/platforms/notification/internal/domain")

type NotificationRepository struct{ db *sqlx.DB }
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository { return &NotificationRepository{db: db} }
func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `INSERT INTO notifications (id, user_id, type, title, body, data, channel, status, priority, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, n.ID, n.UserID, n.Type, n.Title, n.Body, n.Data, n.Channel, n.Status, n.Priority, n.CreatedAt); return err
}
func (r *NotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	var n domain.Notification; err := r.db.GetContext(ctx, &n, "SELECT id, user_id, type, title, body, data, status, read_at, created_at, updated_at FROM notifications WHERE id = ?", id)
	if err == sql.ErrNoRows { return nil, nil }; return &n, err
}
func (r *NotificationRepository) FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*domain.Notification, int64, error) {
	var total int64; r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM notifications WHERE user_id = ?", userID)
	var notifs []*domain.Notification; err := r.db.SelectContext(ctx, &notifs, "SELECT id, user_id, type, title, body, data, status, read_at, created_at, updated_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", userID, limit, offset)
	return notifs, total, err
}
func (r *NotificationRepository) MarkRead(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE notifications SET status = 'read', read_at = NOW() WHERE id = ?", id); return err
}
func (r *NotificationRepository) MarkDelivered(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE notifications SET status = 'delivered' WHERE id = ?", id); return err
}
func (r *NotificationRepository) MarkFailed(ctx context.Context, id, reason string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE notifications SET status = 'failed' WHERE id = ?", id); return err
}

type TemplateRepository struct{ db *sqlx.DB }
func NewTemplateRepository(db *sqlx.DB) *TemplateRepository { return &TemplateRepository{db: db} }
func (r *TemplateRepository) FindByName(ctx context.Context, name string) (*domain.NotificationTemplate, error) {
	var t domain.NotificationTemplate; err := r.db.GetContext(ctx, &t, "SELECT id, name, channel, subject, body_template, variables, is_active, created_at, updated_at FROM notification_templates WHERE name = ? AND is_active = true", name)
	if err == sql.ErrNoRows { return nil, nil }; return &t, err
}
func (r *TemplateRepository) Create(ctx context.Context, t *domain.NotificationTemplate) error {
	query := `INSERT INTO notification_templates (id, name, type, subject, body, variables, version, is_active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Type, t.Subject, t.Body, t.Variables, t.Version, t.IsActive, t.CreatedAt); return err
}

type PreferenceRepository struct{ db *sqlx.DB }
func NewPreferenceRepository(db *sqlx.DB) *PreferenceRepository { return &PreferenceRepository{db: db} }
func (r *PreferenceRepository) GetUserPreferences(ctx context.Context, userID string) ([]*domain.UserPreference, error) {
	var prefs []*domain.UserPreference; err := r.db.SelectContext(ctx, &prefs, "SELECT id, user_id, channel, type, enabled, created_at, updated_at FROM user_preferences WHERE user_id = ?", userID)
	return prefs, err
}
func (r *PreferenceRepository) UpdatePreference(ctx context.Context, p *domain.UserPreference) error {
	query := `INSERT INTO user_preferences (user_id, channel, enabled, quiet_hours) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE enabled = ?, quiet_hours = ?`
	_, err := r.db.ExecContext(ctx, query, p.UserID, p.Channel, p.Enabled, p.QuietHours, p.Enabled, p.QuietHours); return err
}

type DeliveryLogRepository struct{ db *sqlx.DB }
func NewDeliveryLogRepository(db *sqlx.DB) *DeliveryLogRepository { return &DeliveryLogRepository{db: db} }
func (r *DeliveryLogRepository) Create(ctx context.Context, log *domain.DeliveryLog) error {
	query := `INSERT INTO delivery_logs (id, notification_id, channel, provider, status, error_message, attempt_count, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.ID, log.NotificationID, log.Channel, log.Provider, log.Status, log.ErrorMsg, log.AttemptCount, log.CreatedAt); return err
}
func (r *DeliveryLogRepository) FindByNotificationID(ctx context.Context, notificationID string) ([]*domain.DeliveryLog, error) {
	var logs []*domain.DeliveryLog; err := r.db.SelectContext(ctx, &logs, "SELECT id, notification_id, channel, status, error, response, created_at FROM delivery_logs WHERE notification_id = ? ORDER BY created_at DESC LIMIT 100", notificationID)
	return logs, err
}

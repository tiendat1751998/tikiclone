package mysql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.session.create")
	defer span.End()

	query := `INSERT INTO sessions (id, user_id, refresh_token_id, device_id, device_name, device_type, platform, ip, user_agent, location, status, last_active_at, expires_at, created_at)
		VALUES (:id, :user_id, :refresh_token_id, :device_id, :device_name, :device_type, :platform, :ip, :user_agent, :location, :status, :last_active_at, :expires_at, :created_at)`

	_, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("session create: %w", err)
	}
	return nil
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.session.find_by_id")
	defer span.End()

	var session domain.Session
	query := `SELECT id, user_id, refresh_token_id, COALESCE(device_id,'') as device_id, COALESCE(device_name,'') as device_name,
		COALESCE(device_type,'') as device_type, COALESCE(platform,'') as platform, COALESCE(ip,'') as ip,
		COALESCE(user_agent,'') as user_agent, COALESCE(location,'') as location, status, last_active_at, expires_at, created_at
		FROM sessions WHERE id = ? LIMIT 1`

	err := r.db.GetContext(ctx, &session, query, id)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("session find by id: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) FindByRefreshTokenID(ctx context.Context, tokenID string) (*domain.Session, error) {
	query := `SELECT id, user_id, refresh_token_id, COALESCE(device_id,'') as device_id, COALESCE(ip,'') as ip,
		COALESCE(user_agent,'') as user_agent, status, last_active_at, expires_at, created_at
		FROM sessions WHERE refresh_token_id = ? LIMIT 1`

	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, tokenID)
	if err != nil {
		return nil, fmt.Errorf("session find by refresh token id: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.session.find_active_by_user")
	defer span.End()

	var sessions []*domain.Session
	query := `SELECT id, user_id, refresh_token_id, COALESCE(device_id,'') as device_id, COALESCE(device_name,'') as device_name,
		COALESCE(device_type,'') as device_type, COALESCE(platform,'') as platform, COALESCE(ip,'') as ip,
		COALESCE(user_agent,'') as user_agent, COALESCE(location,'') as location, status, last_active_at, expires_at, created_at
		FROM sessions WHERE user_id = ? AND status = 'active' AND expires_at > NOW()
		ORDER BY last_active_at DESC`

	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.Int("count", len(sessions)))
	return sessions, nil
}

func (r *SessionRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM sessions WHERE user_id = ? AND status = 'active' AND expires_at > NOW()`
	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *SessionRepository) FindOldestActiveByUserID(ctx context.Context, userID string) (*domain.Session, error) {
	query := `SELECT id, user_id, refresh_token_id, COALESCE(device_id,'') as device_id, COALESCE(ip,'') as ip,
		COALESCE(user_agent,'') as user_agent, status, last_active_at, expires_at, created_at
		FROM sessions WHERE user_id = ? AND status = 'active' AND expires_at > NOW()
		ORDER BY last_active_at ASC LIMIT 1`

	var session domain.Session
	err := r.db.GetContext(ctx, &session, query, userID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `UPDATE sessions SET status = :status, last_active_at = :last_active_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, session)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR status IN ('expired', 'revoked')`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *SessionRepository) CleanupUserSessions(ctx context.Context, userID string, keepCount int) error {
	query := `DELETE FROM sessions
		WHERE user_id = ? AND status = 'active'
		AND id NOT IN (
			SELECT id FROM (
				SELECT id FROM sessions
				WHERE user_id = ? AND status = 'active'
				ORDER BY last_active_at DESC
				LIMIT ?
			) AS keep_sessions
		)`
	_, err := r.db.ExecContext(ctx, query, userID, userID, keepCount)
	return err
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *SessionRepository) UpdateSessionDevice(ctx context.Context, sessionID, deviceID, deviceName string) error {
	query := `UPDATE sessions SET device_id = ?, device_name = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, deviceID, deviceName, sessionID)
	return err
}

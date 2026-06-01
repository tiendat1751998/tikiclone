package mysql

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
)

type FailedLoginAttemptRepository struct {
	db *sqlx.DB
}

func NewFailedLoginAttemptRepository(db *sqlx.DB) *FailedLoginAttemptRepository {
	return &FailedLoginAttemptRepository{db: db}
}

func (r *FailedLoginAttemptRepository) Create(ctx context.Context, attempt *domain.FailedLoginAttempt) error {
	query := `INSERT INTO failed_login_attempts (id, email, ip_address, attempted_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, attempt.ID, attempt.Email, attempt.IPAddress, attempt.AttemptedAt)
	return err
}

func (r *FailedLoginAttemptRepository) CountByEmailSince(ctx context.Context, email string, since time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM failed_login_attempts WHERE email = ? AND attempted_at > ?", email, since)
	return count, err
}

func (r *FailedLoginAttemptRepository) CountByIPSince(ctx context.Context, ip string, since time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM failed_login_attempts WHERE ip_address = ? AND attempted_at > ?", ip, since)
	return count, err
}

func (r *FailedLoginAttemptRepository) DeleteByEmail(ctx context.Context, email string) error {
	query := `DELETE FROM failed_login_attempts WHERE email = ?`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

func (r *FailedLoginAttemptRepository) DeleteOlderThan(ctx context.Context, before time.Time) error {
	query := `DELETE FROM failed_login_attempts WHERE attempted_at < ?`
	_, err := r.db.ExecContext(ctx, query, before)
	return err
}
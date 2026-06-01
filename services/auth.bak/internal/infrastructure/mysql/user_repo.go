package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/packages/go-shared/pkg/observability"
	"github.com/tikiclone/tiki/services/auth/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.user.create")
	defer span.End()

	query := `INSERT INTO users (id, email, phone, username, password_hash, display_name, status, email_verified, phone_verified, created_at, updated_at)
		VALUES (:id, :email, :phone, :username, :password_hash, :display_name, :status, :email_verified, :phone_verified, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		observability.LogWithTrace(ctx).Error("failed to create user", zap.Error(err))
		return fmt.Errorf("user create: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.user.find_by_id")
	defer span.End()

	var user domain.User
	query := `SELECT id, email, phone, username, password_hash, display_name, status, email_verified, phone_verified,
		mfa_enabled, twofa_secret, last_login_at, last_login_ip, failed_attempts, locked_until, metadata, created_at, updated_at
		FROM users WHERE id = ? LIMIT 1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("user find by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.user.find_by_email")
	defer span.End()

	var user domain.User
	query := `SELECT id, email, phone, username, password_hash, display_name, status, email_verified, phone_verified,
		mfa_enabled, twofa_secret, last_login_at, last_login_ip, failed_attempts, locked_until, metadata, created_at, updated_at
		FROM users WHERE email = ? LIMIT 1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("user find by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.user.find_by_username")
	defer span.End()

	var user domain.User
	query := `SELECT id, email, phone, username, password_hash, display_name, status,
		created_at, updated_at FROM users WHERE username = ? LIMIT 1`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("user find by username: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	ctx, span := otel.Tracer("shopee-auth").Start(ctx, "mysql.user.update")
	defer span.End()

	query := `UPDATE users SET
		email = :email, phone = :phone, display_name = :display_name,
		status = :status, email_verified = :email_verified,
		last_login_at = :last_login_at, last_login_ip = :last_login_ip,
		failed_attempts = :failed_attempts, locked_until = :locked_until,
		updated_at = :updated_at
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("user update: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("password update: %w", err)
	}
	return nil
}

func (r *UserRepository) AssignRole(ctx context.Context, userID, role string) error {
	query := `INSERT INTO user_roles (user_id, role) VALUES (?, ?) ON DUPLICATE KEY UPDATE role = ?`
	_, err := r.db.ExecContext(ctx, query, userID, role, role)
	return err
}

func (r *UserRepository) FindRolesByUserID(ctx context.Context, userID string) ([]domain.Role, error) {
	query := `SELECT role FROM user_roles WHERE user_id = ?`
	var roles []domain.Role
	err := r.db.SelectContext(ctx, &roles, query, userID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		return []domain.Role{domain.RoleBuyer}, nil
	}
	return roles, nil
}

func (r *UserRepository) RemoveRole(ctx context.Context, userID, role string) error {
	query := `DELETE FROM user_roles WHERE user_id = ? AND role = ?`
	_, err := r.db.ExecContext(ctx, query, userID, role)
	return err
}

func (r *UserRepository) UpdateEmailVerified(ctx context.Context, userID string) error {
	query := `UPDATE users SET email_verified = TRUE, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

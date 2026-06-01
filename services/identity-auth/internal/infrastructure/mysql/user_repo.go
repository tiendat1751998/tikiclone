package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at)
		VALUES (:id, :email, :phone, :password_hash, :full_name, :role, :is_verified, :is_active, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at FROM users WHERE email = ? LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at FROM users WHERE phone = ? LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE email = ?", email)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE phone = ?", phone)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, phone, password_hash, full_name, role, is_verified, is_active, created_at, updated_at FROM users WHERE id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/tikiclone/tiki/services/identity-auth/internal/domain"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.GetContext(ctx, &role, "SELECT role_id, name, description, is_system, created_at FROM roles WHERE name = ? LIMIT 1", name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

type UserRoleRepository struct {
	db *sqlx.DB
}

func NewUserRoleRepository(db *sqlx.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) AssignDefaultRole(ctx context.Context, userID string) error {
	query := `INSERT IGNORE INTO user_roles (user_id, role_id) SELECT ?, role_id FROM roles WHERE name = 'BUYER'`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
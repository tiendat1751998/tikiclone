package rbac

import (
	"context"
	"fmt"

	"github.com/tikiclone/tiki/services/auth/internal/domain"
)

type RoleRepository interface {
	AssignRole(ctx context.Context, userID, role string) error
	RemoveRole(ctx context.Context, userID, role string) error
	FindRolesByUserID(ctx context.Context, userID string) ([]domain.Role, error)
}

type Service struct {
	repo RoleRepository
}

func NewService(repo RoleRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserRoles(ctx context.Context, userID string) []domain.Role {
	roles, err := s.repo.FindRolesByUserID(ctx, userID)
	if err != nil || len(roles) == 0 {
		return []domain.Role{domain.RoleBuyer}
	}
	return roles
}

func (s *Service) AssignRole(ctx context.Context, userID string, role domain.Role) error {
	if _, ok := domain.RoleHierarchy[role]; !ok {
		return fmt.Errorf("invalid role: %s", role)
	}
	return s.repo.AssignRole(ctx, userID, string(role))
}

func (s *Service) AssignDefaultRole(ctx context.Context, userID string) error {
	return s.AssignRole(ctx, userID, domain.RoleBuyer)
}

func (s *Service) RemoveRole(ctx context.Context, userID string, role domain.Role) error {
	return s.repo.RemoveRole(ctx, userID, string(role))
}

func (s *Service) HasRole(userRoles []domain.Role, target domain.Role) bool {
	for _, r := range userRoles {
		if r == target {
			return true
		}
	}
	return false
}

func (s *Service) HasPermission(userRoles []domain.Role, permission domain.Permission) bool {
	for _, r := range userRoles {
		def, ok := domain.RoleHierarchy[r]
		if !ok {
			continue
		}
		for _, p := range def.Permissions {
			if p == permission {
				return true
			}
		}
	}
	return false
}

func (s *Service) IsAdmin(userRoles []domain.Role) bool {
	return s.HasRole(userRoles, domain.RoleAdmin)
}

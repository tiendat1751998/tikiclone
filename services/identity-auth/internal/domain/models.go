package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `db:"id" json:"userId"`
	Email       string     `db:"email" json:"email"`
	Phone       *string    `db:"phone" json:"phone,omitempty"`
	PasswordHash string     `db:"password_hash" json:"-"`
	FullName    string     `db:"full_name" json:"fullName"`
	Role        string     `db:"role" json:"role"`
	IsVerified  bool       `db:"is_verified" json:"verified"`
	IsActive    bool       `db:"is_active" json:"-"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
}

func NewUser(email, phone, fullName, passwordHash string) *User {
	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}
	return &User{
		ID:           uuid.New(),
		Email:        email,
		Phone:        phonePtr,
		FullName:     fullName,
		PasswordHash: passwordHash,
		Role:         "BUYER",
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id" json:"-"`
	Token     string    `db:"token" json:"-"`
	UserID    uuid.UUID `db:"user_id" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"-"`
	Revoked   bool      `db:"revoked" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"-"`
}

type FailedLoginAttempt struct {
	ID         uuid.UUID `db:"id"`
	Email      string    `db:"email"`
	IPAddress  string    `db:"ip_address"`
	AttemptedAt time.Time `db:"attempted_at"`
}

type OutboxEvent struct {
	EventID    uuid.UUID `db:"event_id"`
	AggregateType string `db:"aggregate_type"`
	AggregateID string `db:"aggregate_id"`
	EventType   string `db:"event_type"`
	Payload     string `db:"payload"`
	Processed   bool    `db:"processed"`
	ProcessedAt *time.Time `db:"processed_at"`
	RetryCount  int     `db:"retry_count"`
	LastError   *string `db:"last_error"`
	CreatedAt   time.Time `db:"created_at"`
}

type Role struct {
	RoleID    uuid.UUID `db:"role_id"`
	Name      string    `db:"name"`
	Desc      *string   `db:"description"`
	IsSystem  bool      `db:"is_system"`
	CreatedAt time.Time `db:"created_at"`
}

type Permission struct {
	PermissionID uuid.UUID `db:"permission_id"`
	Resource     string    `db:"resource"`
	Action       string    `db:"action"`
	Desc         *string   `db:"description"`
	CreatedAt    time.Time `db:"created_at"`
}

// DTOs matching Java AuthRequest/AuthResponse
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"fullName" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AuthResponse struct {
	UserID       uuid.UUID `json:"userId"`
	Email        string    `json:"email"`
	Phone        *string   `json:"phone,omitempty"`
	FullName     string    `json:"fullName"`
	Role         string    `json:"role"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int64     `json:"expiresIn"`
}

type UserResponse struct {
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	FullName  string    `json:"fullName"`
	Role      string    `json:"role"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type ValidateTokenResponse struct {
	Valid bool `json:"valid"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
package domain

import "errors"

var (
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrPhoneAlreadyExists = errors.New("phone already registered")
	ErrInvalidCredentials = errors.New("Invalid email or password")
	ErrInvalidToken       = errors.New("Invalid or expired refresh token")
	ErrTokenNotFound      = errors.New("Refresh token not found")
	ErrTokenRevoked       = errors.New("Refresh token has been revoked")
	ErrTokenExpired       = errors.New("Refresh token has expired")
	ErrInvalidRefreshToken = errors.New("Invalid refresh token")
	ErrAccountLocked      = errors.New("Account temporarily locked due to too many failed attempts")
	ErrAccountDisabled    = errors.New("Account is deactivated")
	ErrUserNotFound       = errors.New("User not found")
	ErrPasswordTooWeak    = errors.New("Password must be at least 8 characters")
	ErrRateLimited        = errors.New("Too many login attempts. Please try again later.")
	ErrIPBlocked          = errors.New("Too many requests from this IP. Please try again later.")
	ErrTokenReuse          = errors.New("Refresh token reuse detected")
)

func IsValidationError(err error) bool {
	return errors.Is(err, ErrPasswordTooWeak)
}
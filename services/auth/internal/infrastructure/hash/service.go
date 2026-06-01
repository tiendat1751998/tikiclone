package hash

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/tikiclone/tiki/services/auth/internal/config"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	cfg config.PasswordConfig
}

func NewService(cfg config.PasswordConfig) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Hash(ctx context.Context, password string) (string, error) {
	switch s.cfg.Algorithm {
	case "argon2id":
		return s.hashArgon2id(password)
	case "bcrypt":
		return s.hashBcrypt(password)
	default:
		return s.hashBcrypt(password)
	}
}

func (s *Service) Verify(ctx context.Context, password, hash string) bool {
	switch {
	case len(hash) >= 60 && (hash[:4] == "$2a$" || hash[:4] == "$2b$"):
		// Try direct compare first (new format: bcrypt(sha256(password)))
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
			return true
		}
		// Backward compat: if password looks like SHA-256 hex, it won't match old bcrypt(password).
		// There's no way to recover the original password from SHA-256, so this is intentionally
		// a breaking change for users registered before client-side hashing was introduced.
		return false
	default:
		return s.verifyArgon2id(password, hash)
	}
}

func (s *Service) hashBcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.Cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(bytes), nil
}

func (s *Service) hashArgon2id(password string) (string, error) {
	salt := generateSalt(s.cfg.SaltLen)
	hash := argon2.IDKey([]byte(password), salt, s.cfg.Time, s.cfg.Memory, s.cfg.Threads, s.cfg.KeyLen)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%x$%x",
		argon2.Version, s.cfg.Memory, s.cfg.Time, s.cfg.Threads, salt, hash)

	return encoded, nil
}

func (s *Service) verifyArgon2id(password, encoded string) bool {
	var version int
	var memory, time uint32
	var threads uint8
	var salt, hash []byte

	_, err := fmt.Sscanf(encoded, "$argon2id$v=%d$m=%d,t=%d,p=%d$%x$%x",
		&version, &memory, &time, &threads, &salt, &hash)
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	return constantTimeCompare(computedHash, hash)
}

func generateSalt(length int) []byte {
	salt := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		// fallback to a less secure but functional approach
		for i := range salt {
			salt[i] = byte(i * 37)
		}
	}
	return salt
}

func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}
	return result == 0
}

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("auth: not found")
	ErrEmailTaken   = errors.New("auth: email already registered")
	ErrInvalidToken = errors.New("auth: invalid or expired token")
)

type User struct {
	ID              uuid.UUID
	Email           string
	Name            string
	PasswordHash    *string
	EmailVerifiedAt *time.Time
}

type NewUser struct {
	Email        string
	Name         string
	PasswordHash string
}

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUserWithVerificationToken(ctx context.Context, user NewUser, tokenHash []byte, expiresAt time.Time) error
	ReplaceVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	ConsumeVerificationToken(ctx context.Context, tokenHash []byte) (uuid.UUID, time.Time, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
}

type Mailer interface {
	SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error
	SendAccountExistsEmail(ctx context.Context, to, name, signInURL string) error
}

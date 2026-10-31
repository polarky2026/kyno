package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("auth: not found")
	ErrEmailTaken         = errors.New("auth: email already registered")
	ErrInvalidToken       = errors.New("auth: invalid or expired token")
	ErrInvalidCredentials = errors.New("auth: invalid email or password")
	ErrEmailNotVerified   = errors.New("auth: email not verified")
	ErrUnauthenticated    = errors.New("auth: unauthenticated")
)

type User struct {
	ID              uuid.UUID
	Email           string
	Name            string
	PasswordHash    *string
	EmailVerifiedAt *time.Time
	Role            string
}

type NewUser struct {
	Email        string
	Name         string
	PasswordHash string
}

// Session is what the client receives after signing in.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// StoredSession is a session as the store returns it, with its user.
type StoredSession struct {
	User      User
	ExpiresAt time.Time
}

// AuthResult is the outcome of authenticating a request.
type AuthResult struct {
	User      User
	ExpiresAt time.Time
	Renewed   bool
}

type Store interface {
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUserWithVerificationToken(ctx context.Context, user NewUser, tokenHash []byte, expiresAt time.Time) error
	ReplaceVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	ConsumeVerificationToken(ctx context.Context, tokenHash []byte) (uuid.UUID, time.Time, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error

	CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	// GetSession returns ErrNotFound when no session has this token hash.
	GetSession(ctx context.Context, tokenHash []byte) (StoredSession, error)
	ExtendSession(ctx context.Context, tokenHash []byte, expiresAt time.Time) error
	DeleteSession(ctx context.Context, tokenHash []byte) error
}

type Mailer interface {
	SendVerificationEmail(ctx context.Context, to, name, verifyURL string) error
	SendAccountExistsEmail(ctx context.Context, to, name, signInURL string) error
}

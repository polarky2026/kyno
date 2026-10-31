package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/polarky2026/kyno/apps/api/internal/auth/password"
	"github.com/polarky2026/kyno/apps/api/internal/auth/token"
)

const (
	verificationTokenTTL = 24 * time.Hour
	emailSendTimeout     = 15 * time.Second
	sessionTTL           = 30 * 24 * time.Hour
	sessionRenewWithin   = 15 * 24 * time.Hour
)

// dummyPasswordHash lets SignIn spend the same time on unknown emails as on
// real ones, so response times don't reveal which emails have accounts.
var dummyPasswordHash = sync.OnceValue(func() string {
	hash, err := password.Hash("kyno-timing-equalizer")
	if err != nil {
		panic(fmt.Sprintf("auth: hash dummy password: %v", err))
	}
	return hash
})

type Service struct {
	store    Store
	mailer   Mailer
	appURL   string
	now      func() time.Time
	dispatch func(func())
}

func NewService(store Store, mailer Mailer, appURL string) *Service {
	return &Service{
		store:    store,
		mailer:   mailer,
		appURL:   strings.TrimRight(appURL, "/"),
		now:      time.Now,
		dispatch: func(task func()) { go task() },
	}
}

type SignUpInput struct {
	Name     string
	Email    string
	Password string
}

func (s *Service) SignUp(ctx context.Context, in SignUpInput) error {
	name := strings.TrimSpace(in.Name)
	email := normalizeEmail(in.Email)

	if err := validateName(name); err != nil {
		return err
	}
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePassword(in.Password); err != nil {
		return err
	}

	passwordHash, err := password.Hash(in.Password)
	if err != nil {
		return err
	}

	existing, err := s.store.GetUserByEmail(ctx, email)
	switch {
	case err == nil:
		s.sendAccountExists(existing)
		return nil
	case !errors.Is(err, ErrNotFound):
		return fmt.Errorf("look up user: %w", err)
	}

	plainToken, tokenHash, err := token.New()
	if err != nil {
		return err
	}

	newUser := NewUser{Email: email, Name: name, PasswordHash: passwordHash}
	expiresAt := s.now().Add(verificationTokenTTL)

	err = s.store.CreateUserWithVerificationToken(ctx, newUser, tokenHash, expiresAt)
	if errors.Is(err, ErrEmailTaken) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	s.sendVerification(email, name, plainToken)
	return nil
}

func (s *Service) VerifyEmail(ctx context.Context, plainToken string) (uuid.UUID, error) {
	if plainToken == "" {
		return uuid.Nil, ErrInvalidToken
	}

	userID, expiresAt, err := s.store.ConsumeVerificationToken(ctx, token.Hash(plainToken))
	if errors.Is(err, ErrNotFound) {
		return uuid.Nil, ErrInvalidToken
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("consume verification token: %w", err)
	}

	if !s.now().Before(expiresAt) {
		return uuid.Nil, ErrInvalidToken
	}

	if err := s.store.MarkEmailVerified(ctx, userID); err != nil {
		return uuid.Nil, fmt.Errorf("mark email verified: %w", err)
	}

	return userID, nil
}

func (s *Service) ResendVerification(ctx context.Context, email string) error {
	user, err := s.store.GetUserByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("look up user: %w", err)
	}

	if user.EmailVerifiedAt != nil {
		return nil
	}

	return s.issueVerification(ctx, user)
}

type SignInInput struct {
	Email    string
	Password string
}

func (s *Service) SignIn(ctx context.Context, in SignInInput) (Session, error) {
	user, err := s.store.GetUserByEmail(ctx, normalizeEmail(in.Email))
	if errors.Is(err, ErrNotFound) || (err == nil && user.PasswordHash == nil) {
		_, _ = password.Verify(in.Password, dummyPasswordHash())
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, fmt.Errorf("look up user: %w", err)
	}

	ok, err := password.Verify(in.Password, *user.PasswordHash)
	if err != nil {
		return Session{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return Session{}, ErrInvalidCredentials
	}

	if user.EmailVerifiedAt == nil {
		if err := s.issueVerification(ctx, user); err != nil {
			return Session{}, err
		}
		return Session{}, ErrEmailNotVerified
	}

	return s.StartSession(ctx, user.ID)
}

func (s *Service) StartSession(ctx context.Context, userID uuid.UUID) (Session, error) {
	plainToken, tokenHash, err := token.New()
	if err != nil {
		return Session{}, err
	}

	expiresAt := s.now().Add(sessionTTL)
	if err := s.store.CreateSession(ctx, userID, tokenHash, expiresAt); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	return Session{Token: plainToken, ExpiresAt: expiresAt}, nil
}

func (s *Service) Authenticate(ctx context.Context, plainToken string) (AuthResult, error) {
	if plainToken == "" {
		return AuthResult{}, ErrUnauthenticated
	}

	tokenHash := token.Hash(plainToken)
	stored, err := s.store.GetSession(ctx, tokenHash)
	if errors.Is(err, ErrNotFound) {
		return AuthResult{}, ErrUnauthenticated
	}
	if err != nil {
		return AuthResult{}, fmt.Errorf("get session: %w", err)
	}

	now := s.now()
	if !now.Before(stored.ExpiresAt) {
		return AuthResult{}, ErrUnauthenticated
	}

	result := AuthResult{User: stored.User, ExpiresAt: stored.ExpiresAt}
	if stored.ExpiresAt.Sub(now) < sessionRenewWithin {
		result.ExpiresAt = now.Add(sessionTTL)
		if err := s.store.ExtendSession(ctx, tokenHash, result.ExpiresAt); err != nil {
			return AuthResult{}, fmt.Errorf("extend session: %w", err)
		}
		result.Renewed = true
	}

	return result, nil
}

func (s *Service) SignOut(ctx context.Context, plainToken string) error {
	if plainToken == "" {
		return nil
	}
	if err := s.store.DeleteSession(ctx, token.Hash(plainToken)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Service) issueVerification(ctx context.Context, user User) error {
	plainToken, tokenHash, err := token.New()
	if err != nil {
		return err
	}

	expiresAt := s.now().Add(verificationTokenTTL)
	if err := s.store.ReplaceVerificationToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("replace verification token: %w", err)
	}

	s.sendVerification(user.Email, user.Name, plainToken)
	return nil
}

func (s *Service) sendVerification(email, name, plainToken string) {
	verifyURL := s.appURL + "/verify-email?token=" + url.QueryEscape(plainToken)

	s.send("verification", func(ctx context.Context) error {
		return s.mailer.SendVerificationEmail(ctx, email, name, verifyURL)
	})
}

func (s *Service) sendAccountExists(user User) {
	signInURL := s.appURL + "/sign-in"

	s.send("account_exists", func(ctx context.Context) error {
		return s.mailer.SendAccountExistsEmail(ctx, user.Email, user.Name, signInURL)
	})
}

func (s *Service) send(kind string, deliver func(ctx context.Context) error) {
	s.dispatch(func() {
		ctx, cancel := context.WithTimeout(context.Background(), emailSendTimeout)
		defer cancel()

		if err := deliver(ctx); err != nil {
			slog.Error("failed to send email", "kind", kind, "error", err)
		}
	})
}

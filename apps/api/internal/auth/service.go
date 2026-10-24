package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/polarky2026/kyno/apps/api/internal/auth/password"
	"github.com/polarky2026/kyno/apps/api/internal/auth/token"
)

const (
	verificationTokenTTL = 24 * time.Hour
	emailSendTimeout     = 15 * time.Second
)

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

package auth

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/polarky2026/kyno/apps/api/internal/auth/password"
)

type storedToken struct {
	hash      []byte
	userID    uuid.UUID
	expiresAt time.Time
}

type fakeStore struct {
	users  map[string]User
	tokens []storedToken
}

func newFakeStore() *fakeStore {
	return &fakeStore{users: map[string]User{}}
}

func (f *fakeStore) GetUserByEmail(_ context.Context, email string) (User, error) {
	user, ok := f.users[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (f *fakeStore) CreateUserWithVerificationToken(_ context.Context, u NewUser, tokenHash []byte, expiresAt time.Time) error {
	if _, ok := f.users[u.Email]; ok {
		return ErrEmailTaken
	}
	hash := u.PasswordHash
	user := User{ID: uuid.New(), Email: u.Email, Name: u.Name, PasswordHash: &hash}
	f.users[u.Email] = user
	f.tokens = append(f.tokens, storedToken{hash: tokenHash, userID: user.ID, expiresAt: expiresAt})
	return nil
}

func (f *fakeStore) ReplaceVerificationToken(_ context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	kept := f.tokens[:0]
	for _, t := range f.tokens {
		if t.userID != userID {
			kept = append(kept, t)
		}
	}
	f.tokens = append(kept, storedToken{hash: tokenHash, userID: userID, expiresAt: expiresAt})
	return nil
}

func (f *fakeStore) ConsumeVerificationToken(_ context.Context, tokenHash []byte) (uuid.UUID, time.Time, error) {
	for i, t := range f.tokens {
		if bytes.Equal(t.hash, tokenHash) {
			f.tokens = append(f.tokens[:i], f.tokens[i+1:]...)
			return t.userID, t.expiresAt, nil
		}
	}
	return uuid.Nil, time.Time{}, ErrNotFound
}

func (f *fakeStore) MarkEmailVerified(_ context.Context, userID uuid.UUID) error {
	for email, user := range f.users {
		if user.ID == userID && user.EmailVerifiedAt == nil {
			now := time.Now()
			user.EmailVerifiedAt = &now
			f.users[email] = user
		}
	}
	return nil
}

type sentEmail struct {
	kind string
	to   string
	url  string
}

type fakeMailer struct {
	sent []sentEmail
}

func (m *fakeMailer) SendVerificationEmail(_ context.Context, to, _, verifyURL string) error {
	m.sent = append(m.sent, sentEmail{kind: "verification", to: to, url: verifyURL})
	return nil
}

func (m *fakeMailer) SendAccountExistsEmail(_ context.Context, to, _, signInURL string) error {
	m.sent = append(m.sent, sentEmail{kind: "account_exists", to: to, url: signInURL})
	return nil
}

func newTestService() (*Service, *fakeStore, *fakeMailer) {
	store := newFakeStore()
	mailer := &fakeMailer{}
	svc := NewService(store, mailer, "http://localhost:5173/")
	svc.dispatch = func(task func()) { task() }
	return svc, store, mailer
}

func tokenFrom(t *testing.T, verifyURL string) string {
	t.Helper()
	parsed, err := url.Parse(verifyURL)
	if err != nil {
		t.Fatalf("parse verify URL: %v", err)
	}
	return parsed.Query().Get("token")
}

var validInput = SignUpInput{Name: " George ", Email: " George@Example.com ", Password: "a-long-password"}

func TestSignUpCreatesUnverifiedUserAndSendsVerification(t *testing.T) {
	svc, store, mailer := newTestService()

	if err := svc.SignUp(context.Background(), validInput); err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	user, ok := store.users["george@example.com"]
	if !ok {
		t.Fatal("user was not stored with a normalized email")
	}
	if user.Name != "George" {
		t.Errorf("name = %q, want trimmed George", user.Name)
	}
	if user.EmailVerifiedAt != nil {
		t.Error("new user must not be verified")
	}
	if ok, _ := password.Verify("a-long-password", *user.PasswordHash); !ok {
		t.Error("stored hash does not match the password")
	}

	if len(mailer.sent) != 1 || mailer.sent[0].kind != "verification" {
		t.Fatalf("sent = %+v, want one verification email", mailer.sent)
	}
	if tokenFrom(t, mailer.sent[0].url) == "" {
		t.Fatal("verification URL has no token")
	}
}

func TestSignUpWithExistingEmailDoesNotRevealIt(t *testing.T) {
	svc, store, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	mailer.sent = nil

	err := svc.SignUp(context.Background(), SignUpInput{Name: "Impostor", Email: "george@example.com", Password: "another-password"})
	if err != nil {
		t.Fatalf("SignUp with an existing email returned %v; want nil", err)
	}

	if store.users["george@example.com"].Name != "George" {
		t.Error("existing account must not change")
	}
	if len(mailer.sent) != 1 || mailer.sent[0].kind != "account_exists" {
		t.Fatalf("sent = %+v, want one account_exists email", mailer.sent)
	}
}

func TestSignUpValidatesInput(t *testing.T) {
	cases := []struct {
		name  string
		input SignUpInput
		field string
	}{
		{"empty name", SignUpInput{Name: "  ", Email: "a@example.com", Password: "a-long-password"}, "name"},
		{"invalid email", SignUpInput{Name: "A", Email: "not-an-email", Password: "a-long-password"}, "email"},
		{"email with display name", SignUpInput{Name: "A", Email: "A <a@example.com>", Password: "a-long-password"}, "email"},
		{"short password", SignUpInput{Name: "A", Email: "a@example.com", Password: "short"}, "password"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, mailer := newTestService()

			err := svc.SignUp(context.Background(), tc.input)

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != tc.field {
				t.Fatalf("error = %v; want a ValidationError on %s", err, tc.field)
			}
			if len(mailer.sent) != 0 {
				t.Fatal("no email must be sent for invalid input")
			}
		})
	}
}

func TestVerifyEmailMarksUserAsVerifiedOnce(t *testing.T) {
	svc, store, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	plainToken := tokenFrom(t, mailer.sent[0].url)

	userID, err := svc.VerifyEmail(context.Background(), plainToken)
	if err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}

	user := store.users["george@example.com"]
	if user.ID != userID || user.EmailVerifiedAt == nil {
		t.Fatal("user was not marked as verified")
	}

	if _, err := svc.VerifyEmail(context.Background(), plainToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("second use error = %v; want ErrInvalidToken", err)
	}
}

func TestVerifyEmailRejectsExpiredTokens(t *testing.T) {
	svc, store, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	plainToken := tokenFrom(t, mailer.sent[0].url)

	svc.now = func() time.Time { return time.Now().Add(25 * time.Hour) }

	if _, err := svc.VerifyEmail(context.Background(), plainToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v; want ErrInvalidToken", err)
	}
	if store.users["george@example.com"].EmailVerifiedAt != nil {
		t.Fatal("an expired token must not verify the user")
	}
}

func TestVerifyEmailRejectsUnknownTokens(t *testing.T) {
	svc, _, _ := newTestService()

	for _, plainToken := range []string{"", "not-a-real-token"} {
		if _, err := svc.VerifyEmail(context.Background(), plainToken); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("VerifyEmail(%q) error = %v; want ErrInvalidToken", plainToken, err)
		}
	}
}

func TestResendVerificationReplacesTheToken(t *testing.T) {
	svc, _, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	oldToken := tokenFrom(t, mailer.sent[0].url)

	if err := svc.ResendVerification(context.Background(), "GEORGE@example.com"); err != nil {
		t.Fatalf("ResendVerification: %v", err)
	}
	if len(mailer.sent) != 2 {
		t.Fatalf("sent %d emails, want 2", len(mailer.sent))
	}

	if _, err := svc.VerifyEmail(context.Background(), oldToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("the old token must stop working after a resend")
	}
	if _, err := svc.VerifyEmail(context.Background(), tokenFrom(t, mailer.sent[1].url)); err != nil {
		t.Fatalf("the new token must work: %v", err)
	}
}

func TestResendVerificationIsSilentForUnknownOrVerifiedEmails(t *testing.T) {
	svc, _, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	_, _ = svc.VerifyEmail(context.Background(), tokenFrom(t, mailer.sent[0].url))
	mailer.sent = nil

	for _, email := range []string{"george@example.com", "nobody@example.com"} {
		if err := svc.ResendVerification(context.Background(), email); err != nil {
			t.Fatalf("ResendVerification(%q) = %v; want nil", email, err)
		}
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("sent = %+v, want no emails", mailer.sent)
	}
}

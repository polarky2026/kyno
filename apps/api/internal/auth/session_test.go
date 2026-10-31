package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/polarky2026/kyno/apps/api/internal/auth/token"
)

const day = 24 * time.Hour

var base = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

func newVerifiedTestService(t *testing.T) (*Service, *fakeStore, *fakeMailer) {
	t.Helper()
	svc, store, mailer := newTestService()
	svc.now = func() time.Time { return base }
	if err := svc.SignUp(context.Background(), validInput); err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if _, err := svc.VerifyEmail(context.Background(), tokenFrom(t, mailer.sent[0].url)); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
	mailer.sent = nil
	return svc, store, mailer
}

func signIn(t *testing.T, svc *Service) Session {
	t.Helper()
	sess, err := svc.SignIn(context.Background(), SignInInput{Email: "george@example.com", Password: "a-long-password"})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	return sess
}

func TestSignInCreatesThirtyDaySession(t *testing.T) {
	svc, store, _ := newVerifiedTestService(t)

	sess, err := svc.SignIn(context.Background(), SignInInput{Email: " GEORGE@example.com ", Password: "a-long-password"})
	if err != nil {
		t.Fatalf("SignIn: %v", err)
	}
	if sess.Token == "" {
		t.Fatal("session token is empty")
	}
	if want := base.Add(30 * day); !sess.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", sess.ExpiresAt, want)
	}
	stored, ok := store.sessions[string(token.Hash(sess.Token))]
	if !ok {
		t.Fatal("session must be stored by the hash of its token")
	}
	if !stored.expiresAt.Equal(sess.ExpiresAt) {
		t.Error("stored expiry does not match the returned one")
	}
}

func TestSignInRejectsWrongPasswordAndUnknownEmail(t *testing.T) {
	cases := []SignInInput{
		{Email: "george@example.com", Password: "wrong-password"},
		{Email: "nobody@example.com", Password: "a-long-password"},
	}
	for _, in := range cases {
		svc, store, mailer := newVerifiedTestService(t)

		if _, err := svc.SignIn(context.Background(), in); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("SignIn(%s) error = %v; want ErrInvalidCredentials", in.Email, err)
		}
		if len(store.sessions) != 0 {
			t.Error("no session must be created")
		}
		if len(mailer.sent) != 0 {
			t.Error("no email must be sent")
		}
	}
}

func TestSignInUnverifiedWithCorrectPasswordSendsNewLink(t *testing.T) {
	svc, store, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	oldToken := tokenFrom(t, mailer.sent[0].url)
	mailer.sent = nil

	_, err := svc.SignIn(context.Background(), SignInInput{Email: "george@example.com", Password: "a-long-password"})
	if !errors.Is(err, ErrEmailNotVerified) {
		t.Fatalf("error = %v; want ErrEmailNotVerified", err)
	}
	if len(mailer.sent) != 1 || mailer.sent[0].kind != "verification" {
		t.Fatalf("sent = %+v, want one verification email", mailer.sent)
	}
	if _, err := svc.VerifyEmail(context.Background(), oldToken); !errors.Is(err, ErrInvalidToken) {
		t.Error("the old verification link must stop working")
	}
	if len(store.sessions) != 0 {
		t.Error("an unverified user must not get a session")
	}
}

func TestSignInUnverifiedWithWrongPasswordSendsNothing(t *testing.T) {
	svc, _, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	mailer.sent = nil

	_, err := svc.SignIn(context.Background(), SignInInput{Email: "george@example.com", Password: "wrong-password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error = %v; want ErrInvalidCredentials", err)
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("sent = %+v, want no emails", mailer.sent)
	}
}

func TestStartSessionAfterVerification(t *testing.T) {
	svc, _, mailer := newTestService()
	_ = svc.SignUp(context.Background(), validInput)
	userID, err := svc.VerifyEmail(context.Background(), tokenFrom(t, mailer.sent[0].url))
	if err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}

	sess, err := svc.StartSession(context.Background(), userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	res, err := svc.Authenticate(context.Background(), sess.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if res.User.Email != "george@example.com" || res.User.Role != "user" {
		t.Errorf("user = %+v", res.User)
	}
}

func TestAuthenticateDoesNotRenewFreshSession(t *testing.T) {
	svc, _, _ := newVerifiedTestService(t)
	sess := signIn(t, svc)

	svc.now = func() time.Time { return base.Add(10 * day) } // quedan 20 días
	res, err := svc.Authenticate(context.Background(), sess.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if res.Renewed || !res.ExpiresAt.Equal(base.Add(30*day)) {
		t.Errorf("res = %+v; want no renewal", res)
	}
}

func TestAuthenticateRenewsWithLessThanFifteenDaysLeft(t *testing.T) {
	svc, store, _ := newVerifiedTestService(t)
	sess := signIn(t, svc)

	now := base.Add(16 * day) // quedan 14 días
	svc.now = func() time.Time { return now }
	res, err := svc.Authenticate(context.Background(), sess.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	want := now.Add(30 * day)
	if !res.Renewed || !res.ExpiresAt.Equal(want) {
		t.Errorf("res = %+v; want renewal until %v", res, want)
	}
	if !store.sessions[string(token.Hash(sess.Token))].expiresAt.Equal(want) {
		t.Error("renewal must be saved in the store")
	}
}

func TestAuthenticateDoesNotRenewAtExactlyFifteenDays(t *testing.T) {
	svc, _, _ := newVerifiedTestService(t)
	sess := signIn(t, svc)

	svc.now = func() time.Time { return base.Add(15 * day) }
	res, err := svc.Authenticate(context.Background(), sess.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if res.Renewed {
		t.Error("with exactly 15 days left the session must not renew")
	}
}

func TestAuthenticateRejectsExpiredAndUnknownTokens(t *testing.T) {
	svc, _, _ := newVerifiedTestService(t)
	sess := signIn(t, svc)

	svc.now = func() time.Time { return base.Add(30 * day) }
	for _, tok := range []string{sess.Token, "", "not-a-real-token"} {
		if _, err := svc.Authenticate(context.Background(), tok); !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("Authenticate(%q) error = %v; want ErrUnauthenticated", tok, err)
		}
	}
}

func TestSignOutInvalidatesSessionAndIsIdempotent(t *testing.T) {
	svc, _, _ := newVerifiedTestService(t)
	sess := signIn(t, svc)

	if err := svc.SignOut(context.Background(), sess.Token); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	if _, err := svc.Authenticate(context.Background(), sess.Token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("error = %v; want ErrUnauthenticated after sign-out", err)
	}
	if err := svc.SignOut(context.Background(), sess.Token); err != nil {
		t.Fatalf("second SignOut = %v; want nil", err)
	}
}

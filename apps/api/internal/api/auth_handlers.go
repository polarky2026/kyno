package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/polarky2026/kyno/apps/api/internal/auth"
)

const (
	invalidBody    = "Invalid request body"
	signInRequired = "Sign in to continue."
)

func (s *Server) SignUp(ctx context.Context, req SignUpRequestObject) (SignUpResponseObject, error) {
	if req.Body == nil {
		return SignUp400JSONResponse{Error: invalidBody}, nil
	}

	err := s.auth.SignUp(ctx, auth.SignUpInput{
		Name:     req.Body.Name,
		Email:    req.Body.Email,
		Password: req.Body.Password,
	})

	var validationErr *auth.ValidationError
	if errors.As(err, &validationErr) {
		return SignUp400JSONResponse{Error: validationErr.Message}, nil
	}
	if err != nil {
		return nil, err
	}

	return SignUp202Response{}, nil
}

func (s *Server) VerifyEmail(ctx context.Context, req VerifyEmailRequestObject) (VerifyEmailResponseObject, error) {
	if req.Body == nil {
		return VerifyEmail400JSONResponse{Error: invalidBody}, nil
	}

	userID, err := s.auth.VerifyEmail(ctx, req.Body.Token)
	if errors.Is(err, auth.ErrInvalidToken) {
		return VerifyEmail400JSONResponse{Error: "This link is invalid or has expired."}, nil
	}
	if err != nil {
		return nil, err
	}

	sess, err := s.auth.StartSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp, err := s.signedIn(ctx, sess, req.Params.XKynoClient)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) ResendVerification(ctx context.Context, req ResendVerificationRequestObject) (ResendVerificationResponseObject, error) {
	if req.Body != nil {
		if err := s.auth.ResendVerification(ctx, req.Body.Email); err != nil {
			return nil, err
		}
	}

	return ResendVerification202Response{}, nil
}

const iosClient = "ios"

func (s *Server) SignIn(ctx context.Context, req SignInRequestObject) (SignInResponseObject, error) {
	if req.Body == nil {
		return SignIn400JSONResponse{Error: invalidBody}, nil
	}

	sess, err := s.auth.SignIn(ctx, auth.SignInInput{
		Email:    req.Body.Email,
		Password: req.Body.Password,
	})
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		return SignIn401JSONResponse{Error: "Incorrect email or password."}, nil
	case errors.Is(err, auth.ErrEmailNotVerified):
		return SignIn403JSONResponse{Error: "Verify your email to sign in. We sent you a new link."}, nil
	case err != nil:
		return nil, err
	}

	resp, err := s.signedIn(ctx, sess, req.Params.XKynoClient)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Server) SignOut(ctx context.Context, _ SignOutRequestObject) (SignOutResponseObject, error) {
	if rs := sessionFrom(ctx); rs.token != "" {
		if err := s.auth.SignOut(ctx, rs.token); err != nil {
			return nil, err
		}
	}
	return signOutResponse{cookie: s.expiredSessionCookie()}, nil
}

func (s *Server) GetMe(ctx context.Context, _ GetMeRequestObject) (GetMeResponseObject, error) {
	rs := sessionFrom(ctx)
	if rs.result == nil {
		return GetMe401JSONResponse{Error: signInRequired}, nil
	}
	return GetMe200JSONResponse{User: toAPIUser(rs.result.User)}, nil
}

// signedIn builds the response for a new session: a cookie for web clients,
// the token in the body for iOS.
func (s *Server) signedIn(ctx context.Context, sess auth.Session, client *string) (authResponse, error) {
	result, err := s.auth.Authenticate(ctx, sess.Token)
	if err != nil {
		return authResponse{}, fmt.Errorf("load signed-in user: %w", err)
	}

	body := AuthResponse{User: toAPIUser(result.User)}
	if client != nil && *client == iosClient {
		body.Session = &SessionToken{Token: sess.Token, ExpiresAt: sess.ExpiresAt}
		return authResponse{body: body}, nil
	}
	return authResponse{body: body, cookie: s.sessionCookie(sess.Token, sess.ExpiresAt)}, nil
}

func toAPIUser(u auth.User) User {
	return User{Id: u.ID, Name: u.Name, Email: u.Email, Role: UserRole(u.Role)}
}

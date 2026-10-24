package api

import (
	"context"
	"errors"

	"github.com/polarky2026/kyno/apps/api/internal/auth"
)

const invalidBody = "Invalid request body"

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

	_, err := s.auth.VerifyEmail(ctx, req.Body.Token)
	if errors.Is(err, auth.ErrInvalidToken) {
		return VerifyEmail400JSONResponse{Error: "This link is invalid or has expired."}, nil
	}
	if err != nil {
		return nil, err
	}

	return VerifyEmail204Response{}, nil
}

func (s *Server) ResendVerification(ctx context.Context, req ResendVerificationRequestObject) (ResendVerificationResponseObject, error) {
	if req.Body != nil {
		if err := s.auth.ResendVerification(ctx, req.Body.Email); err != nil {
			return nil, err
		}
	}

	return ResendVerification202Response{}, nil
}

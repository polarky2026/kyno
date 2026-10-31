package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/polarky2026/kyno/apps/api/internal/auth"
	"github.com/polarky2026/kyno/apps/api/internal/platform/httpx"
)

const sessionCookieName = "kyno_session"

type requestSession struct {
	token      string
	fromCookie bool
	result     *auth.AuthResult
}

type sessionKey struct{}

// withSession authenticates the request when it carries a session token and
// stores the result in the context. It never rejects a request: handlers that
// need a signed-in user check for one themselves.
func (s *Server) withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs := &requestSession{}
		rs.token, rs.fromCookie = sessionToken(r)

		if rs.token != "" && s.auth != nil {
			result, err := s.auth.Authenticate(r.Context(), rs.token)
			switch {
			case err == nil:
				rs.result = &result
				if result.Renewed && rs.fromCookie {
					http.SetCookie(w, s.sessionCookie(rs.token, result.ExpiresAt))
				}
			case errors.Is(err, auth.ErrUnauthenticated):
				if rs.fromCookie {
					http.SetCookie(w, s.expiredSessionCookie())
				}
			default:
				slog.Error("failed to authenticate request", "error", err)
				httpx.WriteError(w, http.StatusInternalServerError, "Something went wrong")
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, rs)))
	})
}

func sessionFrom(ctx context.Context) *requestSession {
	if rs, ok := ctx.Value(sessionKey{}).(*requestSession); ok {
		return rs
	}
	return &requestSession{}
}

func sessionToken(r *http.Request) (token string, fromCookie bool) {
	if header := r.Header.Get("Authorization"); header != "" {
		if token, ok := strings.CutPrefix(header, "Bearer "); ok {
			return strings.TrimSpace(token), false
		}
		return "", false
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		return cookie.Value, true
	}
	return "", false
}

func (s *Server) sessionCookie(token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Server) expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
}

// authResponse writes the signed-in user and, for web clients, the session cookie.
type authResponse struct {
	body   AuthResponse
	cookie *http.Cookie
}

func (r authResponse) write(w http.ResponseWriter) error {
	if r.cookie != nil {
		http.SetCookie(w, r.cookie)
	}
	httpx.WriteJSON(w, http.StatusOK, r.body)
	return nil
}

func (r authResponse) VisitSignInResponse(w http.ResponseWriter) error      { return r.write(w) }
func (r authResponse) VisitVerifyEmailResponse(w http.ResponseWriter) error { return r.write(w) }

type signOutResponse struct {
	cookie *http.Cookie
}

func (r signOutResponse) VisitSignOutResponse(w http.ResponseWriter) error {
	http.SetCookie(w, r.cookie)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

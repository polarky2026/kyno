package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()

	NewHandler(Services{}, Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status field = %q, want ok", body.Status)
	}
}

func TestUnknownAPIRouteReturnsJSON404(t *testing.T) {
	rec := httptest.NewRecorder()

	NewHandler(Services{}, Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
}

func TestSignUpRejectsMalformedJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up", strings.NewReader("{not json"))
	req.Header.Set("Content-Type", "application/json")

	NewHandler(Services{}, Options{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
}

func TestMeWithoutSessionReturns401(t *testing.T) {
	rec := httptest.NewRecorder()

	NewHandler(Services{}, Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
}

func TestSearchTitlesWithoutSessionReturns401(t *testing.T) {
	rec := httptest.NewRecorder()

	NewHandler(Services{}, Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/titles/search?query=dune", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestSearchTitlesWithoutQueryReturnsJSON400(t *testing.T) {
	rec := httptest.NewRecorder()

	NewHandler(Services{}, Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/titles/search", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
}

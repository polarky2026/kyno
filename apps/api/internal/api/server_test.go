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

func TestTitleEndpointsWithoutSessionReturn401(t *testing.T) {
	requests := []*http.Request{
		httptest.NewRequest(http.MethodPost, "/api/titles", strings.NewReader(`{"kind":"movie","tmdbId":438631}`)),
		httptest.NewRequest(http.MethodGet, "/api/titles/6e6f7a45-1b2c-4d3e-8f90-123456789abc", nil),
	}
	for _, req := range requests {
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		NewHandler(Services{}, Options{}).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s status = %d, want 401", req.Method, req.URL.Path, rec.Code)
		}
	}
}

func TestGetTitleWithInvalidIDReturnsJSON400(t *testing.T) {
	rec := httptest.NewRecorder()

	NewHandler(Services{}, Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/titles/not-a-uuid", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
}

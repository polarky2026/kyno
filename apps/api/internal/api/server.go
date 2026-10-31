package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/polarky2026/kyno/apps/api/internal/auth"
	"github.com/polarky2026/kyno/apps/api/internal/platform/httpx"
)

const maxRequestBodyBytes = 1 << 20

type Options struct {
	SecureCookies bool
}

type Server struct {
	auth          *auth.Service
	secureCookies bool
}

func NewHandler(authService *auth.Service, opts Options) http.Handler {
	server := &Server{auth: authService, secureCookies: opts.SecureCookies}

	strict := NewStrictHandlerWithOptions(server, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  handleRequestError,
		ResponseErrorHandlerFunc: handleResponseError,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", notFound)

	return http.MaxBytesHandler(server.withSession(HandlerFromMux(strict, mux)), maxRequestBodyBytes)
}

func (s *Server) GetHealth(_ context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "ok"}, nil
}

func handleRequestError(w http.ResponseWriter, _ *http.Request, _ error) {
	httpx.WriteError(w, http.StatusBadRequest, "Invalid request body")
}

func handleResponseError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "Something went wrong")
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteError(w, http.StatusNotFound, "Not found")
}

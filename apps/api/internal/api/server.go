package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/polarky2026/kyno/apps/api/internal/auth"
	"github.com/polarky2026/kyno/apps/api/internal/platform/httpx"
)

const maxRequestBodyBytes = 1 << 20

type Server struct {
	auth *auth.Service
}

var _ StrictServerInterface = (*Server)(nil)

func NewHandler(authService *auth.Service) http.Handler {
	server := &Server{auth: authService}

	strict := NewStrictHandlerWithOptions(server, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  handleRequestError,
		ResponseErrorHandlerFunc: handleResponseError,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", notFound)

	return http.MaxBytesHandler(HandlerFromMux(strict, mux), maxRequestBodyBytes)
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

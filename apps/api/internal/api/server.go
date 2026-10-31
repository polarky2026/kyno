package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/polarky2026/kyno/apps/api/internal/auth"
	"github.com/polarky2026/kyno/apps/api/internal/platform/httpx"
	"github.com/polarky2026/kyno/apps/api/internal/titles"
)

const maxRequestBodyBytes = 1 << 20

// Services are the application services the HTTP layer calls into.
type Services struct {
	Auth   *auth.Service
	Titles *titles.Service
}

type Options struct {
	SecureCookies bool
}

type Server struct {
	auth          *auth.Service
	titles        *titles.Service
	secureCookies bool
}

func NewHandler(services Services, opts Options) http.Handler {
	server := &Server{
		auth:          services.Auth,
		titles:        services.Titles,
		secureCookies: opts.SecureCookies,
	}

	strict := NewStrictHandlerWithOptions(server, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  handleRequestError,
		ResponseErrorHandlerFunc: handleResponseError,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", notFound)

	routes := HandlerWithOptions(strict, StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: handleRequestError,
	})

	return http.MaxBytesHandler(server.withSession(routes), maxRequestBodyBytes)
}

func (s *Server) GetHealth(_ context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "ok"}, nil
}

// handleRequestError answers requests the generated code could not parse,
// such as malformed JSON or a missing required query parameter.
func handleRequestError(w http.ResponseWriter, _ *http.Request, _ error) {
	httpx.WriteError(w, http.StatusBadRequest, "Invalid request")
}

func handleResponseError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "Something went wrong")
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteError(w, http.StatusNotFound, "Not found")
}

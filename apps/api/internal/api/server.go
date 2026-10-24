package api

import (
	"context"
	"net/http"

	"github.com/polarky2026/kyno/apps/api/internal/platform/httpx"
)

type Server struct{}

var _ StrictServerInterface = (*Server)(nil)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", notFound)

	return HandlerFromMux(NewStrictHandler(&Server{}, nil), mux)
}

func (s *Server) GetHealth(_ context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{Status: "ok"}, nil
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteError(w, http.StatusNotFound, "Not found")
}
package api

import (
	"net/http"

	"github.com/polarky2026/kyno/apps/api/internal/platform/httpx"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", health)
	mux.HandleFunc("/api/", notFound)

	return mux
}

type healthResponse struct {
	Status string `json:"status"`
}

func health(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteError(w, http.StatusNotFound, "Not found")
}

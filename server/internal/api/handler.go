package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/skryde/booking-check/server/internal/repository"
)

type Handler struct {
	db repository.Repository
}

func NewHandler(db repository.Repository) *Handler {
	return &Handler{db: db}
}

func (h Handler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	subs, err := h.db.Subscribers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status": 500, "message":"internal error"}`))

		slog.Error("error getting subscribers", slog.Any("error", err))
	}

	response, err := json.Marshal(subs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status": 500, "message":"internal error"}`))

		slog.Error("error marshalling response", slog.Any("error", err))
	}

	_, _ = w.Write(response)
}

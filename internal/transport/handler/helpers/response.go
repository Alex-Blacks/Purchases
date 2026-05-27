package helpers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
)

func WriteJSON(w http.ResponseWriter, logger *slog.Logger, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("encoding response failed", "error", err)
	}
}

func WriteError(w http.ResponseWriter, logger *slog.Logger, status int, msg string) {
	logger.Warn("request failed",
		"status", status,
		"error", msg,
	)
	WriteJSON(w, logger, status, dto.ErrorResponse{Error: msg})
}
func WriteInternalError(w http.ResponseWriter, logger *slog.Logger, err error, req any) {
	logger.Error("request failed",
		"error", err,
		"request", req,
	)
	WriteJSON(w, logger, http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
}

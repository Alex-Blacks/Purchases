package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func parseIntQuery(r *http.Request, key string) (int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return 0, fmt.Errorf("%s required", key)
	}
	keyInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s must be int", key)
	}
	return keyInt, nil
}

func getIntQuery(w http.ResponseWriter, r *http.Request, name string, logger *slog.Logger) (int, bool) {
	val, err := parseIntQuery(r, name)
	if err != nil {
		logger.Info("invalid query parameter", "error", err, "param", name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	return val, true
}

func domainErrResponse(w http.ResponseWriter, err error, logger *slog.Logger, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		logger.Info("already exists", "error", err, "details", details)
		w.WriteHeader(http.StatusConflict)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "conflict"}); encodeErr != nil {
			logger.Error("failed to encode error response", "error", encodeErr)
		}

	case errors.Is(err, domain.ErrNotFound):
		logger.Info("not found", "error", err, "details", details)
		w.WriteHeader(http.StatusNotFound)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "not found"}); encodeErr != nil {
			logger.Error("failed to encode error response", "error", encodeErr)
		}

	case errors.Is(err, domain.ErrInvalidInput):
		logger.Info("invalid input", "error", err, "details", details)
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "invalid input"}); encodeErr != nil {
			logger.Error("failed to encode error response", "error", encodeErr)
		}

	default:
		logger.Error("internal server error", "error", err, "details", details)
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "internal server error"}); encodeErr != nil {
			logger.Error("failed to encode error response", "error", encodeErr)
		}
	}
}

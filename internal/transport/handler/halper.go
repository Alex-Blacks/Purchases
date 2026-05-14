package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/go-chi/chi/v5"
)

func parsePositiveIntParam(w http.ResponseWriter, r *http.Request, name string, logger *slog.Logger) (int, bool) {
	valStr := chi.URLParam(r, name)
	if strings.TrimSpace(valStr) == "" {
		err := fmt.Errorf("%s must not be empty", name)
		logger.Info("invalid parameter", "error", err, "param", name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		err := fmt.Errorf("%s must be int", name)
		logger.Info("invalid parameter", "error", err, "param", name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	if val <= 0 {
		logger.Info("invalid input", "param", name)
		http.Error(w, "invalid input:"+name+"must be > 0", http.StatusBadRequest)
		return 0, false
	}
	return val, true
}

func validatePositiveInt(w http.ResponseWriter, name string, val int, logger *slog.Logger) bool {
	if val <= 0 {
		logger.Info("invalid input", "param", name)
		http.Error(w, "invalid input:"+name+"must be > 0", http.StatusBadRequest)
		return false
	}
	return true
}

func decodeHelper(w http.ResponseWriter, r *http.Request, logger *slog.Logger, req any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(req); err != nil {
		logger.Info("decoding failed", "error", err)
		http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func encodeHelper(w http.ResponseWriter, logger *slog.Logger, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("encoding response failed", "error", err)
	}
}

func domainErrResponse(w http.ResponseWriter, err error, logger *slog.Logger, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case errors.Is(err, domain.ErrConflict):
		logger.Info("conflict", "error", err, "details", details)
		w.WriteHeader(http.StatusConflict)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "the field is used in another table"}); encodeErr != nil {
			logger.Error("failed to encode error response", "error", encodeErr)
		}
	case errors.Is(err, domain.ErrAlreadyExists):
		logger.Info("already exists", "error", err, "details", details)
		w.WriteHeader(http.StatusConflict)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "conflict"}); encodeErr != nil {
			logger.Error("failed to encode error response", "error", encodeErr)
		}
	case errors.Is(err, domain.ErrEmptyName):
		logger.Info("empty name", "error", err, "details", details)
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"status": "empty name"}); encodeErr != nil {
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

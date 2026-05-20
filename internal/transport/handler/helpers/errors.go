package helpers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func DomainErrResponse(w http.ResponseWriter, err error, logger *slog.Logger, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")

	type errData struct {
		Code int
		Msg  string
	}

	errorMap := map[error]errData{
		domain.ErrEmailConflict: {http.StatusConflict, "email has already been created"},
		domain.ErrConflict:      {http.StatusConflict, "the field is used in another table"},
		domain.ErrAlreadyExists: {http.StatusConflict, "conflict"},
		domain.ErrEmptyName:     {http.StatusBadRequest, "empty name"},
		domain.ErrNotFound:      {http.StatusNotFound, "not found"},
		domain.ErrInvalidInput:  {http.StatusBadRequest, "invalid input"},
	}

	for domainErr, data := range errorMap {
		if errors.Is(err, domainErr) {
			logger.Warn("domain error", "error", err, "details", details)
			w.WriteHeader(data.Code)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": data.Msg})
			return
		}
	}

	// default
	logger.Error("internal server error", "error", err, "details", details)
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "internal server error"})
}

func AuthErrResponse(w http.ResponseWriter, err error, logger *slog.Logger, details map[string]any) {
	switch {
	case errors.Is(err, domain.ErrStatusBlocked) || errors.Is(err, domain.ErrIncorrectPassword):
		logger.Warn("auth error", "error", err, "details", details)
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "invalid credentials"})
		return
	default:
		logger.Error("internal server error", "error", err, "details", details)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "internal server error"})
		return
	}
}

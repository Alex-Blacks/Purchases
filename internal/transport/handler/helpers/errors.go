package helpers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Alex-Blacks/Purchases/internal/domain"
)

func WriteDomainError(w http.ResponseWriter, logger *slog.Logger, err error, req any) {
	type errData struct {
		Code int
		Msg  string
	}

	errorMap := map[error]errData{
		domain.ErrEmailConflict:     {http.StatusConflict, "email has already been created"},
		domain.ErrConflict:          {http.StatusConflict, "the field is used in another table"},
		domain.ErrAlreadyExists:     {http.StatusConflict, "conflict"},
		domain.ErrEmptyName:         {http.StatusBadRequest, "empty name"},
		domain.ErrNotFound:          {http.StatusNotFound, "not found"},
		domain.ErrInvalidInput:      {http.StatusBadRequest, "invalid input"},
		domain.ErrStatusBlocked:     {http.StatusUnauthorized, "unauthorized"},
		domain.ErrIncorrectPassword: {http.StatusUnauthorized, "unauthorized"},
	}

	for domainErr, data := range errorMap {
		if errors.Is(err, domainErr) {
			WriteError(w, logger, data.Code, data.Msg)
			return
		}
	}

	WriteInternalError(w, logger, err, req)
}

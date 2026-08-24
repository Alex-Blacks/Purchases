package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func ParsePositiveIntParam(r *http.Request, name string) (int, error) {
	valStr := chi.URLParam(r, name)
	if strings.TrimSpace(valStr) == "" {
		return 0, fmt.Errorf("%s must not be empty", name)
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return val, nil
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, validate *validator.Validate, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dest); err != nil {
		logger.WarnContext(r.Context(), "decode failed", "error", err)
		return fmt.Errorf("invalid json")
	}

	if dec.More() {
		logger.WarnContext(r.Context(), "multiple json objects")
		return fmt.Errorf("body must contain single json object")
	}

	if err := validate.Struct(dest); err != nil {
		logger.WarnContext(r.Context(), err.Error())
		return err
	}
	return nil
}

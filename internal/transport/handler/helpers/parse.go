package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5"
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

func DecodeJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dest); err != nil {
		logger.Warn("decode failed", "error", err)
		return fmt.Errorf("invalid json")
	}

	if err := dec.Decode(&struct{}{}); err != nil {
		logger.Warn("multiple json objects", "error", err)
		return fmt.Errorf("body must contain single json object")
	}
	return nil
}

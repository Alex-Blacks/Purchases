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

// parsePositiveIntParam парсит URL параметр и проверяет, что >0
func ParsePositiveIntParam(w http.ResponseWriter, r *http.Request, name string, logger *slog.Logger) (int, bool) {
	valStr := chi.URLParam(r, name)
	if strings.TrimSpace(valStr) == "" {
		err := fmt.Errorf("%s must not be empty", name)
		logger.Warn("invalid parameter", "error", err, "param", name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		err := fmt.Errorf("%s must be a positive integer", name)
		logger.Warn("invalid parameter", "error", err, "param", name)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	return val, true
}

// DecodeJSONHelper декодирует тело запроса в структуру
func DecodeJSONHelper(w http.ResponseWriter, r *http.Request, logger *slog.Logger, dest any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dest); err != nil {
		logger.Warn("decoding failed", "error", err)
		http.Error(w, "bad request: invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

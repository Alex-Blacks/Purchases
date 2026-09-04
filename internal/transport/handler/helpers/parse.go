package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/Alex-Blacks/Purchases/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

// ParsePositiveIntParam парсит и валидирует целочисленный параметр из URL (например, /users/{id})
// Возвращает ошибку, обёрнутую в domain.ErrInvalidInput.
func ParsePositiveIntParam(r *http.Request, name string) (int, error) {
	valStr := chi.URLParam(r, name)
	if strings.TrimSpace(valStr) == "" {
		return 0, domain.ErrInvalidInput
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return 0, domain.ErrInvalidInput
	}
	return val, nil
}

// ParsePositiveIntQuery парсит и валидирует целочисленный параметр из Query (например, /users/by-product?productId)
// Возвращает ошибку, обёрнутую в domain.ErrInvalidInput.
func ParsePositiveIntQuery(r *http.Request, name string) (int, error) {
	valStr := r.URL.Query().Get(name)
	if strings.TrimSpace(valStr) == "" {
		return 0, domain.ErrInvalidInput
	}
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return 0, domain.ErrInvalidInput
	}
	return val, nil
}

// ParsePositiveStrQuery парсит и валидирует строчный параметр из Query (например, /users/by-product?productId)
// Возвращает ошибку, обёрнутую в domain.ErrInvalidInput.
func ParsePositiveStrQuery(r *http.Request, name string) (string, error) {
	valStr := r.URL.Query().Get(name)
	if strings.TrimSpace(valStr) == "" {
		return "", domain.ErrInvalidInput
	}
	return valStr, nil
}

// ParseOptionalIntParam парсит опциональный целочисленный параметр из query (например, ?limit=10)
// Если параметр отсутствует – возвращает nil. Если присутствует, но невалидный – возвращает ошибку.
func ParseOptionalIntParam(r *http.Request, key string) (*int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	if i < 1 {
		return nil, domain.ErrInvalidInput
	}
	return &i, nil
}

// DecodeJSON декодирует JSON из тела запроса.
// Проверяет: размер (1 МБ), отсутствие неизвестных полей, единственный JSON-объект.
// При успехе – валидирует структуру с помощью validate.
// Возвращает domain.ErrInvalidInput для ошибок декодирования, а для ошибок валидации – validator.ValidationErrors.
func DecodeJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, validate *validator.Validate, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dest); err != nil {
		logger.WarnContext(r.Context(), "decode failed", "error", err)
		return domain.ErrInvalidInput
	}

	if dec.More() {
		logger.WarnContext(r.Context(), "multiple json objects")
		return domain.ErrInvalidInput
	}

	if err := validate.Struct(dest); err != nil {
		logger.WarnContext(r.Context(), err.Error())
		return err
	}
	return nil
}

var FormDecoder *form.Decoder

func init() {
	FormDecoder = form.NewDecoder()
	FormDecoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		if len(vals) == 0 || vals[0] == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, vals[0])
		if err != nil {
			return nil, fmt.Errorf("invalid time format (RFC3339): %w", err)
		}
		return &t, nil
	}, (*time.Time)(nil))
}

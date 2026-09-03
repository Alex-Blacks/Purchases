package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type CountResponse struct {
	Count int `json:"count"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// Кастомная валидация для пары полей From/To
func ValidateDateRange(fl validator.FieldLevel) bool {
	// Получаем значение текущего поля (например, CreatedFrom)
	fromPtr, ok := fl.Field().Interface().(*time.Time)
	if !ok {
		return false
	}

	// Получаем родительскую структуру, чтобы добраться до парного поля
	parent := fl.Parent()
	// Имя парного поля берём из параметра тега (например, "CreatedTo")
	toFieldName := fl.Param() // параметр, переданный в тег: validate:"daterange=CreatedTo"
	toField := parent.FieldByName(toFieldName)
	if !toField.IsValid() {
		return false
	}
	toPtr, ok := toField.Interface().(*time.Time)
	if !ok {
		return false
	}

	now := time.Now().UTC()

	// Случаи
	switch {
	case fromPtr == nil && toPtr == nil:
		return true
	case fromPtr != nil && toPtr == nil:
		return fromPtr.Before(now)
	case fromPtr == nil && toPtr != nil:
		return toPtr.After(now)
	default: // оба не nil
		return fromPtr.Before(*toPtr)
	}
}

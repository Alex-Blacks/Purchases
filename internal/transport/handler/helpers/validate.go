package helpers

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
)

// ValidateStruct проверяет поля с тегом `validate:"required"`
func ValidateStruct(s any) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("validate")
		name := t.Field(i).Name

		if tag == "required" {
			empty := false
			switch field.Kind() {
			case reflect.String:
				empty = strings.TrimSpace(field.String()) == ""
			case reflect.Int:
				empty = field.Int() == 0
			}
			if empty {
				return fmt.Errorf("%s must not be empty", name)
			}
		}
	}
	return nil
}

func ValidatePositiveInt(w http.ResponseWriter, name string, val int, logger *slog.Logger) bool {
	if val <= 0 {
		logger.Warn("invalid input", "param", name)
		http.Error(w, "invalid input:"+name+"must be > 0", http.StatusBadRequest)
		return false
	}
	return true
}

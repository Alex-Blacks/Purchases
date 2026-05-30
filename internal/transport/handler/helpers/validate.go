package helpers

import (
	"fmt"
	"strings"

	"github.com/Alex-Blacks/Purchases/internal/transport/handler/dto"
)

func ValidateCreateUser(input dto.UserRequest) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("password must not be empty")
	}
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email must not be empty")
	}
	if input.Role != nil {
		if strings.TrimSpace(*input.Role) == "" {
			return fmt.Errorf("role must not be empty")
		}
	}

	return nil
}

func ValidatePositiveInt(name string, val int) error {
	if val <= 0 {
		return fmt.Errorf("invalid input: %s must be > 0", name)
	}
	return nil
}

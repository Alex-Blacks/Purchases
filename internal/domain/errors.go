package domain

import "errors"

// Общие ошибки
var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrNoFieldsToUpdate  = errors.New("no fields to update")
	ErrEmptyName         = errors.New("empty name")
	ErrGroupIDRequired   = errors.New("group ID is required for admin")
	ErrInvalidGroupID    = errors.New("group ID must be positive")
	ErrGroupIDNotAllowed = errors.New("group ID cannot be specified by non-admin")
)

// Ошибки конфликтов
var (
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrEmailConflict = errors.New("email has already been created")
)

// Ошибки приглашений
var (
	ErrUserAlreadyInGroup       = errors.New("user already in group")
	ErrSelfInvite               = errors.New("cannot invite yourself")
	ErrInviteRejected           = errors.New("invite was rejected")
	ErrInviteExpired            = errors.New("invite has expired")
	ErrInviteRejectedStillValid = errors.New("rejected invite is still valid")
	ErrInviteAlreadyPending     = errors.New("invite already pending")
)

// Ошибки статуса и аутентификации
var (
	ErrStatusBlocked      = errors.New("status blocked")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// IsNotFound возвращает true, если ошибка является ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

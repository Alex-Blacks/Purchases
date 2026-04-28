package domain

import "errors"

var (
	ErrEmptyName = errors.New("empty name")
	ErrInvalidId = errors.New("invalid id")
	ErrNotFound  = errors.New("not found")
)

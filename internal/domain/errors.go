package domain

import "errors"

var (
	ErrEmptyName     = errors.New("empty name")
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

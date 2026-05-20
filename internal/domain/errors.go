package domain

import "errors"

var (
	ErrEmptyName        = errors.New("empty name")
	ErrInvalidInput     = errors.New("invalid input")
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrConflict         = errors.New("conflict")
	ErrNoFieldsToUpdate = errors.New("no fields to update")

	ErrStatusBlocked     = errors.New("status blocked")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrNoRights          = errors.New("no rights")
	ErrEmailConflict     = errors.New("email has already been created")
)

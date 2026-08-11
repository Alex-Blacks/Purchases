package domain

import "errors"

var (
	ErrEmptyName                = errors.New("empty name")
	ErrInvalidInput             = errors.New("invalid input")
	ErrNotFound                 = errors.New("not found")
	ErrAlreadyExists            = errors.New("already exists")
	ErrConflict                 = errors.New("conflict")
	ErrConflictGroups           = errors.New("conflict groups")
	ErrNoFieldsToUpdate         = errors.New("no fields to update")
	ErrUserAlreadyInGroup       = errors.New("user already in group")
	ErrSelfInvite               = errors.New("cannot invite yourself")
	ErrInviteRejected           = errors.New("invite was rejected")
	ErrInviteExpired            = errors.New("invite has expired")
	ErrInviteRejectedStillValid = errors.New("rejected invite is still valid")
	ErrInviteAlreadyPending     = errors.New("invite already pending")

	ErrStatusBlocked      = errors.New("status blocked")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoRights           = errors.New("no rights")
	ErrEmailConflict      = errors.New("email has already been created")
)

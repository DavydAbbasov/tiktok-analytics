package models

import "errors"

// Sentinel errors
var (
	// ErrNotFound indicates a resource was not found
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidRequest indicates invalid request data
	ErrInvalidRequest = errors.New("invalid request")

	// ErrUnauthorized indicates authentication failure
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates authorization failure
	ErrForbidden = errors.New("forbidden")

	// ErrInternalError indicates internal server error
	ErrInternalError = errors.New("internal error")

	// ErrConflict indicates a conflict with current state
	ErrConflict = errors.New("conflict")
)

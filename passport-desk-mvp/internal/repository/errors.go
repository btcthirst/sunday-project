package repository

import "errors"

var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates the resource already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabaseError indicates a database error occurred
	ErrDatabaseError = errors.New("database error")
)

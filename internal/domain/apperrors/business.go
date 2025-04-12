package apperrors

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Reception business errors
var (
	ErrReceptionAlreadyClosed    = errors.New("reception is already closed")
	ErrReceptionCannotBeModified = errors.New("closed reception cannot be modified")
	ErrActiveReceptionExists     = errors.New("cannot create a new reception while previous one is not closed")
	ErrNoActiveReception         = errors.New("no active reception for this PVZ")
)

// Product business errors
var (
	ErrNoProductsToDelete = errors.New("no products to delete in the current reception")
)

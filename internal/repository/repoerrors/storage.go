package repoerrors

import (
	"errors"
)

// Generic storage errors
var (
	ErrConnectionFailed  = errors.New("database connection failed")
	ErrQueryFailed       = errors.New("database query failed")
	ErrTransactionFailed = errors.New("database transaction failed")
)

// User storage errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// PVZ storage errors
var (
	ErrPVZNotFound      = errors.New("pickup point not found")
	ErrPVZAlreadyExists = errors.New("pickup point with this ID already exists")
)

// Reception storage errors
var (
	ErrReceptionNotFound      = errors.New("reception not found")
	ErrReceptionAlreadyExists = errors.New("reception with this ID already exists")
)

// Product storage errors
var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product with this ID already exists")
)

// IsDuplicateKeyError checks if the error is due to a duplicate key
func IsDuplicateKeyError(err error) bool {
	return err != nil && (errors.Is(err, ErrUserAlreadyExists) ||
		errors.Is(err, ErrPVZAlreadyExists) ||
		errors.Is(err, ErrReceptionAlreadyExists) ||
		errors.Is(err, ErrProductAlreadyExists))
}

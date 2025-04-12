package apperrors

import "errors"

// User validation errors
var (
	ErrEmailRequired    = errors.New("email is required")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordRequired = errors.New("password is required")
	ErrInvalidPassword  = errors.New("password must be at least 6 characters long")
	ErrInvalidRole      = errors.New("invalid role, must be 'employee' or 'moderator'")
)

// PVZ validation errors
var (
	ErrCityRequired = errors.New("city is a required field")
	ErrInvalidCity  = errors.New("invalid city, only Moscow, St. Petersburg and Kazan are allowed")
	ErrInvalidPVZID = errors.New("invalid pickup point ID")
)

// Reception validation errors
var (
	ErrInvalidReceptionID = errors.New("invalid reception ID")
)

// Product validation errors
var (
	ErrProductTypeRequired = errors.New("product type is required")
	ErrInvalidProductType  = errors.New("invalid product type, only electronics, clothes and shoes are allowed")
	ErrInvalidProductID    = errors.New("invalid product ID")
)

package models

import (
	"avito-backend-trainee-assignment-spring-2025/pkg/hasher"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	RoleEmployee  = "employee"
	RoleModerator = "moderator"
)

var (
	ErrEmailRequired      = errors.New("email is required")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrPasswordRequired   = errors.New("password is required")
	ErrInvalidPassword    = errors.New("password must be at least 6 characters long")
	ErrInvalidRole        = errors.New("invalid role, must be 'employee' or 'moderator'")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUser(email, password, role string) (*User, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	if password == "" {
		return nil, ErrPasswordRequired
	}

	if len(password) < 6 {
		return nil, ErrInvalidPassword
	}

	if role != RoleEmployee && role != RoleModerator {
		return nil, ErrInvalidRole
	}

	passwordHash, err := hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) IsEmployee() bool {
	return u.Role == RoleEmployee
}

func (u *User) IsModerator() bool {
	return u.Role == RoleModerator
}

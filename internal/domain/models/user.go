package models

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/pkg/hasher"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	RoleEmployee  = "employee"
	RoleModerator = "moderator"
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
		return nil, apperrors.ErrEmailRequired
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, apperrors.ErrInvalidEmail
	}

	if password == "" {
		return nil, apperrors.ErrPasswordRequired
	}

	if len(password) < 6 {
		return nil, apperrors.ErrInvalidPassword
	}

	if role != RoleEmployee && role != RoleModerator {
		return nil, apperrors.ErrInvalidRole
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

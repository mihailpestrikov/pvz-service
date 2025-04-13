package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/auth"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/hasher"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type UserService struct {
	repo      interfaces.TxUserRepository
	jwtConfig config.JWTConfig
	txManager postgres.TxManager
}

func NewUserService(
	repo interfaces.TxUserRepository,
	jwtConfig config.JWTConfig,
	txManager postgres.TxManager,
) *UserService {
	return &UserService{
		repo:      repo,
		jwtConfig: jwtConfig,
		txManager: txManager,
	}
}

func (s *UserService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	var user *models.User

	err := s.txManager.RunTransaction(ctx, func(tx *sql.Tx) error {
		txRepo := s.repo.WithTx(tx)

		_, err := txRepo.GetByEmail(ctx, email)
		if err == nil {
			log.Info().
				Str("email", email).
				Msg("Registration failed: user already exists")
			return repoerrors.ErrUserAlreadyExists
		}

		if !errors.Is(err, repoerrors.ErrUserNotFound) {
			return fmt.Errorf("failed to check if user exists: %w", err)
		}

		newUser, err := models.NewUser(email, password, role)
		if err != nil {
			log.Info().
				Err(err).
				Str("email", email).
				Msg("User validation failed during registration")
			return err
		}

		if err := txRepo.Create(ctx, newUser); err != nil {
			return fmt.Errorf("failed to save user: %w", err)
		}

		user = newUser
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info().
		Str("id", user.ID.String()).
		Str("email", email).
		Str("role", role).
		Msg("User registered successfully")

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repoerrors.ErrUserNotFound) {
			log.Info().
				Str("email", email).
				Msg("Login failed: user not found")
			return "", apperrors.ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to get user by email: %w", err)
	}

	if !hasher.Verify(user.PasswordHash, password) {
		log.Info().
			Str("email", email).
			Msg("Login failed: invalid password")
		return "", apperrors.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, user.Role, s.jwtConfig.Secret, s.jwtConfig.Expiration)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", user.ID.String()).
			Msg("Failed to generate JWT token")
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Info().
		Str("user_id", user.ID.String()).
		Str("email", email).
		Msg("User logged in successfully")
	return token, nil
}

func (s *UserService) DummyLogin(role string) (string, error) {
	if role != models.RoleEmployee && role != models.RoleModerator {
		log.Info().
			Str("role", role).
			Msg("Dummy login failed: invalid role")
		return "", apperrors.ErrInvalidRole
	}

	token, err := auth.GenerateDummyToken(role, s.jwtConfig.Secret, s.jwtConfig.Expiration)
	if err != nil {
		log.Error().
			Err(err).
			Str("role", role).
			Msg("Failed to generate dummy JWT token")
		return "", fmt.Errorf("failed to generate dummy token: %w", err)
	}

	log.Info().
		Str("role", role).
		Msg("Dummy token generated successfully")
	return token, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.txManager.RunTransaction(ctx, func(tx *sql.Tx) error {
		txRepo := s.repo.WithTx(tx)

		_, err := txRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		if err := txRepo.Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		log.Info().
			Str("user_id", id.String()).
			Msg("User deleted successfully")
		return nil
	})
}

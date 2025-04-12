package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/auth"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/hasher"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService struct {
	repo      UserRepository
	jwtConfig config.JWTConfig
}

func NewUserService(repo UserRepository, jwtConfig config.JWTConfig) *UserService {
	return &UserService{
		repo:      repo,
		jwtConfig: jwtConfig,
	}
}

func (s *UserService) Register(ctx context.Context, email, password, role string) (*models.User, error) {
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, models.ErrUserAlreadyExists
	}
	if !errors.Is(err, models.ErrUserNotFound) {
		log.Error().Err(err).Str("email", email).Msg("Failed to check if user exists")
		return nil, err
	}

	user, err := models.NewUser(email, password, role)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to create user")
		return nil, err
	}

	if err := s.repo.Create(ctx, user); err != nil {
		log.Error().Err(err).Str("email", email).Msg("Failed to save user")
		return nil, err
	}

	log.Info().Str("id", user.ID.String()).Str("email", email).Str("role", role).Msg("User registered successfully")

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return "", models.ErrInvalidCredentials
		}
		log.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return "", err
	}

	if !hasher.Verify(user.PasswordHash, password) {
		log.Warn().Str("email", email).Msg("Invalid login attempt: wrong password")
		return "", models.ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, user.Role, s.jwtConfig.Secret, s.jwtConfig.Expiration)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to generate token")
		return "", err
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", email).Msg("User logged in successfully")
	return token, nil
}

func (s *UserService) DummyLogin(role string) (string, error) {
	if role != models.RoleEmployee && role != models.RoleModerator {
		return "", models.ErrInvalidRole
	}

	token, err := auth.GenerateDummyToken(role, s.jwtConfig.Secret, s.jwtConfig.Expiration)
	if err != nil {
		log.Error().Err(err).Str("role", role).Msg("Failed to generate dummy token")
		return "", err
	}

	log.Info().Str("role", role).Msg("Dummy token generated successfully")
	return token, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to get user by ID")
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to delete user")
		return err
	}

	log.Info().Str("user_id", id.String()).Msg("User deleted successfully")
	return nil
}

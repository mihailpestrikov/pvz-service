package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
)

type UserRepository struct {
	db *DB
	sb squirrel.StatementBuilderType
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := r.sb.Insert("users").
		Columns("id", "email", "password_hash", "role").
		Values(user.ID, user.Email, user.PasswordHash, user.Role)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error().Err(err).
			Str("email", user.Email).
			Str("role", user.Role).
			Msg("Failed to create user")

		if isDuplicateKeyError(err) {
			return models.ErrUserAlreadyExists
		}

		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Info().
		Str("id", user.ID.String()).
		Str("email", user.Email).
		Str("role", user.Role).
		Msg("User created successfully")

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := r.sb.Select("id", "email", "password_hash", "role").
		From("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	user := &models.User{}
	err = row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := r.sb.Select("id", "email", "password_hash", "role").
		From("users").
		Where(squirrel.Eq{"email": email})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	user := &models.User{}
	err = row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.sb.Delete("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrUserNotFound
	}

	log.Info().
		Str("id", id.String()).
		Msg("User deleted successfully")

	return nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`
}

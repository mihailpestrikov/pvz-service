package postgres

import (
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
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
		log.Error().Err(err).Msg("Failed to build SQL query for user creation")
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return repoerrors.ErrUserAlreadyExists
		}
		log.Error().Err(err).
			Str("email", user.Email).
			Str("user_id", user.ID.String()).
			Msg("Database error during user creation")
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := r.sb.Select("id", "email", "password_hash", "role").
		From("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for user retrieval by ID")
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
			return nil, repoerrors.ErrUserNotFound
		}
		log.Error().Err(err).
			Str("user_id", id.String()).
			Msg("Database error while scanning user row")
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
		log.Error().Err(err).Msg("Failed to build SQL query for user retrieval by email")
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
			return nil, repoerrors.ErrUserNotFound
		}
		log.Error().Err(err).
			Str("email", email).
			Msg("Database error while scanning user row")
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.sb.Delete("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for user deletion")
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", id.String()).
			Msg("Database error during user deletion")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).
			Str("user_id", id.String()).
			Msg("Failed to get rows affected after user deletion")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repoerrors.ErrUserNotFound
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`
}

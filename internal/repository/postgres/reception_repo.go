package postgres

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
)

type ReceptionRepository struct {
	db *DB
	sb squirrel.StatementBuilderType
}

func NewReceptionRepository(db *DB) *ReceptionRepository {
	return &ReceptionRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ReceptionRepository) Create(ctx context.Context, reception *models.Reception) error {
	query := r.sb.Insert("reception").
		Columns("id", "date_time", "pvz_id", "status").
		Values(reception.ID, reception.DateTime, reception.PVZID, reception.Status)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error().Err(err).
			Str("pvz_id", reception.PVZID.String()).
			Str("status", reception.Status).
			Msg("Failed to create reception")

		if isDuplicateKeyError(err) {
			return repoerrors.ErrReceptionAlreadyExists
		}

		return fmt.Errorf("failed to create reception: %w", err)
	}

	log.Info().
		Str("id", reception.ID.String()).
		Str("pvz_id", reception.PVZID.String()).
		Str("status", reception.Status).
		Msg("Reception created successfully")

	metrics.ReceptionsCreatedTotal.Inc()

	return nil
}

func (r *ReceptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	query := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("reception").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	reception := &models.Reception{}
	err = row.Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.ErrReceptionNotFound
		}
		return nil, fmt.Errorf("failed to get reception by ID: %w", err)
	}

	reception.Products, err = r.getProductsForReception(ctx, reception.ID)
	if err != nil {
		log.Error().Err(err).
			Str("reception_id", reception.ID.String()).
			Msg("Failed to get products for reception")
	}

	return reception, nil
}

func (r *ReceptionRepository) GetLastActiveByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	query := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("reception").
		Where(squirrel.Eq{"pvz_id": pvzID, "status": models.ReceptionStatusInProgress}).
		OrderBy("date_time DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	reception := &models.Reception{}
	err = row.Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNoActiveReception
		}
		return nil, fmt.Errorf("failed to get active reception for PVZ: %w", err)
	}

	reception.Products, err = r.getProductsForReception(ctx, reception.ID)
	if err != nil {
		log.Error().Err(err).
			Str("reception_id", reception.ID.String()).
			Msg("Failed to get products for reception")
	}

	return reception, nil
}

func (r *ReceptionRepository) CloseReception(ctx context.Context, id uuid.UUID) error {
	reception, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if reception.Status == models.ReceptionStatusClosed {
		return apperrors.ErrReceptionAlreadyClosed
	}

	query := r.sb.Update("reception").
		Set("status", models.ReceptionStatusClosed).
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to close reception: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repoerrors.ErrReceptionNotFound
	}

	log.Info().
		Str("id", id.String()).
		Msg("Reception closed successfully")

	return nil
}

func (r *ReceptionRepository) GetLastReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	query := r.sb.Select("id", "date_time", "pvz_id", "status").
		From("reception").
		Where(squirrel.Eq{"pvz_id": pvzID}).
		OrderBy("date_time DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	reception := &models.Reception{}
	err = row.Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PVZID,
		&reception.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.ErrReceptionNotFound
		}
		return nil, fmt.Errorf("failed to get last reception for PVZ: %w", err)
	}

	reception.Products, err = r.getProductsForReception(ctx, reception.ID)
	if err != nil {
		log.Error().Err(err).
			Str("reception_id", reception.ID.String()).
			Msg("Failed to get products for reception")
	}

	return reception, nil
}

func (r *ReceptionRepository) getProductsForReception(ctx context.Context, receptionID uuid.UUID) ([]models.Product, error) {
	query := r.sb.Select("id", "date_time", "type", "reception_id").
		From("product").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time ASC")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.DateTime,
			&product.Type,
			&product.ReceptionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through product rows: %w", err)
	}

	return products, nil
}

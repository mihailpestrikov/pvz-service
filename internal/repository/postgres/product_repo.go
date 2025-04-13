package postgres

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ProductRepository struct {
	db Querier
	sb squirrel.StatementBuilderType
}

func NewProductRepository(db Querier) interfaces.TxProductRepository {
	return &ProductRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ProductRepository) WithTx(tx *sql.Tx) interfaces.ProductRepository {
	return &ProductRepository{
		db: tx,
		sb: r.sb,
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := r.sb.Insert("product").
		Columns("id", "date_time", "type", "reception_id").
		Values(product.ID, product.DateTime, product.Type, product.ReceptionID)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for product creation")
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		if isDuplicateKeyError(err) {
			return repoerrors.ErrProductAlreadyExists
		}

		log.Error().Err(err).
			Str("product_id", product.ID.String()).
			Str("type", product.Type).
			Str("reception_id", product.ReceptionID.String()).
			Msg("Database error during product creation")

		return fmt.Errorf("failed to create product: %w", err)
	}

	metrics.ProductsAddedTotal.Inc()

	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := r.sb.Select("id", "date_time", "type", "reception_id").
		From("product").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for product retrieval")
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	product := &models.Product{}
	err = row.Scan(
		&product.ID,
		&product.DateTime,
		&product.Type,
		&product.ReceptionID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.ErrProductNotFound
		}

		log.Error().Err(err).
			Str("product_id", id.String()).
			Msg("Database error while scanning product row")
		return nil, fmt.Errorf("failed to get product by ID: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) DeleteLastFromReception(ctx context.Context, receptionID uuid.UUID) error {
	query := r.sb.Select("id").
		From("product").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for retrieving last product")
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	var productID uuid.UUID
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repoerrors.ErrProductNotFound
		}

		log.Error().Err(err).
			Str("reception_id", receptionID.String()).
			Msg("Database error while retrieving last product")
		return fmt.Errorf("failed to get last product from reception: %w", err)
	}

	deleteQuery := r.sb.Delete("product").
		Where(squirrel.Eq{"id": productID})

	sqlQuery, args, err = deleteQuery.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for product deletion")
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error().Err(err).
			Str("product_id", productID.String()).
			Msg("Database error while deleting product")
		return fmt.Errorf("failed to delete last product: %w", err)
	}

	return nil
}

func (r *ProductRepository) GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]models.Product, error) {
	query := r.sb.Select("id", "date_time", "type", "reception_id").
		From("product").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time ASC")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error().Err(err).Msg("Failed to build SQL query for product retrieval by reception ID")
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error().Err(err).
			Str("reception_id", receptionID.String()).
			Msg("Database error while querying products by reception ID")
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
			log.Error().Err(err).
				Str("reception_id", receptionID.String()).
				Msg("Database error while scanning product row")
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		log.Error().Err(err).
			Str("reception_id", receptionID.String()).
			Msg("Error while iterating product rows")
		return nil, fmt.Errorf("error iterating through product rows: %w", err)
	}

	return products, nil
}

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
	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
)

type PVZRepository struct {
	db *DB
	sb squirrel.StatementBuilderType
}

func NewPVZRepository(db *DB) *PVZRepository {
	return &PVZRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *PVZRepository) Create(ctx context.Context, pvz *models.PVZ) error {
	query := r.sb.Insert("pvz").
		Columns("id", "registration_date", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error().Err(err).
			Str("city", pvz.City).
			Msg("Failed to create PVZ")

		if isDuplicateKeyError(err) {
			return repoerrors.ErrPVZAlreadyExists
		}

		return fmt.Errorf("failed to create PVZ: %w", err)
	}

	log.Info().
		Str("id", pvz.ID.String()).
		Str("city", pvz.City).
		Time("registration_date", pvz.RegistrationDate).
		Msg("PVZ created successfully")

	metrics.PVZCreatedTotal.Inc()

	return nil
}

func (r *PVZRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	query := r.sb.Select("id", "registration_date", "city").
		From("pvz").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	row := r.db.QueryRowContext(ctx, sqlQuery, args...)

	pvz := &models.PVZ{}
	err = row.Scan(
		&pvz.ID,
		&pvz.RegistrationDate,
		&pvz.City,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repoerrors.ErrPVZNotFound
		}
		return nil, fmt.Errorf("failed to get PVZ by ID: %w", err)
	}

	return pvz, nil
}

func (r *PVZRepository) GetAll(ctx context.Context, filter models.PVZFilter) ([]*models.PVZ, int, error) {
	countQuery := r.sb.Select("COUNT(*)").From("pvz")

	selectQuery := r.sb.Select("id", "registration_date", "city").From("pvz")

	if filter.StartDate != nil && filter.EndDate != nil {
		countQuery = countQuery.Where(`
			EXISTS (
				SELECT 1 FROM reception 
				WHERE reception.pvz_id = pvz.id 
				AND reception.date_time BETWEEN ? AND ?
			)`, filter.StartDate, filter.EndDate)

		selectQuery = selectQuery.Where(`
			EXISTS (
				SELECT 1 FROM reception 
				WHERE reception.pvz_id = pvz.id 
				AND reception.date_time BETWEEN ? AND ?
			)`, filter.StartDate, filter.EndDate)
	}

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count SQL query: %w", err)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countSql, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count PVZs: %w", err)
	}

	if filter.Limit > 0 {
		selectQuery = selectQuery.Limit(uint64(filter.Limit))
	} else {
		selectQuery = selectQuery.Limit(10)
	}

	if filter.Page > 0 {
		offset := uint64((filter.Page - 1) * filter.Limit)
		selectQuery = selectQuery.Offset(offset)
	}

	sqlQuery, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query PVZs: %w", err)
	}
	defer rows.Close()

	var pvzs []*models.PVZ
	for rows.Next() {
		pvz := &models.PVZ{}
		err := rows.Scan(
			&pvz.ID,
			&pvz.RegistrationDate,
			&pvz.City,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan PVZ row: %w", err)
		}
		pvzs = append(pvzs, pvz)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating through PVZ rows: %w", err)
	}

	return pvzs, total, nil
}

func (r *PVZRepository) GetAllWithReceptions(ctx context.Context, filter models.PVZFilter) ([]models.PVZWithReceptions, int, error) {
	pvzs, total, err := r.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if len(pvzs) == 0 {
		return []models.PVZWithReceptions{}, total, nil
	}

	var pvzIDs []interface{}
	for _, pvz := range pvzs {
		pvzIDs = append(pvzIDs, pvz.ID)
	}

	receptionQuery := r.sb.Select(
		"r.id",
		"r.date_time",
		"r.pvz_id",
		"r.status",
	).
		From("reception r").
		Where(squirrel.Eq{"r.pvz_id": pvzIDs})

	if filter.StartDate != nil && filter.EndDate != nil {
		receptionQuery = receptionQuery.Where(squirrel.And{
			squirrel.GtOrEq{"r.date_time": filter.StartDate},
			squirrel.LtOrEq{"r.date_time": filter.EndDate},
		})
	}

	receptionSQL, receptionArgs, err := receptionQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build receptions SQL query: %w", err)
	}

	receptionRows, err := r.db.QueryContext(ctx, receptionSQL, receptionArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query receptions: %w", err)
	}
	defer receptionRows.Close()

	pvzReceptions := make(map[uuid.UUID][]*models.Reception)

	for receptionRows.Next() {
		reception := &models.Reception{}
		err := receptionRows.Scan(
			&reception.ID,
			&reception.DateTime,
			&reception.PVZID,
			&reception.Status,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan reception row: %w", err)
		}

		pvzReceptions[reception.PVZID] = append(pvzReceptions[reception.PVZID], reception)
	}

	if err = receptionRows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating through reception rows: %w", err)
	}

	var receptionIDs []interface{}
	for _, receptions := range pvzReceptions {
		for _, reception := range receptions {
			receptionIDs = append(receptionIDs, reception.ID)
		}
	}

	var result []models.PVZWithReceptions
	if len(receptionIDs) == 0 {
		for _, pvz := range pvzs {
			result = append(result, models.PVZWithReceptions{
				PVZ:        pvz,
				Receptions: []*models.Reception{},
			})
		}
		return result, total, nil
	}

	productQuery := r.sb.Select(
		"p.id",
		"p.date_time",
		"p.type",
		"p.reception_id",
	).
		From("product p").
		Where(squirrel.Eq{"p.reception_id": receptionIDs})

	productSQL, productArgs, err := productQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build products SQL query: %w", err)
	}

	productRows, err := r.db.QueryContext(ctx, productSQL, productArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer productRows.Close()

	receptionProducts := make(map[uuid.UUID][]models.Product)

	for productRows.Next() {
		product := models.Product{}
		err := productRows.Scan(
			&product.ID,
			&product.DateTime,
			&product.Type,
			&product.ReceptionID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}

		receptionProducts[product.ReceptionID] = append(receptionProducts[product.ReceptionID], product)
	}

	if err = productRows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating through product rows: %w", err)
	}

	for _, receptions := range pvzReceptions {
		for _, reception := range receptions {
			reception.Products = receptionProducts[reception.ID]
		}
	}

	for _, pvz := range pvzs {
		result = append(result, models.PVZWithReceptions{
			PVZ:        pvz,
			Receptions: pvzReceptions[pvz.ID],
		})
	}

	return result, total, nil
}

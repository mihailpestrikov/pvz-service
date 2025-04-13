package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PVZService struct {
	repo      interfaces.TxPVZRepository
	txManager postgres.TxManager
}

func NewPVZService(
	repo interfaces.TxPVZRepository,
	txManager postgres.TxManager,
) *PVZService {
	return &PVZService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *PVZService) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	if !models.IsValidCity(city) {
		log.Info().
			Str("city", city).
			Msg("PVZ creation failed: invalid city")
		return nil, apperrors.ErrInvalidCity
	}

	var pvz *models.PVZ

	err := s.txManager.RunTransaction(ctx, func(tx *sql.Tx) error {
		txRepo := s.repo.WithTx(tx)

		newPvz, err := models.NewPVZ(city)
		if err != nil {
			log.Info().
				Err(err).
				Str("city", city).
				Msg("PVZ creation failed: invalid data")
			return err
		}

		if err := txRepo.Create(ctx, newPvz); err != nil {
			return fmt.Errorf("failed to save PVZ: %w", err)
		}

		pvz = newPvz
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info().
		Str("pvz_id", pvz.ID.String()).
		Str("city", pvz.City).
		Time("registration_date", pvz.RegistrationDate).
		Msg("PVZ created successfully")

	return pvz, nil
}

func (s *PVZService) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	pvz, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get PVZ: %w", err)
	}
	return pvz, nil
}

func (s *PVZService) GetAllPVZ(ctx context.Context, filter models.PVZFilter) ([]*models.PVZ, int, error) {
	return s.repo.GetAll(ctx, filter)
}

func (s *PVZService) GetAllPVZWithReceptions(ctx context.Context, filter models.PVZFilter) ([]models.PVZWithReceptions, int, error) {
	pvzList, total, err := s.repo.GetAllWithReceptions(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get PVZ list with receptions: %w", err)
	}

	log.Info().
		Int("total_count", total).
		Int("returned_count", len(pvzList)).
		Int("page", filter.Page).
		Int("limit", filter.Limit).
		Msg("Retrieved PVZ list with receptions successfully")

	return pvzList, total, nil
}

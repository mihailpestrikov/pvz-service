package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PVZRepository interface {
	Create(ctx context.Context, pvz *models.PVZ) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error)
	GetAll(ctx context.Context, filter models.PVZFilter) ([]*models.PVZ, int, error)
	GetAllWithReceptions(ctx context.Context, filter models.PVZFilter) ([]models.PVZWithReceptions, int, error)
}

type PVZService struct {
	repo PVZRepository
}

func NewPVZService(repo PVZRepository) *PVZService {
	return &PVZService{
		repo: repo,
	}
}

func (s *PVZService) CreatePVZ(ctx context.Context, city string) (*models.PVZ, error) {
	if !models.IsValidCity(city) {
		log.Info().
			Str("city", city).
			Msg("PVZ creation failed: invalid city")
		return nil, apperrors.ErrInvalidCity
	}

	pvz, err := models.NewPVZ(city)
	if err != nil {
		log.Info().
			Err(err).
			Str("city", city).
			Msg("PVZ creation failed: invalid data")
		return nil, err
	}

	if err := s.repo.Create(ctx, pvz); err != nil {
		return nil, fmt.Errorf("failed to save PVZ: %w", err)
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

package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"github.com/google/uuid"
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
		return nil, models.ErrInvalidCity
	}

	pvz, err := models.NewPVZ(city)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, pvz); err != nil {
		return nil, err
	}

	return pvz, nil
}

func (s *PVZService) GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PVZService) GetAllPVZ(ctx context.Context, filter models.PVZFilter) ([]*models.PVZ, int, error) {
	return s.repo.GetAll(ctx, filter)
}

func (s *PVZService) GetAllPVZWithReceptions(ctx context.Context, filter models.PVZFilter) ([]models.PVZWithReceptions, int, error) {
	return s.repo.GetAllWithReceptions(ctx, filter)
}

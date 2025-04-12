package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"errors"
	"github.com/google/uuid"
)

type ReceptionRepository interface {
	Create(ctx context.Context, reception *models.Reception) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Reception, error)
	GetLastActiveByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetLastReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, id uuid.UUID) error
}

type ReceptionService struct {
	receptionRepo ReceptionRepository
	pvzRepo       PVZRepository
}

func NewReceptionService(receptionRepo ReceptionRepository, pvzRepo PVZRepository) *ReceptionService {
	return &ReceptionService{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
	}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	_, err := s.pvzRepo.GetByID(ctx, pvzID)
	if err != nil {
		return nil, err
	}

	_, err = s.receptionRepo.GetLastActiveByPVZID(ctx, pvzID)
	if err == nil {
		return nil, apperrors.ErrActiveReceptionExists
	}
	if !errors.Is(err, apperrors.ErrNoActiveReception) {
		return nil, err
	}

	reception, err := models.NewReception(pvzID)
	if err != nil {
		return nil, err
	}

	if err := s.receptionRepo.Create(ctx, reception); err != nil {
		return nil, err
	}

	return reception, nil
}

func (s *ReceptionService) GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	return s.receptionRepo.GetByID(ctx, id)
}

func (s *ReceptionService) GetLastActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	return s.receptionRepo.GetLastActiveByPVZID(ctx, pvzID)
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	reception, err := s.receptionRepo.GetLastActiveByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}

	if err := s.receptionRepo.CloseReception(ctx, reception.ID); err != nil {
		return nil, err
	}

	return s.receptionRepo.GetByID(ctx, reception.ID)
}

func (s *ReceptionService) GetLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	return s.receptionRepo.GetLastReceptionByPVZID(ctx, pvzID)
}

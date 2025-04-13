package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/interfaces"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type ReceptionService struct {
	receptionRepo interfaces.TxReceptionRepository
	pvzRepo       interfaces.TxPVZRepository
	txManager     postgres.TxManager
}

func NewReceptionService(
	receptionRepo interfaces.TxReceptionRepository,
	pvzRepo interfaces.TxPVZRepository,
	txManager postgres.TxManager,
) *ReceptionService {
	return &ReceptionService{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		txManager:     txManager,
	}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	var reception *models.Reception

	err := s.txManager.RunTransaction(ctx, func(tx *sql.Tx) error {
		txPvzRepo := s.pvzRepo.WithTx(tx)
		txReceptionRepo := s.receptionRepo.WithTx(tx)

		_, err := txPvzRepo.GetByID(ctx, pvzID)
		if err != nil {
			return err
		}

		_, err = txReceptionRepo.GetLastActiveByPVZID(ctx, pvzID)
		if err == nil {
			return apperrors.ErrActiveReceptionExists
		}
		if !errors.Is(err, apperrors.ErrNoActiveReception) {
			return err
		}

		newReception, err := models.NewReception(pvzID)
		if err != nil {
			return err
		}

		if err := txReceptionRepo.Create(ctx, newReception); err != nil {
			return err
		}

		reception = newReception
		return nil
	})

	if err != nil {
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
	var closedReceptionID uuid.UUID

	err := s.txManager.RunTransaction(ctx, func(tx *sql.Tx) error {
		txReceptionRepo := s.receptionRepo.WithTx(tx)

		reception, err := txReceptionRepo.GetLastActiveByPVZID(ctx, pvzID)
		if err != nil {
			return err
		}

		if err := txReceptionRepo.CloseReception(ctx, reception.ID); err != nil {
			return err
		}

		closedReceptionID = reception.ID
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.receptionRepo.GetByID(ctx, closedReceptionID)
}

func (s *ReceptionService) GetLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	return s.receptionRepo.GetLastReceptionByPVZID(ctx, pvzID)
}

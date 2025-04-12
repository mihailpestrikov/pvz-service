package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]models.Product, error)
	DeleteLastFromReception(ctx context.Context, receptionID uuid.UUID) error
}

type ProductService struct {
	productRepo   ProductRepository
	receptionRepo ReceptionRepository
}

func NewProductService(productRepo ProductRepository, receptionRepo ReceptionRepository) *ProductService {
	return &ProductService{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
	}
}

func (s *ProductService) AddProduct(ctx context.Context, productType string, pvzID uuid.UUID) (*models.Product, error) {
	if !models.IsValidProductType(productType) {
		return nil, apperrors.ErrInvalidProductType
	}

	reception, err := s.receptionRepo.GetLastActiveByPVZID(ctx, pvzID)
	if err != nil {
		return nil, err
	}

	if !reception.IsInProgress() {
		return nil, apperrors.ErrReceptionCannotBeModified
	}

	product, err := models.NewProduct(productType, reception.ID)
	if err != nil {
		return nil, err
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]models.Product, error) {
	return s.productRepo.GetByReceptionID(ctx, receptionID)
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	reception, err := s.receptionRepo.GetLastActiveByPVZID(ctx, pvzID)
	if err != nil {
		return err
	}

	if !reception.IsInProgress() {
		return apperrors.ErrReceptionCannotBeModified
	}

	products, err := s.productRepo.GetByReceptionID(ctx, reception.ID)
	if err != nil {
		return err
	}

	if len(products) == 0 {
		return apperrors.ErrNoProductsToDelete
	}

	return s.productRepo.DeleteLastFromReception(ctx, reception.ID)
}

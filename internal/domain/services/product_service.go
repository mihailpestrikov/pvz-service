package services

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
		log.Info().
			Str("product_type", productType).
			Str("pvz_id", pvzID.String()).
			Msg("Product validation failed: invalid product type")
		return nil, apperrors.ErrInvalidProductType
	}

	reception, err := s.receptionRepo.GetLastActiveByPVZID(ctx, pvzID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reception: %w", err)
	}

	if !reception.IsInProgress() {
		log.Info().
			Str("reception_id", reception.ID.String()).
			Str("pvz_id", pvzID.String()).
			Str("status", reception.Status).
			Msg("Cannot add product to closed reception")
		return nil, apperrors.ErrReceptionCannotBeModified
	}

	product, err := models.NewProduct(productType, reception.ID)
	if err != nil {
		log.Info().
			Err(err).
			Str("product_type", productType).
			Str("reception_id", reception.ID.String()).
			Msg("Failed to create product model")
		return nil, err
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	log.Info().
		Str("product_id", product.ID.String()).
		Str("type", product.Type).
		Str("reception_id", reception.ID.String()).
		Msg("Product added successfully")

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
		return fmt.Errorf("failed to get active reception: %w", err)
	}

	if !reception.IsInProgress() {
		log.Info().
			Str("reception_id", reception.ID.String()).
			Str("pvz_id", pvzID.String()).
			Str("status", reception.Status).
			Msg("Cannot delete product from closed reception")
		return apperrors.ErrReceptionCannotBeModified
	}

	products, err := s.productRepo.GetByReceptionID(ctx, reception.ID)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	if len(products) == 0 {
		log.Info().
			Str("reception_id", reception.ID.String()).
			Str("pvz_id", pvzID.String()).
			Msg("Cannot delete product: no products in reception")
		return apperrors.ErrNoProductsToDelete
	}

	if err := s.productRepo.DeleteLastFromReception(ctx, reception.ID); err != nil {
		return fmt.Errorf("failed to delete last product: %w", err)
	}

	log.Info().
		Str("reception_id", reception.ID.String()).
		Str("pvz_id", pvzID.String()).
		Msg("Last product deleted successfully from reception")

	return nil
}

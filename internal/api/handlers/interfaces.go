package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"github.com/google/uuid"
)

type UserServiceInterface interface {
	Register(ctx context.Context, email, password, role string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	DummyLogin(role string) (string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type PVZServiceInterface interface {
	CreatePVZ(ctx context.Context, city string) (*models.PVZ, error)
	GetPVZByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error)
	GetAllPVZ(ctx context.Context, filter models.PVZFilter) ([]*models.PVZ, int, error)
	GetAllPVZWithReceptions(ctx context.Context, filter models.PVZFilter) ([]models.PVZWithReceptions, int, error)
}

type ReceptionServiceInterface interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetReceptionByID(ctx context.Context, id uuid.UUID) (*models.Reception, error)
	GetLastActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
}

type ProductServiceInterface interface {
	AddProduct(ctx context.Context, productType string, pvzID uuid.UUID) (*models.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetProductsByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}

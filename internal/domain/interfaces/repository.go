package interfaces

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

/*
В отдельном пакете, чтобы избежать проблемы циклических импортов,
при организации кода по месту использования. Компромис. Можно было бы сделать
пакет адаптеров, который соединяет интерфейсы и реализации или организовать код
по принципу гексагональной архитектуры с портами. Но оба подхода сильно усложнили бы код.
*/

type TxProvider[T any] interface {
	WithTx(tx *sql.Tx) T
}

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetByReceptionID(ctx context.Context, receptionID uuid.UUID) ([]models.Product, error)
	DeleteLastFromReception(ctx context.Context, receptionID uuid.UUID) error
}

type TxProductRepository interface {
	ProductRepository
	TxProvider[ProductRepository]
}

type ReceptionRepository interface {
	Create(ctx context.Context, reception *models.Reception) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Reception, error)
	GetLastActiveByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetLastReceptionByPVZID(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, id uuid.UUID) error
}

type TxReceptionRepository interface {
	ReceptionRepository
	TxProvider[ReceptionRepository]
}

type PVZRepository interface {
	Create(ctx context.Context, pvz *models.PVZ) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PVZ, error)
	GetAll(ctx context.Context, filter models.PVZFilter) ([]*models.PVZ, int, error)
	GetAllWithReceptions(ctx context.Context, filter models.PVZFilter) ([]models.PVZWithReceptions, int, error)
}

type TxPVZRepository interface {
	PVZRepository
	TxProvider[PVZRepository]
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TxUserRepository interface {
	UserRepository
	TxProvider[UserRepository]
}

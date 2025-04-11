package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Типы товаров
const (
	ProductTypeElectronics = "электроника"
	ProductTypeClothes     = "одежда"
	ProductTypeShoes       = "обувь"
)

var AllowedProductTypes = []string{
	ProductTypeElectronics,
	ProductTypeClothes,
	ProductTypeShoes,
}

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductTypeRequired  = errors.New("product type is required")
	ErrInvalidProductType   = errors.New("invalid product type, only electronics, clothes and shoes are allowed")
	ErrProductAlreadyExists = errors.New("product with this ID already exists")
	ErrInvalidProductID     = errors.New("invalid product ID")
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionId"`
}

func NewProduct(productType string, receptionID uuid.UUID) (*Product, error) {
	if productType == "" {
		return nil, ErrProductTypeRequired
	}

	validType := false
	for _, allowedType := range AllowedProductTypes {
		if productType == allowedType {
			validType = true
			break
		}
	}

	if !validType {
		return nil, ErrInvalidProductType
	}

	if receptionID == uuid.Nil {
		return nil, ErrInvalidReceptionID
	}

	return &Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: receptionID,
	}, nil
}

func IsValidProductType(productType string) bool {
	for _, allowedType := range AllowedProductTypes {
		if productType == allowedType {
			return true
		}
	}
	return false
}

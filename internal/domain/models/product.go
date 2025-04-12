package models

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"time"

	"github.com/google/uuid"
)

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

type Product struct {
	ID          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionId"`
}

func NewProduct(productType string, receptionID uuid.UUID) (*Product, error) {
	if productType == "" {
		return nil, apperrors.ErrProductTypeRequired
	}

	validType := false
	for _, allowedType := range AllowedProductTypes {
		if productType == allowedType {
			validType = true
			break
		}
	}

	if !validType {
		return nil, apperrors.ErrInvalidProductType
	}

	if receptionID == uuid.Nil {
		return nil, apperrors.ErrInvalidReceptionID
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

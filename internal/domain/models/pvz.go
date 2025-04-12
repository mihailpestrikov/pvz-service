package models

import (
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"time"

	"github.com/google/uuid"
)

const (
	CityMoscow    = "Москва"
	CitySaintPete = "Санкт-Петербург"
	CityKazan     = "Казань"
)

var AllowedCities = []string{CityMoscow, CitySaintPete, CityKazan}

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type PVZWithReceptions struct {
	PVZ        *PVZ
	Receptions []*Reception
}

func NewPVZ(city string) (*PVZ, error) {
	if city == "" {
		return nil, apperrors.ErrCityRequired
	}

	validCity := false
	for _, allowedCity := range AllowedCities {
		if city == allowedCity {
			validCity = true
			break
		}
	}

	if !validCity {
		return nil, apperrors.ErrInvalidCity
	}

	return &PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             city,
	}, nil
}

func IsValidCity(city string) bool {
	for _, allowedCity := range AllowedCities {
		if city == allowedCity {
			return true
		}
	}
	return false
}

type PVZFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
}

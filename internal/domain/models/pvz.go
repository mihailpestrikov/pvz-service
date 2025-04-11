package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	CityMoscow    = "Москва"
	CitySaintPete = "Санкт-Петербург"
	CityKazan     = "Казань"
)

var AllowedCities = []string{CityMoscow, CitySaintPete, CityKazan}

var (
	ErrPVZNotFound      = errors.New("pickup point not found")
	ErrCityRequired     = errors.New("city is a required field")
	ErrInvalidCity      = errors.New("invalid city, only Moscow, St. Petersburg and Kazan are allowed")
	ErrPVZAlreadyExists = errors.New("pickup point with this ID already exists")
	ErrInvalidPVZID     = errors.New("invalid pickup point ID")
)

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

func NewPVZ(city string) (*PVZ, error) {
	if city == "" {
		return nil, ErrCityRequired
	}

	validCity := false
	for _, allowedCity := range AllowedCities {
		if city == allowedCity {
			validCity = true
			break
		}
	}

	if !validCity {
		return nil, ErrInvalidCity
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

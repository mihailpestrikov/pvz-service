package dto

import (
	"time"
)

type DummyLoginRequestDTO struct {
	Role string `json:"role" validate:"required,oneof=employee moderator"`
}

type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=employee moderator"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type PVZRequestDTO struct {
	City string `json:"city" validate:"required,oneof=Москва Санкт-Петербург Казань"`
}

type PVZListFilterDTO struct {
	StartDate *time.Time `json:"startDate" validate:"omitempty"`
	EndDate   *time.Time `json:"endDate" validate:"omitempty,gtfield=StartDate"`
	Page      int        `json:"page" validate:"omitempty,min=1"`
	Limit     int        `json:"limit" validate:"omitempty,min=1,max=30"`
}

type ReceptionCreateRequestDTO struct {
	PVZID string `json:"pvzId" validate:"required,uuid4"`
}

type ProductCreateRequestDTO struct {
	Type  string `json:"type" validate:"required,oneof=электроника одежда обувь"`
	PVZID string `json:"pvzId" validate:"required,uuid4"`
}

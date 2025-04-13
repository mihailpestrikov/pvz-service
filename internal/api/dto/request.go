package dto

import (
	"time"
)

type DummyLoginRequestDTO struct {
	Role string `json:"role" binding:"required,oneof=employee moderator"`
}

type RegisterRequestDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=employee moderator"`
}

type LoginRequestDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type PVZRequestDTO struct {
	City string `json:"city" binding:"required,oneof=Москва Санкт-Петербург Казань"`
}

type PVZListFilterDTO struct {
	StartDate *time.Time `json:"startDate" form:"startDate" binding:"omitempty"`
	EndDate   *time.Time `json:"endDate" form:"endDate" binding:"omitempty,gtfield=StartDate"`
	Page      int        `json:"page" form:"page" binding:"omitempty,min=1"`
	Limit     int        `json:"limit" form:"limit" binding:"omitempty,min=1,max=30"`
}

type ReceptionCreateRequestDTO struct {
	PVZID string `json:"pvzId" binding:"required,uuid4"`
}

type ProductCreateRequestDTO struct {
	Type  string `json:"type" binding:"required,oneof=электроника одежда обувь"`
	PVZID string `json:"pvzId" binding:"required,uuid4"`
}

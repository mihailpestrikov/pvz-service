package dto

import "time"

type TokenResponseDTO struct {
	Token string `json:"token"`
}

type UserResponseDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type PVZResponseDTO struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type ReceptionResponseDTO struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PVZID    string    `json:"pvzId"`
	Status   string    `json:"status"`
}

type ProductResponseDTO struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
}

type ProductWithinReceptionDTO struct {
	Product ProductResponseDTO `json:"product"`
}

type ReceptionWithProductsDTO struct {
	Reception ReceptionResponseDTO `json:"reception"`
	Products  []ProductResponseDTO `json:"products"`
}

type PVZWithReceptionsResponseDTO struct {
	PVZ        PVZResponseDTO             `json:"pvz"`
	Receptions []ReceptionWithProductsDTO `json:"receptions"`
}

type PVZListResponseDTO struct {
	Items      []PVZWithReceptionsResponseDTO `json:"items"`
	TotalCount int                            `json:"totalCount"`
	Page       int                            `json:"page"`
	Limit      int                            `json:"limit"`
}

type ErrorResponseDTO struct {
	Message string `json:"message"`
}

func NewErrorResponse(message string) ErrorResponseDTO {
	return ErrorResponseDTO{
		Message: message,
	}
}

package dto

type PVZListResponseDTO struct {
	Items      []PVZWithReceptionsResponseDTO `json:"items"`
	TotalCount int                            `json:"totalCount"`
	Page       int                            `json:"page"`
	Limit      int                            `json:"limit"`
}

type PVZWithReceptionsResponseDTO struct {
	PVZ        PVZ                        `json:"pvz"`
	Receptions []ReceptionWithProductsDTO `json:"receptions"`
}

type ReceptionWithProductsDTO struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}

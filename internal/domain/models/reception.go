package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	ReceptionStatusInProgress = "in_progress"
	ReceptionStatusClosed     = "close"
)

type Reception struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PVZID    uuid.UUID `json:"pvzId"`
	Status   string    `json:"status"`
	Products []Product `json:"products,omitempty"`
}

func NewReception(pvzID uuid.UUID) (*Reception, error) {
	if pvzID == uuid.Nil {
		return nil, errors.New("pickup point ID is required")
	}

	return &Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   ReceptionStatusInProgress,
		Products: []Product{},
	}, nil
}

func (r *Reception) IsInProgress() bool {
	return r.Status == ReceptionStatusInProgress
}

func (r *Reception) IsClosed() bool {
	return r.Status == ReceptionStatusClosed
}

func (r *Reception) ProductCount() int {
	return len(r.Products)
}

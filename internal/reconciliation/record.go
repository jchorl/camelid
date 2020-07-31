package reconciliation

import (
	"github.com/google/uuid"
)

type Status int

const (
	StatusCreated Status = iota
	StatusAccepted
)

type Record interface {
	GetID() string
	GetAlpacaOrderID() string
	GetStatus() Status
	SetAccepted(alpacaOrderID string)
}

type record struct {
	// need exported fields for the dynamo marshaler
	ID            string
	AlpacaOrderID string
	Status        Status
	Reconciled    bool
}

func NewRecord() Record {
	return &record{
		ID:     uuid.New().String(),
		Status: StatusCreated,
	}
}

func (r *record) GetID() string {
	return r.ID
}

func (r *record) GetAlpacaOrderID() string {
	return r.AlpacaOrderID
}

func (r *record) GetStatus() Status {
	return r.Status
}

func (r *record) SetAccepted(alpacaOrderID string) {
	r.AlpacaOrderID = alpacaOrderID
	r.Status = StatusAccepted
}

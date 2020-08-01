package reconciliation

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Status int

func (s Status) String() string {
	return strconv.Itoa(int(s))
}

const (
	StatusUnreconciled Status = iota
	StatusReconciled
)

type Record interface {
	GetID() string
	GetAlpacaOrderID() string
	GetStatus() Status
	GetCreatedAt() time.Time
	GetSubmittedAt() *time.Time
	GetReconciledAt() *time.Time
	SetAccepted(alpacaOrderID string)
}

type record struct {
	// need exported fields for the dynamo marshaler
	ID            string
	AlpacaOrderID string
	Status        Status

	CreatedAt    time.Time
	SubmittedAt  *time.Time
	ReconciledAt *time.Time
}

func NewRecord() Record {
	return &record{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Status:    StatusUnreconciled,
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

func (r *record) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r *record) GetSubmittedAt() *time.Time {
	return r.SubmittedAt
}

func (r *record) GetReconciledAt() *time.Time {
	return r.ReconciledAt
}

func (r *record) SetAccepted(alpacaOrderID string) {
	r.AlpacaOrderID = alpacaOrderID
	now := time.Now()
	r.SubmittedAt = &now
}

package reconciliation

import (
	"github.com/google/uuid"
)

type Status int

const (
	StatusSubmitted Status = iota
	StatusAccepted
)

type TradeRecord struct {
	id            string
	AlpacaOrderID string
	Status        Status

	reconciled bool
}

func NewTradeRecord() TradeRecord {
	return TradeRecord{
		id: uuid.New().String(),
	}
}

func (r TradeRecord) GetID() string {
	return r.id
}

package exchange

import "github.com/alpacahq/alpaca-trade-api-go/alpaca"

var _ Client = (*alpaca.Client)(nil)

package exchange

import "github.com/alpacahq/alpaca-trade-api-go/alpaca"

type Client interface {
	GetAccount() (*alpaca.Account, error)
	GetLastQuote(string) (*alpaca.LastQuoteResponse, error)
	GetOrder(string) (*alpaca.Order, error)
	PlaceOrder(alpaca.PlaceOrderRequest) (*alpaca.Order, error)
}

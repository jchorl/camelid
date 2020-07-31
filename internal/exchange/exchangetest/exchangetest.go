package exchangetest

import (
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/jchorl/camelid/internal/exchange"
)

var _ exchange.Client = (*MockClient)(nil)

type MockClient struct {
	accountID string
	quotes    map[string]*alpaca.LastQuoteResponse
	orderReqs []alpaca.PlaceOrderRequest
	orders    []*alpaca.Order
}

func NewMockClient(accountID string) *MockClient {
	return &MockClient{
		accountID: accountID,
		quotes:    map[string]*alpaca.LastQuoteResponse{},
	}
}

func (c *MockClient) GetAccount() (*alpaca.Account, error) {
	return &alpaca.Account{ID: c.accountID}, nil
}

func (c *MockClient) GetLastQuote(ticker string) (*alpaca.LastQuoteResponse, error) {
	if quote, ok := c.quotes[ticker]; ok {
		return quote, nil
	}

	return nil, fmt.Errorf("quote not found for %s", ticker)
}

func (c *MockClient) PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	c.orderReqs = append(c.orderReqs, req)
	order := &alpaca.Order{
		ID:            uuid.New().String(),
		ClientOrderID: uuid.New().String(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		SubmittedAt:   time.Now(),
		Symbol:        *req.AssetKey,
		Exchange:      "Class:us_equity",
		Qty:           req.Qty,
		FilledQty:     decimal.Decimal{},
		Type:          req.Type,
		Side:          req.Side,
		TimeInForce:   req.TimeInForce,
		Status:        "accepted",
	}
	c.orders = append(c.orders, order)
	return order, nil
}

// helpers, not part of the API
func (c *MockClient) SetQuote(ticker string, resp *alpaca.LastQuoteResponse) {
	c.quotes[ticker] = resp
}

func (c *MockClient) GetOrderReqs() []alpaca.PlaceOrderRequest {
	return c.orderReqs
}

func (c *MockClient) GetOrders() []*alpaca.Order {
	return c.orders
}

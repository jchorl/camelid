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
	cash      decimal.Decimal
	quotes    map[string]*alpaca.LastQuoteResponse
	orderReqs []alpaca.PlaceOrderRequest
	orders    []*alpaca.Order
	positions []alpaca.Position
}

func NewMockClient(accountID string) *MockClient {
	return &MockClient{
		accountID: accountID,
		quotes:    map[string]*alpaca.LastQuoteResponse{},
	}
}

func (c *MockClient) GetAccount() (*alpaca.Account, error) {
	return &alpaca.Account{
		ID:   c.accountID,
		Cash: c.cash,
	}, nil
}

func (c *MockClient) GetLastQuote(ticker string) (*alpaca.LastQuoteResponse, error) {
	if quote, ok := c.quotes[ticker]; ok {
		return quote, nil
	}

	return nil, fmt.Errorf("quote not found for %s", ticker)
}

func (c *MockClient) GetOrder(orderID string) (*alpaca.Order, error) {
	for _, order := range c.orders {
		if order.ID == orderID {
			return order, nil
		}
	}

	return nil, fmt.Errorf("no order found with ID %s", orderID)
}

func (c *MockClient) ListPositions() ([]alpaca.Position, error) {
	return c.positions, nil
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

func (c *MockClient) AddOrder(order *alpaca.Order) {
	c.orders = append(c.orders, order)
}

func (c *MockClient) GetOrderReqs() []alpaca.PlaceOrderRequest {
	return c.orderReqs
}

func (c *MockClient) GetOrders() []*alpaca.Order {
	return c.orders
}

func (c *MockClient) SetCash(cash decimal.Decimal) {
	c.cash = cash
}

func (c *MockClient) SetPositions(positions []alpaca.Position) {
	c.positions = positions
}

func NewFilledOrder(id string) *alpaca.Order {
	fillPrice := decimal.NewFromFloat(298.45)
	order := newOrder(id)

	now := time.Now()
	order.FilledAt = &now
	order.FilledQty = decimal.NewFromInt(3)
	order.FilledAvgPrice = &fillPrice
	order.Status = "filled"
	return order
}

func NewUnfilledOrder(id string) *alpaca.Order {
	order := newOrder(id)
	order.Status = "accepted"
	return order
}

func newOrder(id string) *alpaca.Order {
	now := time.Now()
	return &alpaca.Order{
		ID:            id,
		ClientOrderID: uuid.New().String(),
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now,
		SubmittedAt:   now.Add(-time.Hour),
		FilledAt:      &now,
		AssetID:       "4f75ad35-b947-4717-87db-19aa3dbf637d",
		Symbol:        "VOO",
		Exchange:      "Class:us_equity",
		Qty:           decimal.NewFromInt(3),
		Type:          alpaca.Market,
		Side:          alpaca.Buy,
		TimeInForce:   alpaca.Day,
		ExtendedHours: false,
	}
}

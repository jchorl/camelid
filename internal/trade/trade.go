package trade

import (
	"context"
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/golang/glog"
	"github.com/shopspring/decimal"

	"github.com/jchorl/camelid/internal/exchange"
	"github.com/jchorl/camelid/internal/reconciliation"
)

type Client struct {
	exchangeClient exchange.Client
	reconciler     reconciliation.Client
}

func New(exchangeClient exchange.Client, reconciler reconciliation.Client) *Client {
	return &Client{exchangeClient, reconciler}
}

func (c *Client) Buy(ctx context.Context, ticker string, dollarAmount decimal.Decimal) error {
	return c.trade(ctx, ticker, dollarAmount, alpaca.Buy)
}

func (c *Client) trade(ctx context.Context, ticker string, dollarAmount decimal.Decimal, side alpaca.Side) error {
	account, err := c.exchangeClient.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %w", err)
	}

	lastQuote, err := c.exchangeClient.GetLastQuote(ticker)
	if err != nil {
		return fmt.Errorf("GetLastQuote(%s): %w", ticker, err)
	}

	// depending on buy/sell, select for bid/ask
	var price decimal.Decimal
	if side == alpaca.Buy {
		price = decimal.NewFromFloat32(lastQuote.Last.BidPrice)
	} else {
		price = decimal.NewFromFloat32(lastQuote.Last.AskPrice)
	}

	qty := dollarAmount.Div(price).Floor()

	if qty.LessThan(decimal.NewFromInt(1)) {
		glog.Infof("not buying %s at $%s, $%s is too little to buy even 1 share", ticker, price.StringFixed(2), dollarAmount.StringFixed(2))
		return nil
	}

	record := reconciliation.NewRecord()

	err = c.reconciler.Record(ctx, record)
	if err != nil {
		return err
	}

	request := alpaca.PlaceOrderRequest{
		AccountID:     account.ID,
		AssetKey:      &ticker,
		Qty:           qty,
		Side:          side,
		Type:          alpaca.Market,
		TimeInForce:   alpaca.Day,
		ClientOrderID: record.GetID(),
	}

	glog.Infof("placing order %+v, estimated price: %v", request, price)

	order, err := c.exchangeClient.PlaceOrder(request)
	if err != nil {
		return fmt.Errorf("placing order %v: %w", request, err)
	}

	record.SetAccepted(order.ID)
	err = c.reconciler.Record(ctx, record)
	if err != nil {
		return err
	}

	glog.Infof("order completed: %+v", order)

	return nil
}

package trade

import (
	"context"
	"fmt"
	"math"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/golang/glog"
	"github.com/shopspring/decimal"

	"github.com/jchorl/camelid/internal/exchange"
)

func trade(ctx context.Context, ticker string, dollarAmount float32, side alpaca.Side) error {
	client := exchange.FromContext(ctx)
	account, err := client.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %w", err)
	}

	lastQuote, err := client.GetLastQuote(ticker)
	if err != nil {
		return fmt.Errorf("GetLastQuote(%s): %w", ticker, err)
	}

	// depending on buy/sell, select for bid/ask
	var price float32
	if side == alpaca.Buy {
		price = lastQuote.Last.BidPrice
	} else {
		price = lastQuote.Last.AskPrice
	}

	qty := math.Floor(float64(dollarAmount / price))

	request := alpaca.PlaceOrderRequest{
		AccountID:   account.ID,
		AssetKey:    &ticker,
		Qty:         decimal.NewFromFloat(qty),
		Side:        side,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	}

	glog.Infof("placing order %+v, estimated price: %v", request, price)

	order, err := client.PlaceOrder(request)
	if err != nil {
		return fmt.Errorf("placing order %v: %w", request, err)
	}

	glog.Infof("order completed: %+v", order)

	return nil
}
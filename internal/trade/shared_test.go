package trade

import (
	"context"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/jchorl/camelid/internal/exchange"
	"github.com/jchorl/camelid/internal/exchange/exchangetest"
)

func TestTrade(t *testing.T) {
	client := exchangetest.NewMockClient("6")
	ticker := "SPY"
	client.SetQuote(ticker, &alpaca.LastQuoteResponse{
		Status: "success",
		Symbol: "SPY",
		Last: alpaca.LastQuote{
			AskPrice:    326.41,
			AskSize:     5,
			AskExchange: 2,
			BidPrice:    326.35,
			BidSize:     1,
			BidExchange: 17,
			Timestamp:   1596226084553000000,
		},
	})
	ctx := exchange.NewContext(context.TODO(), client)
	err := trade(ctx, ticker, 3000.0, alpaca.Buy)
	require.NoError(t, err)
	require.ElementsMatch(t, []alpaca.PlaceOrderRequest{{
		AccountID:   "6",
		AssetKey:    &ticker,
		Qty:         decimal.NewFromInt(9),
		Side:        alpaca.Buy,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	}}, client.GetOrderReqs())
}

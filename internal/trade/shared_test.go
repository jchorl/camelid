package trade

import (
	"context"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/jchorl/camelid/internal/exchange"
	"github.com/jchorl/camelid/internal/exchange/exchangetest"
	"github.com/jchorl/camelid/internal/reconciliation"
)

func TestTrade(t *testing.T) {
	alpacaClient := exchangetest.NewMockClient("6")
	ticker := "SPY"
	alpacaClient.SetQuote(ticker, &alpaca.LastQuoteResponse{
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

	ctx := exchange.NewContext(context.TODO(), alpacaClient)
	reconciler := &mockReconciler{}
	ctx = reconciliation.NewContext(ctx, reconciler)

	err := trade(ctx, ticker, 3000.0, alpaca.Buy)
	require.NoError(t, err)

	require.Len(t, alpacaClient.GetOrderReqs(), 1)
	receivedReq := alpacaClient.GetOrderReqs()[0]
	require.Equal(t, "6", receivedReq.AccountID)
	require.Equal(t, &ticker, receivedReq.AssetKey)
	require.Equal(t, decimal.NewFromInt(9), receivedReq.Qty)
	require.Equal(t, alpaca.Buy, receivedReq.Side)
	require.Equal(t, alpaca.Market, receivedReq.Type)
	require.Equal(t, alpaca.Day, receivedReq.TimeInForce)

	require.Len(t, reconciler.records, 2)
	require.Equal(t, receivedReq.ClientOrderID, reconciler.records[0].GetID())

	// cant check status of the 0th record, because it gets updated
	require.Equal(t, receivedReq.ClientOrderID, reconciler.records[1].GetID())
	require.Equal(t, reconciliation.StatusAccepted, reconciler.records[1].GetStatus())
	require.Len(t, alpacaClient.GetOrders(), 1)
	require.Equal(t, alpacaClient.GetOrders()[0].ID, reconciler.records[1].GetAlpacaOrderID())
}

type mockReconciler struct {
	records []reconciliation.Record
}

func (r *mockReconciler) Record(_ context.Context, record reconciliation.Record) error {
	r.records = append(r.records, record)
	return nil
}

func (r *mockReconciler) Reconcile(_ context.Context) ([]reconciliation.Record, error) {
	return nil, nil
}

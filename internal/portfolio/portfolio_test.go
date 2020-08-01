package portfolio

import (
	"context"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/jchorl/camelid/internal/exchange/exchangetest"
)

func TestGetDeltas(t *testing.T) {
	cases := []struct {
		name             string
		currentPositions []alpaca.Position
		desiredRatios    map[string]float64
		amountToInvest   decimal.Decimal
		expectedDeltas   map[string]decimal.Decimal
	}{
		{
			name:             "none held",
			currentPositions: []alpaca.Position{},
			desiredRatios: map[string]float64{
				"SPY": 80,
				"VBD": 20,
			},
			amountToInvest: decimal.NewFromInt(1000),
			expectedDeltas: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(800),
				"VBD": decimal.NewFromInt(200),
			},
		},
		{
			name: "one held",
			currentPositions: []alpaca.Position{
				newPosition("VBD", decimal.NewFromInt(500)),
			},
			desiredRatios: map[string]float64{
				"SPY": 80,
				"VBD": 20,
			},
			amountToInvest: decimal.NewFromInt(10000),
			expectedDeltas: map[string]decimal.Decimal{ // brings totals to 2100/10500=0.2, 8400/10500=0.8
				"SPY": decimal.NewFromInt(8400),
				"VBD": decimal.NewFromInt(1600),
			},
		},
		{
			name: "not enough money",
			currentPositions: []alpaca.Position{
				newPosition("SPY", decimal.NewFromInt(1000)),
			},
			desiredRatios: map[string]float64{
				"SPY": 80,
				"VBD": 20,
			},
			amountToInvest: decimal.NewFromInt(100),
			expectedDeltas: map[string]decimal.Decimal{ // brings totals to 220/1100=0.2, 880/1100=0.8
				"SPY": decimal.NewFromInt(-120),
				"VBD": decimal.NewFromInt(220),
			},
		},
		{
			name: "unknown holding",
			currentPositions: []alpaca.Position{
				newPosition("VOO", decimal.NewFromInt(1000)),
				newPosition("SPY", decimal.NewFromInt(1000)),
				newPosition("VBD", decimal.NewFromInt(500)),
			},
			desiredRatios: map[string]float64{
				"SPY": 80,
				"VBD": 20,
			},
			amountToInvest: decimal.NewFromInt(500),
			expectedDeltas: map[string]decimal.Decimal{ // brings totals to 600/3000=0.2, 2400/3000=0.8
				"VOO": decimal.NewFromInt(-1000),
				"SPY": decimal.NewFromInt(1400),
				"VBD": decimal.NewFromInt(100),
			},
		},
		{
			name: "grab bag",
			currentPositions: []alpaca.Position{
				newPosition("VOO", decimal.NewFromInt(1000)),
				newPosition("SPY", decimal.NewFromInt(1000)),
				newPosition("VBD", decimal.NewFromInt(500)),
				newPosition("AAPL", decimal.NewFromInt(500)),
			},
			desiredRatios: map[string]float64{
				"SPY": 200,
				"VBD": 100,
				"VTI": 100,
			},
			amountToInvest: decimal.NewFromInt(2000),
			expectedDeltas: map[string]decimal.Decimal{
				"SPY":  decimal.NewFromInt(1500),  // (1500+1000)/5000=200/400
				"VBD":  decimal.NewFromInt(750),   // (750+500)/5000=100/400
				"AAPL": decimal.NewFromInt(-500),  // 0
				"VOO":  decimal.NewFromInt(-1000), // 0
				"VTI":  decimal.NewFromInt(1250),  // 1250/5000=100/400
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			alpacaClient := exchangetest.NewMockClient("6")
			alpacaClient.SetPositions(tc.currentPositions)

			portfolio := New(alpacaClient, tc.desiredRatios)
			deltas, err := portfolio.GetDeltas(context.TODO(), tc.amountToInvest)
			require.NoError(t, err)
			require.Equal(
				t, len(tc.expectedDeltas), len(deltas),
				"expected %d deltas, got %d. expected: %#v, actual: %#v",
				len(tc.expectedDeltas), len(deltas), tc.expectedDeltas, deltas,
			)
			for ticker, delta := range deltas {
				expectedDelta := tc.expectedDeltas[ticker]
				require.True(t, delta.Equal(expectedDelta), "[%s] expected delta: %s, actual delta: %s", ticker, expectedDelta, delta)
			}
		})
	}
}

func TestGetDeltas_ErrorsWithNoRatios(t *testing.T) {
	alpacaClient := exchangetest.NewMockClient("6")

	portfolio := New(alpacaClient, map[string]float64{})
	_, err := portfolio.GetDeltas(context.TODO(), decimal.NewFromInt(300))
	require.Error(t, err)
}

func newPosition(ticker string, marketValue decimal.Decimal) alpaca.Position {
	return alpaca.Position{
		AssetID:     "4f75ad35-b947-4717-87db-19aa3dbf637d",
		Symbol:      ticker,
		Exchange:    "ARCA",
		Class:       "us_equity",
		AccountID:   "",
		Side:        "long",
		MarketValue: marketValue,
	}
}

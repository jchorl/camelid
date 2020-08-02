package portfolio

import (
	"context"
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/jchorl/camelid/internal/exchange/exchangetest"
)

func TestGetDeltasWithoutSales(t *testing.T) {
	cases := []struct {
		name             string
		currentPositions []alpaca.Position
		desiredRatios    map[string]decimal.Decimal
		amountToInvest   decimal.Decimal
		expectedDeltas   map[string]decimal.Decimal
	}{
		{
			name:             "none held",
			currentPositions: []alpaca.Position{},
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
			},
			amountToInvest: decimal.NewFromInt(100),
			expectedDeltas: map[string]decimal.Decimal{
				"VBD": decimal.NewFromInt(100),
			},
		},
		{
			name: "unknown holding",
			currentPositions: []alpaca.Position{
				newPosition("VOO", decimal.NewFromInt(1000)),
				newPosition("SPY", decimal.NewFromInt(1000)),
				newPosition("VBD", decimal.NewFromInt(500)),
			},
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
			},
			amountToInvest: decimal.NewFromInt(500),
			// we cant achieve desired ratios.
			// ideally we'd have 2400 SPY and 600 VBD.
			// i.e. extra 1400 SPY and 100 VBD.
			// we buy proportional 500*1400/1500=466.66 and 500*100/1500=33.33
			expectedDeltas: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(400).Add(decimal.NewFromInt(200).Div(decimal.NewFromInt(3))), // 466.666
				"VBD": decimal.NewFromInt(100).Div(decimal.NewFromInt(3)),                              // 33.333
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(200),
				"VBD": decimal.NewFromInt(100),
				"VTI": decimal.NewFromInt(100),
			},
			amountToInvest: decimal.NewFromInt(2000),
			expectedDeltas: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(1500).Mul(decimal.NewFromInt(2000)).Div(decimal.NewFromInt(3500)), // 1500*2000/3500=857.14
				"VBD": decimal.NewFromInt(750).Mul(decimal.NewFromInt(2000)).Div(decimal.NewFromInt(3500)),  // 750*2000/3500=428.57
				"VTI": decimal.NewFromInt(1250).Mul(decimal.NewFromInt(2000)).Div(decimal.NewFromInt(3500)), // 1250*2000/3500=714.28
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			alpacaClient := exchangetest.NewMockClient("6")
			alpacaClient.SetPositions(tc.currentPositions)

			portfolio := New(alpacaClient, tc.desiredRatios)
			deltas, err := portfolio.GetDeltasWithoutSales(context.TODO(), tc.amountToInvest)
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

func TestGetDeltasWithSales(t *testing.T) {
	cases := []struct {
		name             string
		currentPositions []alpaca.Position
		desiredRatios    map[string]decimal.Decimal
		amountToInvest   decimal.Decimal
		expectedDeltas   map[string]decimal.Decimal
	}{
		{
			name:             "none held",
			currentPositions: []alpaca.Position{},
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(80),
				"VBD": decimal.NewFromInt(20),
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
			desiredRatios: map[string]decimal.Decimal{
				"SPY": decimal.NewFromInt(200),
				"VBD": decimal.NewFromInt(100),
				"VTI": decimal.NewFromInt(100),
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
			deltas, err := portfolio.GetDeltasWithSales(context.TODO(), tc.amountToInvest)
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

func TestGetDeltasWithSales_ErrorsWithNoRatios(t *testing.T) {
	alpacaClient := exchangetest.NewMockClient("6")

	portfolio := New(alpacaClient, map[string]decimal.Decimal{})
	_, err := portfolio.GetDeltasWithSales(context.TODO(), decimal.NewFromInt(300))
	require.Error(t, err)
}

func TestGetAmountToInvest(t *testing.T) {
	cases := []struct {
		name          string
		maxInvestment decimal.Decimal
		cash          decimal.Decimal
		expected      decimal.Decimal
	}{
		{
			name:          "max",
			maxInvestment: decimal.NewFromInt(1),
			cash:          decimal.NewFromInt(2),
			expected:      decimal.NewFromInt(1),
		},
		{
			name:          "not enough money",
			maxInvestment: decimal.NewFromInt(12),
			cash:          decimal.NewFromInt(2),
			expected:      decimal.NewFromInt(2),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			alpacaClient := exchangetest.NewMockClient("6")
			alpacaClient.SetCash(tc.cash)
			pfolio := New(alpacaClient, nil)
			toInvest, err := pfolio.GetAmountToInvest(tc.maxInvestment)
			require.NoError(t, err)
			require.True(t, toInvest.Equal(tc.expected), "expected %s to equal %s", tc.expected.String(), toInvest.String())
		})
	}
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

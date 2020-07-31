package trade

import (
	"context"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
)

func Buy(ctx context.Context, ticker string, dollarAmount float32) error {
	return trade(ctx, ticker, dollarAmount, alpaca.Buy)
}

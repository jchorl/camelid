package portfolio

import (
	"context"
	"errors"
	"fmt"

	"github.com/jchorl/camelid/internal/exchange"
	"github.com/shopspring/decimal"
)

type Portfolio struct {
	exchangeClient exchange.Client
	ratios         map[string]float64 `json:"ratios"` // ownership ratios, ticker -> shares
}

func New(exchangeClient exchange.Client, ratios map[string]float64) Portfolio {
	return Portfolio{
		exchangeClient: exchangeClient,
		ratios:         ratios,
	}
}

func (p *Portfolio) GetDeltas(ctx context.Context, amountToInvest decimal.Decimal) (map[string]decimal.Decimal, error) {
	if len(p.ratios) == 0 {
		return nil, errors.New("cannot get deltas with no holding ratios defined")
	}

	holdings, err := p.getCurrentHoldingsInDollars(ctx)
	if err != nil {
		return nil, err
	}

	total := decimal.Decimal{}
	for _, holding := range holdings {
		total = total.Add(holding)
	}

	totalShares := decimal.Decimal{}
	for _, shares := range p.ratios {
		totalShares = totalShares.Add(decimal.NewFromFloat(shares))
	}

	// divvy up the future pie
	futureTotal := total.Add(amountToInvest)
	desiredAmountDollars := map[string]decimal.Decimal{}
	for ticker, shares := range p.ratios {
		desiredAmountDollars[ticker] = decimal.NewFromFloat(shares).Div(totalShares).Mul(futureTotal)
	}

	deltas := map[string]decimal.Decimal{}
	for ticker, desiredDollars := range desiredAmountDollars {
		var holding decimal.Decimal
		if h, ok := holdings[ticker]; ok {
			holding = h
		}
		deltas[ticker] = desiredDollars.Sub(holding)
	}

	// sell any undesired holdings
	for ticker, holding := range holdings {
		// if any amount of this stock is desired, it was dealt with above
		if _, ok := desiredAmountDollars[ticker]; ok {
			continue
		}
		deltas[ticker] = holding.Neg()
	}
	return deltas, nil
}

func (p *Portfolio) getCurrentHoldingsInDollars(ctx context.Context) (map[string]decimal.Decimal, error) {
	positions, err := p.exchangeClient.ListPositions()
	if err != nil {
		return nil, fmt.Errorf("listing positions: %w", err)
	}

	holdings := map[string]decimal.Decimal{}
	for _, position := range positions {
		holdings[position.Symbol] = position.MarketValue
	}

	return holdings, nil
}

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
	ratios         map[string]decimal.Decimal // ownership ratios, ticker -> shares
}

func New(exchangeClient exchange.Client, ratios map[string]decimal.Decimal) Portfolio {
	return Portfolio{
		exchangeClient: exchangeClient,
		ratios:         ratios,
	}
}

func (p *Portfolio) GetDeltasWithoutSales(ctx context.Context, amountToInvest decimal.Decimal) (map[string]decimal.Decimal, error) {
	deltas, err := p.GetDeltasWithSales(ctx, amountToInvest)
	if err != nil {
		return nil, err
	}

	// filter out all sales
	filteredDeltas := map[string]decimal.Decimal{}
	for ticker, delta := range deltas {
		if !delta.IsPositive() {
			continue
		}

		filteredDeltas[ticker] = delta
	}

	totalDesiredSpend := sumMapValuesDecimal(filteredDeltas)

	// total spend can easily be > amountToInvest, because it accounts for sales.
	// it's the desired state, assuming you could reinvest every dollar today.
	// so scale down all buys to fit within budget.
	for ticker, delta := range filteredDeltas {
		filteredDeltas[ticker] = delta.Mul(amountToInvest).Div(totalDesiredSpend)
	}

	return filteredDeltas, nil
}

func (p *Portfolio) GetDeltasWithSales(ctx context.Context, amountToInvest decimal.Decimal) (map[string]decimal.Decimal, error) {
	if len(p.ratios) == 0 {
		return nil, errors.New("cannot get deltas with no holding ratios defined")
	}

	holdings, err := p.getCurrentHoldingsInDollars(ctx)
	if err != nil {
		return nil, err
	}

	total := sumMapValuesDecimal(holdings)
	total = total.Add(amountToInvest)

	totalShares := sumMapValuesDecimal(p.ratios)

	// divvy up the future pie
	desiredAmountDollars := map[string]decimal.Decimal{}
	for ticker, shares := range p.ratios {
		desiredAmountDollars[ticker] = shares.Div(totalShares).Mul(total)
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

func (p *Portfolio) GetAmountToInvest(maxAmount decimal.Decimal) (decimal.Decimal, error) {
	acct, err := p.exchangeClient.GetAccount()
	if err != nil {
		return decimal.Decimal{}, err
	}

	return decimal.Min(maxAmount, acct.Cash), nil
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

func sumMapValuesDecimal(m map[string]decimal.Decimal) decimal.Decimal {
	sum := decimal.Zero
	for _, v := range m {
		sum = sum.Add(v)
	}
	return sum
}

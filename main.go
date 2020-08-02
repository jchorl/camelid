package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/glog"
	"github.com/shopspring/decimal"

	"github.com/jchorl/camelid/internal/portfolio"
	"github.com/jchorl/camelid/internal/reconciliation"
	"github.com/jchorl/camelid/internal/trade"
)

func main() {
	flag.Parse()

	dryRun := true

	// lambda only supports env vars, not CLI flags...
	ratios := map[string]float64{}
	ratiosJSON := os.Getenv("CAMELID_RATIOS")
	err := json.Unmarshal([]byte(ratiosJSON), &ratios)
	if err != nil {
		glog.Fatalf("parsing config: %v", err)
	}
	ratiosDecimal := map[string]decimal.Decimal{}
	for ticker, shares := range ratios {
		ratiosDecimal[ticker] = decimal.NewFromFloat(shares)
	}

	maxInvestment, err := decimal.NewFromString(os.Getenv("CAMELID_MAX_INVESTMENT"))
	if err != nil {
		glog.Fatalf("parsing max investment: %v", err)
	}

	err = run(context.TODO(), dryRun, ratiosDecimal, maxInvestment)
	if err != nil {
		glog.Fatalf("failed: %v", err)
	}
}

func run(ctx context.Context, dryRun bool, ratios map[string]decimal.Decimal, maxInvestment decimal.Decimal) error {
	alpacaClient := alpaca.NewClient(common.Credentials())

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynamoClient := dynamodb.New(awsSession)

	reconciler := reconciliation.New(dynamoClient, alpacaClient)
	tradingClient := trade.New(alpacaClient, reconciler)

	err := reconciler.Reconcile(ctx)
	if err != nil {
		glog.Fatalf("failed to reconcile: %v", err)
	}

	pfolio := portfolio.New(alpacaClient, ratios)
	amountToInvest, err := pfolio.GetAmountToInvest(maxInvestment)
	if err != nil {
		glog.Fatalf("getting amount to invest: %v", err)
	}

	deltas, err := pfolio.GetDeltasWithoutSales(ctx, amountToInvest)
	if err != nil {
		glog.Fatalf("getting deltas: %v", err)
	}

	// TODO divvy up the buy pie, because we cant use dollars from sales
	for ticker, delta := range deltas {
		if dryRun {
			glog.Infof("DRY-RUN would have traded $%s of %s", delta.StringFixed(2), ticker)
		} else if delta.GreaterThan(decimal.Zero) {
			err := tradingClient.Buy(ctx, ticker, delta)
			if err != nil {
				return err
			}
		} else {
			glog.Warningf("selling is not supported yet, not selling $%s of %s", delta.Abs().StringFixed(2), ticker)
		}
	}
	return nil
}

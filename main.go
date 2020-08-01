package main

import (
	"context"
	"flag"

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

	err := run(context.TODO(), dryRun)
	if err != nil {
		glog.Fatalf("failed: %v", err)
	}
}

func run(ctx context.Context, dryRun bool) error {
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

	pfolio := portfolio.New(alpacaClient, map[string]float64{})    // TODO get ratios from config
	deltas, err := pfolio.GetDeltas(ctx, decimal.NewFromInt(1000)) // TODO get the amount to invest properly
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

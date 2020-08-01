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

	"github.com/jchorl/camelid/internal/db"
	"github.com/jchorl/camelid/internal/exchange"
	"github.com/jchorl/camelid/internal/portfolio"
	"github.com/jchorl/camelid/internal/reconciliation"
)

func main() {
	flag.Parse()

	ctx := context.TODO()
	alpacaClient := alpaca.NewClient(common.Credentials())
	ctx = exchange.NewContext(ctx, alpacaClient)

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynamoClient := dynamodb.New(awsSession)
	ctx = db.NewContext(ctx, dynamoClient)

	reconciler := reconciliation.NewReconciler()
	ctx = reconciliation.NewContext(ctx, reconciler)

	err := run(ctx)
	if err != nil {
		glog.Fatalf("failed: %v", err)
	}
}

func run(ctx context.Context) error {
	reconciler := reconciliation.FromContext(ctx)
	err := reconciler.Reconcile(ctx)
	if err != nil {
		glog.Fatalf("failed to reconcile: %v", err)
	}

	pfolio := portfolio.New(map[string]float64{})
	deltas, err := pfolio.GetDeltas(ctx, decimal.NewFromInt(1000)) // TODO get the amount to invest properly
	if err != nil {
		glog.Fatalf("getting deltas: %v", err)
	}

	glog.Infof("found deltas: %v", deltas)
	return nil
}

package main

import (
	"context"
	"flag"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/glog"

	"github.com/jchorl/camelid/internal/db"
	"github.com/jchorl/camelid/internal/exchange"
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

	// pfolio := portfolio.New(portfolio.Config{Ratios: map[string]float64{}})

	// err := reconciler.Reconcile(ctx)
	// if err != nil {
	// 	glog.Fatalf("failed to reconcile: %v", err)
	// }

	positions, err := alpacaClient.ListPositions()
	if err != nil {
		glog.Fatalf("listing positions: %v", err)
	}
	glog.Infof("%#v", positions)

	// _, err := pfolio.GetDeltas(ctx)
	// if err != nil {
	// 	glog.Fatalf("getting deltas: %v", err)
	// }
}

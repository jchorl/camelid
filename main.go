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

	var err error
	// err = trade.Buy(ctx, "SPY", 500.0)
	// if err != nil {
	// 	glog.Fatalf("buying: %+v", err)
	// }

	err = reconciler.Reconcile(ctx)
	if err != nil {
		glog.Fatalf("reconciling: %+v", err)
	}
}

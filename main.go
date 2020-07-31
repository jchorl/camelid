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
)

func main() {
	flag.Parse()

	ctx := context.TODO()
	alpacaClient := alpaca.NewClient(common.Credentials())
	ctx = exchange.NewContext(ctx, alpacaClient)

	awsSession := session.Must(session.NewSession())
	dynamoClient := dynamodb.New(awsSession)
	ctx = db.NewContext(ctx, dynamoClient)

	q, _ := alpacaClient.GetLastQuote("SPY")
	glog.Infof("%+v", q)
}

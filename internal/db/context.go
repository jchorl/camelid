package db

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ctxKey int

var dynamoClientKey ctxKey

func NewContext(ctx context.Context, client *dynamodb.DynamoDB) context.Context {
	return context.WithValue(ctx, dynamoClientKey, client)
}

func FromContext(ctx context.Context) *dynamodb.DynamoDB {
	return ctx.Value(dynamoClientKey).(*dynamodb.DynamoDB)
}

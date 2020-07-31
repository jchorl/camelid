package db

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type ctxKey int

var dynamoClientKey ctxKey

func NewContext(ctx context.Context, client dynamodbiface.DynamoDBAPI) context.Context {
	return context.WithValue(ctx, dynamoClientKey, client)
}

func FromContext(ctx context.Context) dynamodbiface.DynamoDBAPI {
	return ctx.Value(dynamoClientKey).(dynamodbiface.DynamoDBAPI)
}

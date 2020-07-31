package exchange

import (
	"context"
)

type ctxKey int

var alpacaClientKey ctxKey

func NewContext(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, alpacaClientKey, client)
}

func FromContext(ctx context.Context) Client {
	return ctx.Value(alpacaClientKey).(Client)
}

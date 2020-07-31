package reconciliation

import (
	"context"
)

type ctxKey int

var reconcilerClientKey ctxKey

func NewContext(ctx context.Context, reconciler Reconciler) context.Context {
	return context.WithValue(ctx, reconcilerClientKey, reconciler)
}

func FromContext(ctx context.Context) Reconciler {
	return ctx.Value(reconcilerClientKey).(Reconciler)
}

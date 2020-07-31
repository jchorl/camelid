package reconciliation

import (
	"context"
	"testing"

	"github.com/jchorl/camelid/internal/db"
	"github.com/jchorl/camelid/internal/db/dbtest"
	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	dynamoClient := dbtest.NewMockClient(dynamoTable)
	ctx := db.NewContext(context.TODO(), dynamoClient)

	reconciler := NewReconciler()
	rec := &record{
		ID:            "123",
		AlpacaOrderID: "alpaca_111",
		Status:        StatusAccepted,
	}
	err := reconciler.Record(ctx, rec)
	require.NoError(t, err)
}

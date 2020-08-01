package reconciliation

import (
	"context"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/jchorl/camelid/internal/db"
	"github.com/jchorl/camelid/internal/db/dbtest"
	"github.com/jchorl/camelid/internal/exchange"
	"github.com/jchorl/camelid/internal/exchange/exchangetest"
	"github.com/stretchr/testify/require"
)

func TestRecord(t *testing.T) {
	dynamoClient := dbtest.NewMockClient(dynamoTable)
	ctx := db.NewContext(context.TODO(), dynamoClient)

	reconciler := NewReconciler()
	rec := &record{
		ID:            "123",
		AlpacaOrderID: "alpaca_111",
		Status:        StatusUnreconciled,
	}
	err := reconciler.Record(ctx, rec)
	require.NoError(t, err)
}

func TestReconcile(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name        string
		orders      []*alpaca.Order
		dbRecords   []record
		expectedErr bool
	}{
		{
			name:        "no data",
			expectedErr: false,
		},
		{
			name: "only reconciled",
			dbRecords: []record{
				{
					ID:            "trade1",
					AlpacaOrderID: "alpaca11",
					Status:        StatusReconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
					ReconciledAt:  &now,
				},
			},
			expectedErr: false,
		},
		{
			name: "unfound order",
			dbRecords: []record{
				{
					ID:            "trade1",
					AlpacaOrderID: "alpaca11",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
			},
			expectedErr: true,
		},
		{
			name:   "unreconciled and unfilled",
			orders: []*alpaca.Order{exchangetest.NewUnfilledOrder("alpaca11")},
			dbRecords: []record{
				{
					ID:            "trade1",
					AlpacaOrderID: "alpaca11",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
			},
			expectedErr: true,
		},
		{
			name:   "unreconciled and filled",
			orders: []*alpaca.Order{exchangetest.NewFilledOrder("alpaca11")},
			dbRecords: []record{
				{
					ID:            "trade1",
					AlpacaOrderID: "alpaca11",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
			},
			expectedErr: false,
		},
		{
			name: "grab bag, unreconciled",
			orders: []*alpaca.Order{
				exchangetest.NewFilledOrder("alpaca11"),
				exchangetest.NewFilledOrder("alpaca12"),
				exchangetest.NewUnfilledOrder("alpaca13"),
			},
			dbRecords: []record{
				{
					ID:            "trade1",
					AlpacaOrderID: "alpaca11",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
				{
					ID:            "trade2",
					AlpacaOrderID: "alpaca12",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
				{
					ID:            "trade3",
					AlpacaOrderID: "alpaca13",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
			},
			expectedErr: true,
		},
		{
			name: "grab bag, reconciled",
			orders: []*alpaca.Order{
				exchangetest.NewFilledOrder("alpaca11"),
				exchangetest.NewFilledOrder("alpaca12"),
				exchangetest.NewFilledOrder("alpaca13"),
			},
			dbRecords: []record{
				{
					ID:            "trade1",
					AlpacaOrderID: "alpaca11",
					Status:        StatusReconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
				{
					ID:            "trade2",
					AlpacaOrderID: "alpaca12",
					Status:        StatusReconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
				{
					ID:            "trade3",
					AlpacaOrderID: "alpaca13",
					Status:        StatusUnreconciled,
					CreatedAt:     now,
					SubmittedAt:   &now,
				},
			},
			expectedErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			reconciler := NewReconciler()

			alpacaClient := exchangetest.NewMockClient("6")
			ctx := exchange.NewContext(context.TODO(), alpacaClient)
			for _, order := range tc.orders {
				alpacaClient.AddOrder(order)
			}

			dbClient := dbtest.NewMockClient(dynamoTable)
			ctx = db.NewContext(ctx, dbClient)
			for _, rec := range tc.dbRecords {
				reconciler.Record(ctx, &rec)
			}

			err := reconciler.Reconcile(ctx)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

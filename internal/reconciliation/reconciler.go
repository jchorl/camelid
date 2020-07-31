package reconciliation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/jchorl/camelid/internal/db"
)

type Reconciler interface {
	Record(context.Context, TradeRecord) error
	Reconcile(context.Context) ([]TradeRecord, error)
}

const dynamoTable = "TradeRecordsTest"

type reconciler struct{}

func NewReconciler() Reconciler {
	return &reconciler{}
}

func (r *reconciler) Record(ctx context.Context, record TradeRecord) error {
	dynamoClient := db.FromContext(ctx)
	av, err := dynamodbattribute.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("marshaling record (%+v): %w", r, err)
	}

	_, err = dynamoClient.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(dynamoTable),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
	}

	return nil
}

func (r *reconciler) Reconcile(ctx context.Context) ([]TradeRecord, error) {
	// query all unreconciled
	// loop through and check status
	return nil, nil
}

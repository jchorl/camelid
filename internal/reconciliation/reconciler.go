package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/jchorl/camelid/internal/exchange"
)

type Client interface {
	Record(context.Context, Record) error
	Reconcile(context.Context) error
}

const dynamoTable = "TradeRecordsTest"

type client struct {
	db             dynamodbiface.DynamoDBAPI
	exchangeClient exchange.Client
}

func New(db dynamodbiface.DynamoDBAPI, exchangeClient exchange.Client) Client {
	return &client{db, exchangeClient}
}

func (c *client) Record(ctx context.Context, rec Record) error {
	av, err := dynamodbattribute.MarshalMap(rec)
	if err != nil {
		return fmt.Errorf("marshaling record (%+v): %w", rec, err)
	}

	_, err = c.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(dynamoTable),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("put item: %w", err)
	}

	return nil
}

func (c *client) Reconcile(ctx context.Context) error {
	// query all unreconciled
	unreconciled, err := c.getUnreconciled(ctx)
	if err != nil {
		return err
	} else if len(unreconciled) == 0 {
		return nil
	}

	// loop through and check status
	var stillUnreconciledIDs []string
	for _, rec := range unreconciled {
		order, err := c.exchangeClient.GetOrder(rec.AlpacaOrderID)
		if err != nil {
			return fmt.Errorf("getting order from alpaca (%s): %w", rec.AlpacaOrderID, err)
		}

		if isTerminalState(order.Status) {
			err := c.setReconciled(ctx, rec.ID, order.UpdatedAt)
			if err != nil {
				return err
			}
			continue
		}

		stillUnreconciledIDs = append(stillUnreconciledIDs, rec.AlpacaOrderID)
	}

	if len(stillUnreconciledIDs) > 0 {
		return fmt.Errorf("some orders could not be reconciled, alpaca IDs: %v", stillUnreconciledIDs)
	}

	return nil
}

func isTerminalState(status string) bool {
	terminalStates := []string{"filled", "canceled", "expired", "rejected"}
	for _, state := range terminalStates {
		if status == state {
			return true
		}
	}

	return false
}

func (c *client) setReconciled(ctx context.Context, id string, reconciledAt time.Time) error {
	resp, err := c.db.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(dynamoTable),
	})
	if err != nil {
		return fmt.Errorf("GetItem(%s) from dynamo: %w", id, err)
	}

	rec := record{}
	err = dynamodbattribute.UnmarshalMap(resp.Item, &rec)
	if err != nil {
		return fmt.Errorf("unmarshaling item from dynamo: %w", err)
	}

	rec.ReconciledAt = &reconciledAt
	rec.Status = StatusReconciled
	err = c.Record(ctx, &rec)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) getUnreconciled(ctx context.Context) ([]record, error) {
	filt := expression.Name("Status").NotEqual(expression.Value(StatusReconciled))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("building unreconciled query: %w", err)
	}

	var records []record

	var unmarshalErr error
	err = c.db.ScanPagesWithContext(ctx, &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(dynamoTable),
		IndexName:                 aws.String("StatusIndex"),
	}, func(page *dynamodb.ScanOutput, last bool) bool {
		recs := []record{}

		err := dynamodbattribute.UnmarshalListOfMaps(page.Items, &recs)
		if err != nil {
			unmarshalErr = err
			return false
		}

		records = append(records, recs...)

		return true // keep paging
	})
	if err != nil {
		return nil, fmt.Errorf("ScanPages: %w", err)
	} else if unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshaling dynamo results: %w", err)
	}

	return records, nil
}

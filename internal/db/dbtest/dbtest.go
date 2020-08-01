package dbtest

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/golang/glog"
)

var _ dynamodbiface.DynamoDBAPI = (*MockClient)(nil)

type dynamoItem map[string]*dynamodb.AttributeValue

func (i dynamoItem) getIntVal(key string) int {
	intVal, err := strconv.Atoi(aws.StringValue(i[key].N))
	if err != nil {
		glog.Fatalf("failed to convert key %s to int, AttributeValue: %+v: %v", key, i[key], err)
	}
	return intVal
}

type MockClient struct {
	dynamodbiface.DynamoDBAPI

	items map[string]map[string]dynamoItem // table_name -> key -> item
}

func NewMockClient(tableNames ...string) *MockClient {
	items := map[string]map[string]dynamoItem{}
	for _, name := range tableNames {
		items[name] = map[string]dynamoItem{}
	}

	return &MockClient{
		items: items,
	}
}

func (c *MockClient) GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error) {
	table, ok := c.items[aws.StringValue(input.TableName)]
	if !ok {
		return nil, fmt.Errorf("table %s not found", aws.StringValue(input.TableName))
	}

	item, ok := table[aws.StringValue(input.Key["ID"].S)]
	if !ok {
		return nil, fmt.Errorf("item %s, table %s not found", aws.StringValue(input.Key["ID"].S), aws.StringValue(input.TableName))
	}

	return &dynamodb.GetItemOutput{Item: item}, nil
}

func (c *MockClient) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	c.items[aws.StringValue(input.TableName)][aws.StringValue(input.Item["ID"].S)] = input.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (c *MockClient) ScanPagesWithContext(ctx aws.Context, input *dynamodb.ScanInput, fn func(*dynamodb.ScanOutput, bool) bool, opts ...request.Option) error {
	// instead of reimplementing the dynamo API
	// just check for a specific query
	if (aws.StringValue(input.ExpressionAttributeNames["#0"]) == "Status") &&
		(aws.StringValue(input.ExpressionAttributeValues[":0"].N) == "1") &&
		(aws.StringValue(input.FilterExpression) == "#0 <> :0") &&
		(aws.StringValue(input.IndexName) == "StatusIndex") {
		for _, record := range c.items[aws.StringValue(input.TableName)] {
			if record.getIntVal("Status") != 1 {
				if !fn(
					&dynamodb.ScanOutput{Items: []map[string]*dynamodb.AttributeValue{record}},
					false, // who even uses this param
				) {
					break
				}
			}
		}
		return nil
	}

	return fmt.Errorf("please implement this query in MockClient.ScanPagesWithContext")
}

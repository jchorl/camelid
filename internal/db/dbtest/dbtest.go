package dbtest

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var _ dynamodbiface.DynamoDBAPI = (*MockClient)(nil)

type MockClient struct {
	dynamodbiface.DynamoDBAPI

	items map[string]map[string]interface{} // table_name -> key -> item
}

func NewMockClient(tableNames ...string) *MockClient {
	items := map[string]map[string]interface{}{}
	for _, name := range tableNames {
		items[name] = map[string]interface{}{}
	}

	return &MockClient{
		items: items,
	}
}

func (c *MockClient) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	c.items[aws.StringValue(input.TableName)][aws.StringValue(input.Item["ID"].S)] = input.Item
	return &dynamodb.PutItemOutput{}, nil
}

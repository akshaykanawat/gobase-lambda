package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"gobase-lambda/log"
)

type DynamoDb struct {
	_      struct{}
	Client *dynamodb.DynamoDB
	log    *log.Log
	ctx    context.Context
}

var defaultDynamoDbClient *dynamodb.DynamoDB

func GetAWSDynamoDbClient(awsSession *session.Session) *dynamodb.DynamoDB {
	return dynamodb.New(awsSession)
}

func GetDefaultDynamoDbClient(ctx context.Context) *DynamoDb {
	if defaultDynamoDbClient == nil {
		defaultDynamoDbClient = GetAWSDynamoDbClient(defaultAWSSession)
	}
	return GetDynamoDbClient(ctx, defaultDynamoDbClient)
}

func GetDynamoDbClient(ctx context.Context, client *dynamodb.DynamoDB) *DynamoDb {
	return &DynamoDb{Client: client, log: log.GetDefaultLogger(), ctx: ctx}
}

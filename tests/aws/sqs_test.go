package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gobase-lambda/aws"
)

func TestSQSClient(t *testing.T) {
	queueUrl, err := aws.GetQueueURL("DIGIO_WEBHOOK", aws.GetAWSSQSClient(aws.GetDefaultAWSSession()), context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	sqsClient := aws.GetDefaultSQSClient(context.TODO(), *queueUrl)
	err = sqsClient.SendMessage(map[string]string{
		"hello": "Fasdfasdf-" + uuid.New().String(),
	}, map[string]string{
		"id": uuid.NewString(),
	}, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err := sqsClient.ReceiveMessage(10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = sqsClient.DeleteMessage(messageList[0].ReceiptHandle)
	if err != nil {
		t.Fatal(err)
	}
	sqsMessageList := make([]*aws.BatchQueueMessage, 10)

	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		sqsMessageList[i] = &aws.BatchQueueMessage{
			Id: &id,
			Message: map[string]string{
				"hello": "Fasdfasdf-" + uuid.New().String(),
			},
			Attribute: map[string]string{
				"id": uuid.NewString(),
			},
		}
	}
	_, err = sqsClient.SendMessageBatch(sqsMessageList, 1)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err = sqsClient.ReceiveMessage(10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	deleteMap := make(map[string]*string, len(messageList))
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatch(deleteMap)
	if err != nil {
		t.Fatal(err)
	}

}

func TestSQSFIFOClient(t *testing.T) {
	queueUrl, err := aws.GetQueueURL("DIGIO_WEBHOOK.fifo", aws.GetAWSSQSClient(aws.GetDefaultAWSSession()), context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	sqsClient := aws.GetDefaultSQSClient(context.TODO(), *queueUrl)
	groupId, dedupId := uuid.NewString(), uuid.NewString()
	err = sqsClient.SendMessage(map[string]string{
		"hello": "Fasdfasdf-" + uuid.New().String(),
	}, map[string]string{
		"id": uuid.NewString(),
	}, 0, &groupId, &dedupId)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err := sqsClient.ReceiveMessage(10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	err = sqsClient.DeleteMessage(messageList[0].ReceiptHandle)
	if err != nil {
		t.Fatal(err)
	}
	sqsMessageList := make([]*aws.BatchQueueMessage, 10)
	groupId = "data"
	for i := 0; i < 10; i++ {
		id := uuid.NewString()
		sqsMessageList[i] = &aws.BatchQueueMessage{
			Id: &id,
			Message: map[string]string{
				"hello": "Fasdfasdf-" + uuid.New().String(),
			},
			Attribute: map[string]string{
				"id": uuid.NewString(),
			},
			MessageDeduplicationId: &id,
			MessageGroupId:         &groupId,
		}
	}
	_, err = sqsClient.SendMessageBatch(sqsMessageList, 0)
	if err != nil {
		t.Fatal(err)
	}
	messageList, err = sqsClient.ReceiveMessage(10, 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	deleteMap := make(map[string]*string, len(messageList))
	for _, m := range messageList {
		id := uuid.NewString()
		deleteMap[id] = m.ReceiptHandle
	}
	_, err = sqsClient.DeleteMessageBatch(deleteMap)
	if err != nil {
		t.Fatal(err)
	}

}

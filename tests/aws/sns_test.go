package tests

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	"gobase-lambda/aws"
)

func TestSNSClient(t *testing.T) {
	//os.Setenv("snsTopicPrefix", "MFCORE")
	// arn, err := aws.GetSNSARN("SYSTEM_ALERTS")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	arn := "arn:aws:sns:ap-south-1:722285188889:staging_BEDROCK_SYSTEM_ALERTS"
	snsClient := aws.GetDefaultSNSClient(context.TODO())
	data := map[string]map[string]interface{}{
		"payment": {
			"id":     "pay_14341234",
			"amount": 123,
		},
		"bank": {
			"id":                "bank_fadsfas",
			"bankAccountNumber": "0000021312",
		},
		"customer": {
			"id": "cust_fasdfsa",
		},
	}
	message := snsClient.GetSNSDataTemplate("payment.success", data, "bank", "customer")
	err := snsClient.Publish(&arn, nil, message, nil)
	if err != nil {
		t.Fatal(err)
	}
	data1 := "hello"
	req := &sns.PublishInput{
		TopicArn: &arn,
		Subject:  &data1,
		Message:  &data1,
	}
	_, _ = snsClient.Client.PublishWithContext(context.TODO(), req)
}

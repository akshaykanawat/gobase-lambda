package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"gobase-lambda/eventprocessor"
	"gobase-lambda/example"
)

type SNSEvent events.SNSEvent

func TestSNSManagerTransactionRejected(t *testing.T) {
	var SNSPayl SNSEvent
	handler := eventprocessor.GetHandler(false, example.NewManager)
	payload_json, err := ioutil.ReadFile("samples/sample_sns_transaction_rejected.json")
	if err != nil {
		fmt.Print(err)
		return
	}
	json.Unmarshal(payload_json, &SNSPayl)
	request := events.SNSEvent(SNSPayl)
	response, err := handler.HandleSNSRequest(context.TODO(), request)
	fmt.Println(response)
	fmt.Println(err)
}

func TestSNSManagerTransactionConfirmed(t *testing.T) {
	var SNSPayl SNSEvent
	handler := eventprocessor.GetHandler(false, example.NewManager)
	payload_json, err := ioutil.ReadFile("samples/sample_sns_transaction_accepted.json")
	if err != nil {
		fmt.Print(err)
		return
	}
	json.Unmarshal(payload_json, &SNSPayl)
	request := events.SNSEvent(SNSPayl)
	response, err := handler.HandleSNSRequest(context.TODO(), request)
	fmt.Println(response)
	fmt.Println(err)
}

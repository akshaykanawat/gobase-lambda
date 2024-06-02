package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"gobase-lambda/eventprocessor"
	"gobase-lambda/example"
)

func TestAPIManagerGET(t *testing.T) {
	handler := eventprocessor.GetHandler(false, example.NewManager)
	request := events.APIGatewayProxyRequest{
		Resource:   "/{customerId}",
		HTTPMethod: "GET",
		QueryStringParameters: map[string]string{
			"name":    "hello",
			"status":  "test",
			"numList": "1",
			"flag":    "true",
		},
		MultiValueQueryStringParameters: map[string][]string{
			"name":    {"hello"},
			"status":  {"test", "test2"},
			"numList": {"1", "2"},
			"flag":    {"true"},
		},
		PathParameters: map[string]string{
			"customerId": "cust_fasdfafsdf",
		},
	}
	response, err := handler.HandleAPIRequest(context.TODO(), request)
	fmt.Printf("%+v\n", response)
	fmt.Println(err)
}

func TestAPIManagerPOST(t *testing.T) {
	handler := eventprocessor.GetHandler(false, example.NewManager)
	body := map[string]interface{}{
		"name":   "Sabariram",
		"gender": "M",
		"address": map[string]string{
			"addressLine1": "7/151 Jodukuli",
			"addressLine2": "Kadyampatti TK",
			"addressLine3": "Salem",
		},
	}
	bodyBlob, _ := json.Marshal(body)
	request := events.APIGatewayProxyRequest{
		Resource:   "/{customerId}",
		HTTPMethod: "POST",
		Body:       string(bodyBlob),
		PathParameters: map[string]string{
			"customerId": "cust_fasdfafsdf",
		},
	}
	response, err := handler.HandleAPIRequest(context.TODO(), request)
	fmt.Printf("%+v\n", response)
	fmt.Println(err)
}

func TestAPIUpload(t *testing.T) {
	handler := eventprocessor.GetHandler(false, example.NewManager)
	body := map[string]interface{}{
		"url": uploadToS3(t),
	}
	bodyBlob, _ := json.Marshal(body)
	request := events.APIGatewayProxyRequest{
		Resource:   "/upload",
		HTTPMethod: "POST",
		Body:       string(bodyBlob),
		PathParameters: map[string]string{
			"customerId": "cust_fasdfafsdf",
		},
	}
	_, err := handler.HandleAPIRequest(context.TODO(), request)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAPIDocumentPII(t *testing.T) {
	handler := eventprocessor.GetHandler(false, example.NewManager)
	body := map[string]interface{}{
		"url": uploadToS3(t),
	}
	bodyBlob, _ := json.Marshal(body)
	request := events.APIGatewayProxyRequest{
		Resource:   "/pii/upload",
		HTTPMethod: "POST",
		Body:       string(bodyBlob),
		PathParameters: map[string]string{
			"customerId": "cust_fasdfafsdf",
		},
	}
	response, err := handler.HandleAPIRequest(context.TODO(), request)
	if err != nil {
		t.Fatal(err)
	}
	resBody := make(map[string]string)
	loadBody(t, &response, &resBody)
	body = map[string]interface{}{
		"FileName": resBody["uploadPath"],
	}
	bodyBlob, _ = json.Marshal(body)
	request = events.APIGatewayProxyRequest{
		Resource:   "/pii/download",
		HTTPMethod: "POST",
		Body:       string(bodyBlob),
	}
	response, err = handler.HandleAPIRequest(context.TODO(), request)
	if err != nil {
		t.Fatal(err)
	}
}

func TestApiSFNInvocation(t *testing.T) {
	handler := eventprocessor.GetHandler(true, example.NewManager)
	request := events.APIGatewayProxyRequest{
		Resource:              "/step-func",
		HTTPMethod:            "GET",
		QueryStringParameters: nil,
		PathParameters:        nil,
	}
	handler.HandleAPIRequest(context.TODO(), request)
}

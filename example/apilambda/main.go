package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"gobase-lambda/eventprocessor"
	"gobase-lambda/example"
	"gobase-lambda/log"
)

func main() {
	handler := eventprocessor.GetHandler(true, example.NewManager)
	logger := log.GetDefaultLogger()
	logger.Info("Lambda initiated", nil)
	lambda.Start(handler.HandleAPIRequest)
}

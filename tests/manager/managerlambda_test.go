package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"gobase-lambda/eventprocessor"
	"gobase-lambda/example"
)

func TestLambdaManagerActionOne(t *testing.T) {
	var LambdaEvent eventprocessor.LambdaEvent
	handler := eventprocessor.GetHandler(false, example.NewManager)
	jsonStr := "{\"isLambdaInvocation\":true,\"actionName\":\"ACTION_1\",\"payload\":{\"hello\":\"world\"}}"
	json.Unmarshal([]byte(jsonStr), &LambdaEvent)
	response, err := handler.HandleLambdaInvocation(context.TODO(), LambdaEvent)
	fmt.Println(response)
	fmt.Println(err)
}

func TestLambdaManagerActionTwo(t *testing.T) {
	// var LambdaEvent eventprocessor.LambdaEvent
	handler := eventprocessor.GetHandler(false, example.NewManager)
	jsonStr := "{\"isLambdaInvocation\":true,\"actionName\":\"ACTION_2\",\"payload\":{\"hello\":\"worlssd\"}}"
	// json.Unmarshal([]byte(jsonStr), &LambdaEvent)
	response, err := handler.HandleLambdaInvocation(context.TODO(), jsonStr)
	fmt.Println(response)
	fmt.Println(err)
}

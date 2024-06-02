package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"gobase-lambda/eventprocessor"
	"gobase-lambda/example"
)

func TestCronManagerActionOne(t *testing.T) {
	var LambdaEvent eventprocessor.CronEvent
	handler := eventprocessor.GetHandler(false, example.NewManager)
	jsonStr := "{\"isCrons\":true,\"actionNames\":\"CRON_ACTION_1\",\"payload\":{\"hello\":\"worldss\"}}"
	json.Unmarshal([]byte(jsonStr), &LambdaEvent)
	response, err := handler.HandleCronInvocation(context.TODO(), LambdaEvent)
	fmt.Println(response)
	fmt.Println(err)
}

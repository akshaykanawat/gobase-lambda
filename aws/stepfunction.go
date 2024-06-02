package aws

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"gobase-lambda/log"
)

type StepFunction struct {
	Client *sfn.SFN
	log    *log.Log
	ctx    context.Context
}

var defaultSFNClient *sfn.SFN

func GetDefaultSFNClient(ctx context.Context) *StepFunction {
	if defaultSFNClient == nil {
		defaultSFNClient = GetAWSSFNClient(defaultAWSSession)
	}
	return GetSFNClient(ctx, defaultSFNClient)
}

func GetAWSSFNClient(awsSession *session.Session) *sfn.SFN {
	client := sfn.New(awsSession)
	return client
}

func GetSFNClient(ctx context.Context, sfnClient *sfn.SFN) *StepFunction {
	return &StepFunction{Client: sfnClient, log: log.GetDefaultLogger(), ctx: ctx}
}

func (s *StepFunction) StartExecution(stateMachineArn, executionName string, payload interface{}) (err error) {
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		s.log.Error("State machine payload marshal error", err)
		return
	}
	stringifiedMarshalledPayload := string(marshalledPayload)
	s.log.Info("Starting execution of state machine", map[string]string{
		"arn":     stateMachineArn,
		"payload": stringifiedMarshalledPayload,
		"name":    executionName,
	})
	res, err := s.Client.StartExecutionWithContext(s.ctx, &sfn.StartExecutionInput{
		Input:           &stringifiedMarshalledPayload,
		Name:            &executionName,
		StateMachineArn: &stateMachineArn,
	})
	if err != nil {
		s.log.Error("State machine start execution error", err)
		return
	}
	s.log.Debug("State machine start execution response", res)
	return
}

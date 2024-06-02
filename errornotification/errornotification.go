package errornotification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gobase-lambda/aws"
	"gobase-lambda/log"
	"gobase-lambda/utils"
)

type ErrorNotifier struct {
	StackTrace   string `json:"stackTrace"`
	ErrorMessage string `json:"errorMessage"`
	StatusCode   string `json:"statusCode"`
	Log          log.Log
	Ctx          context.Context
}

func (en *ErrorNotifier) PublishToSqs() {
	newQueueName := utils.Getenv("error_notification_queue", "")
	if newQueueName != "" {
		queueUrl, err := aws.GetQueueURL(newQueueName, aws.GetAWSSQSClient(aws.GetDefaultAWSSession()), en.Ctx)
		if err != nil {
			en.Log.Error(fmt.Sprintf("Error while getting error queue url: %s", newQueueName), err)
		}
		sqsClient := aws.GetAWSSQSClient(aws.GetDefaultAWSSession())
		payload := map[string]string{
			"Status Code":   en.StatusCode,
			"Error Message": en.ErrorMessage,
			"Stack Trace":   en.StackTrace,
		}
		//newPayload, _ := utils.GetString(payload)
		bodyBlob := bytes.NewBuffer([]byte{})
		jsonEncoder := json.NewEncoder(bodyBlob)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.Encode(payload)
		newPayload := string(bodyBlob.String())
		req := &sqs.SendMessageInput{
			QueueUrl:          queueUrl,
			DelaySeconds:      nil,
			MessageBody:       &newPayload,
			MessageAttributes: nil,
		}
		_, err = sqsClient.SendMessageWithContext(en.Ctx, req)
		if err != nil {
			en.Log.Error("Error while pushing error message to error queue: ", err)
		}
		en.Log.Info("Error notification pushed to channel", nil)
	} else {
		en.Log.Info("Error notification queue not configured. Please set env variable error_notification_queue", nil)
	}

}

func (en *ErrorNotifier) PublishToSns() {
	errTopicArn := utils.Getenv("error_notification_sns_topic", "")
	if errTopicArn != "" {
		payload := map[string]string{
			"Status Code":   en.StatusCode,
			"Error Message": en.ErrorMessage,
			"Stack Trace":   en.StackTrace,
		}
		//newPayload, _ := utils.GetString(payload)
		bodyBlob := bytes.NewBuffer([]byte{})
		jsonEncoder := json.NewEncoder(bodyBlob)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.Encode(payload)
		newPayload := string(bodyBlob.String())
		snsClient := aws.GetDefaultSNSClient(en.Ctx)
		if errTopicArn != "" {
			subject := "Ckyc Event Error"
			req := &sns.PublishInput{
				TopicArn: &errTopicArn,
				Subject:  &subject,
				Message:  &newPayload,
			}
			publishOutput, err := snsClient.Client.PublishWithContext(en.Ctx, req)
			if err != nil {
				en.Log.Error(fmt.Sprintf("Error while publishing sns: %+s", err), err)
			}
			en.Log.Info("CKyc Evenet failure sns published", publishOutput)
		}
		en.Log.Info("Error notification pushed to sns channel", nil)
	} else {
		en.Log.Info("Error notification sns topic not configured. Please set env variable error_notification_sns_topic", nil)
	}

}

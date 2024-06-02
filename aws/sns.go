package aws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"gobase-lambda/log"
	"gobase-lambda/utils"
)

type SNS struct {
	_      struct{}
	Client *sns.SNS
	Log    *log.Log
	Ctx    context.Context
}

type SNSPayload struct {
	Entity map[string]interface{} `json:"entity"`
}

type SNSMessage struct {
	Entity   string                 `json:"entity"`
	Event    string                 `json:"event"`
	Contains []string               `json:"contains"`
	Payload  map[string]*SNSPayload `json:"payload"`
}

var defaultSNSClient *sns.SNS

func GetDefaultSNSClient(ctx context.Context) *SNS {
	if defaultSNSClient == nil {
		defaultSNSClient = GetAWSSNSClient(defaultAWSSession)
	}
	return GetSNSClient(ctx, defaultSNSClient)
}

func GetAWSSNSClient(awsSession *session.Session) *sns.SNS {
	client := sns.New(awsSession)
	return client
}

func GetSNSClient(ctx context.Context, client *sns.SNS) *SNS {
	return &SNS{Client: client, Log: log.GetDefaultLogger(), Ctx: ctx}
}

func (s *SNS) GetSNSDataTemplate(event string, eventData map[string]map[string]interface{}, attachment ...string) *SNSMessage {
	entity := strings.Split(event, ".")
	attachment = append(attachment, entity[0])
	message := &SNSMessage{
		Entity:   "event",
		Event:    event,
		Contains: attachment,
	}
	message.Payload = make(map[string]*SNSPayload)
	for _, v := range attachment {
		data, ok := eventData[v]
		if ok {
			data["entity"] = v
			message.Payload[v] = &SNSPayload{
				Entity: data,
			}

		} else {
			message.Payload[v] = &SNSPayload{
				Entity: map[string]interface{}{
					"entity": v,
				},
			}
		}
	}
	s.Log.Debug("SNS message generated", message)
	return message
}

func GetSNSARN(topicName string) (*string, error) {
	log := log.GetDefaultLogger()
	prefix := utils.Getenv("stage", "dev")
	systemPefix := utils.Getenv("snsTopicPrefix", "")
	if systemPefix != "" {
		prefix = fmt.Sprintf("%v_%v", prefix, systemPefix)
	}
	region := utils.GetenvMust("region")
	accountId := utils.GetenvMust("account_id")
	arn := fmt.Sprintf("arn:aws:sns:%v:%v:%v_%v", region, accountId, prefix, topicName)
	log.Debug("Topic Arn", arn)
	return &arn, nil
}

func (s *SNS) Publish(topicArn, subject *string, payload *SNSMessage, attributes map[string]string) error {
	// blob, _ := json.Marshal(payload)
	// message := string(blob)
	bodyBlob := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bodyBlob)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(payload)
	// bodyBlob, err := json.Marshal(response)
	// if err != nil {
	// 	return
	// }
	message := string(bodyBlob.String())
	req := &sns.PublishInput{
		TopicArn:          topicArn,
		Subject:           subject,
		Message:           &message,
		MessageAttributes: s.GetAttribure(attributes),
	}
	s.Log.Debug("SNS publish request", req)
	res, err := s.Client.PublishWithContext(s.Ctx, req)
	if err != nil {
		s.Log.Error("SNS publish error", err)
		return err
	}
	s.Log.Debug("SNS publish response", res)
	return nil
}

func (s *SNS) GetAttribure(attribute map[string]string) map[string]*sns.MessageAttributeValue {
	if len(attribute) == 0 {
		return nil
	}
	messageAttributes := make(map[string]*sns.MessageAttributeValue, len(attribute))
	for key, value := range attribute {
		messageAttributes[key] = &sns.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}
	return messageAttributes
}

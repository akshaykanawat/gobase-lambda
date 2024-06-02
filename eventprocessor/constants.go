package eventprocessor

import (
	"context"

	"gobase-lambda/log"
)

type EventType string

type NewEventProcessor func(context.Context, *log.Log, interface{}, EventType) EventProcessor

type EventProcessor interface {
	GetAPIHandler() map[string]map[string]*API
	GetSNSHandler() map[string]map[string]*SNS
	GetCronHandler() map[string]*CronInvocation
	GetSQSEventHandler() map[string]*SQS
	GetS3EventHandler() map[string]map[string]map[string]*S3Trigger
}

const (
	EventAPI  EventType = "API"
	EventCRON EventType = "CRON"
	EventSNS  EventType = "SNS"
	EventSQS  EventType = "SQS"
	EventS3   EventType = "S3"
)

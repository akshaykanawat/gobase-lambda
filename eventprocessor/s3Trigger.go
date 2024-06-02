package eventprocessor

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"gobase-lambda/errornotification"
	"gobase-lambda/utils"
)

type S3TriggerHandler func(payload interface{}) error

type S3Trigger struct {
	KeyPrefix        string
	S3TriggerHandler S3TriggerHandler
}

func (h *Handler) HandleS3TriggerRequest(ctx context.Context, request events.S3Event) (res events.APIGatewayProxyResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("Full Request", request)
			h.log.Error("Panic Stack", string(debug.Stack()))
			h.log.Error("Panic Recovery", r)
			switch v := r.(type) {
			case *utils.Error:
				res.StatusCode = v.StatusCode
				res.Body = fmt.Sprintf(`{"error":%v}`, v)
			default:
				res.StatusCode = http.StatusInternalServerError
				res.Body = `{"error": "Error occurred please try after some time, if persist contact technical support"}`
				notification := errornotification.ErrorNotifier{
					StatusCode:   "500",
					StackTrace:   string(debug.Stack()),
					ErrorMessage: fmt.Sprintf("%s", r),
					Log:          *h.log,
				}
				notification.PublishToSqs()
				notification.PublishToSns()
			}
		} else {
			if err != nil {
				h.log.Error("SQS Event Error", err)
				custErr, ok := err.(*utils.Error)
				if ok {
					res.StatusCode = custErr.StatusCode
				} else {
					res.StatusCode = http.StatusInternalServerError
				}
				res.Body = fmt.Sprintf(`{"error":%v}`, err)
			}
		}
		h.log.Info("Full Response", res)
	}()
	h.setCorrelationParams(map[string]string{})
	h.log.Debug("Full Request", request)
	eventProcessor := h.eventProcessorFunc(ctx, h.log, &request, EventS3)
	s3TriggerMap := eventProcessor.GetS3EventHandler()
	bucket := request.Records[0].S3.Bucket.Name
	objectKey := request.Records[0].S3.Object.Key
	eventName := request.Records[0].EventName
	newHandler := extractS3TriggerHandler(s3TriggerMap, eventName, objectKey, bucket)
	err = newHandler.S3TriggerHandler(&request)

	return
}

func extractS3TriggerHandler(s3TriggerMap map[string]map[string]map[string]*S3Trigger, eventName, objectKey, bucket string) *S3Trigger {
	if s3TriggerMap[bucket] != nil {
		if s3TriggerMap[bucket][eventName] != nil {
			keyMap := s3TriggerMap[bucket][eventName]
			keyPrefixes := make([]string, 0, len(keyMap))
			for k := range keyMap {
				keyPrefixes = append(keyPrefixes, k)
			}
			for _, a := range keyPrefixes {
				if strings.Contains(objectKey, a) {
					triggerObj := s3TriggerMap[bucket][eventName][a]
					return triggerObj
				}
			}
		}
	}
	panic(utils.NewHTTPBadRequestError("Unknown S3 Path of triggered event name configured", s3TriggerMap))
}

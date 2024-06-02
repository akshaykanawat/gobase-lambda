package eventprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"gobase-lambda/log"
	"gobase-lambda/utils"
)

type SNSHandler func(payload interface{}) error

type SNS struct {
	Topic      string
	Event      string
	SnsHandler SNSHandler
}

func (h *Handler) HandleSNSRequest(ctx context.Context, request events.SNSEvent) (res events.APIGatewayProxyResponse, err error) {
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
			}
		} else {
			if err != nil {
				h.log.Error("SNS Error", err)
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
	eventProcessor := h.eventProcessorFunc(ctx, h.log, &request, EventSNS)
	snsMap := eventProcessor.GetSNSHandler()
	event, topic, payload := extractSNSRequest(&request)
	topicMap, topicExists := snsMap[topic]
	if topicExists {
		handler, eventExists := topicMap[event]
		if eventExists {
			err = handler.SnsHandler(payload)
			if err != nil {
				panic(err)
			}
			res.StatusCode = http.StatusNoContent
			res.Body = ""
		} else {
			errorMessage := fmt.Sprintf("event %v %v is not mapped", topic, event)
			h.log.Alert(errorMessage, snsMap)
			panic(utils.NewHTTPNotFoundError(errorMessage, nil))
		}
	} else {
		errorMessage := fmt.Sprintf("topic %v not mapped", topic)
		h.log.Alert(errorMessage, snsMap)
		panic(utils.NewHTTPNotFoundError(errorMessage, nil))
	}
	return
}

func extractSNSRequest(request *events.SNSEvent) (event, topic string, payload map[string]interface{}) {
	logger := log.GetDefaultLogger()
	topicArnSplit := strings.Split(request.Records[0].SNS.TopicArn, ":")
	topic = topicArnSplit[len(topicArnSplit)-1]
	unmarshallError := json.Unmarshal([]byte(request.Records[0].SNS.Message), &payload)
	if unmarshallError != nil {
		logger.Error("payload unmarshal failed", unmarshallError.Error())
		return
	}
	eventIf, ok := payload["event"]
	if !ok {
		logger.Error("event not found", payload)
		panic(utils.NewHTTPBadRequestError("Unknown event structure", payload))
	}
	event = eventIf.(string)
	logger.Info("Topic", topic)
	logger.Info("Event", event)
	logger.Info("Payload", payload)
	return
}

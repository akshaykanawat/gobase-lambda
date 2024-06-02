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

type SQSHandler func(payload interface{}) error

type SQS struct {
	SQSHandler SQSHandler
}

func (h *Handler) HandleSQSRequest(ctx context.Context, request events.SQSEvent) (res events.APIGatewayProxyResponse, err error) {
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
	eventProcessor := h.eventProcessorFunc(ctx, h.log, &request, EventSQS)
	sqsMap := eventProcessor.GetSQSEventHandler()
	queueArn := request.Records[0].EventSourceARN
	newHandler := extractQueueHandler(sqsMap, queueArn)
	err = newHandler.SQSHandler(&request)
	return
}

func extractQueueHandler(sqsMap map[string]*SQS, queueArn string) *SQS {
	stage := utils.GetenvMust("stage")
	for key, value := range sqsMap {
		queueSlice := strings.Split(queueArn, ":")
		queueName := queueSlice[len(queueSlice)-1]
		key = fmt.Sprintf(`%s_%s`, stage, key)
		if strings.EqualFold(queueName, key) {
			return value
		}
	}
	panic(utils.NewHTTPBadRequestError("Unknown Queue name configured", sqsMap))
}

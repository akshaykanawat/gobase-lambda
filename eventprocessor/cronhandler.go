package eventprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/aws/aws-lambda-go/events"
	"gobase-lambda/utils"
)

type CronEvent struct {
	IsCron     bool        `json:"isCron"`
	ActionName string      `json:"actionName"`
	Payload    interface{} `json:"payload"`
}

type CronHandler func(payload interface{}) (statusCode int, response interface{}, err error)

type CronInvocation struct {
	CronHandlerFunc CronHandler
}

func (h *Handler) HandleCronInvocation(ctx context.Context, request CronEvent) (res events.APIGatewayProxyResponse, err error) {
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
				h.log.Error("Cron Error", err)
				custErr, ok := err.(*utils.Error)
				if ok {
					res.StatusCode = custErr.StatusCode
				} else {
					res.StatusCode = http.StatusInternalServerError
				}
				res.Body = fmt.Sprintf(`{"error":%v}`, err)
			}
		}
		h.log.Info("Response", res)
	}()
	h.setCorrelationParams(map[string]string{})
	h.log.Debug("Full Request", request)

	eventProcessor := h.eventProcessorFunc(ctx, h.log, &request, EventCRON)
	cronHandlerMap := eventProcessor.GetCronHandler()
	actionName, payload := request.ActionName, request.Payload
	actionHandler, ok := cronHandlerMap[actionName]
	var response interface{}
	var statusCode int
	if ok {
		statusCode, response, err = actionHandler.CronHandlerFunc(payload)
		if err != nil {
			return
		}
		resObj, ok := response.(events.APIGatewayProxyResponse)
		if ok {
			res = resObj
		} else {
			res.StatusCode = statusCode
			var bodyBlob []byte
			bodyBlob, err = json.Marshal(response)
			if err != nil {
				return
			}
			res.Body = string(bodyBlob)
		}
	} else {
		errorMessage := fmt.Sprintf("action %v is not mapped", actionName)
		h.log.Alert(errorMessage, cronHandlerMap)
		panic(utils.NewHTTPNotFoundError(errorMessage, nil))
	}
	return
}

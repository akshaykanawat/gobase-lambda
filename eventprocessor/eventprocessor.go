package eventprocessor

import (
	"strconv"

	"gobase-lambda/aws"
	"gobase-lambda/log"
	"gobase-lambda/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Handler struct {
	isLambda           bool
	log                *log.Log
	eventProcessorFunc NewEventProcessor
}

func (h *Handler) setAWSSession() {
	if h.isLambda {
		awsSession := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		aws.SetDefaultAWSSession(awsSession)
	} else {
		if aws.GetDefaultAWSSession() == nil {
			panic("AWS Session is not set")
		}
	}
}

func GetHandler(isAWSEnv bool, eventProcessorCreator NewEventProcessor) *Handler {
	var logLevel log.LogLevel
	logLevelVal, err := strconv.Atoi(utils.Getenv("LOG_LEVEL", strconv.Itoa(int(log.INFO))))
	if err != nil {
		logLevel = log.INFO
	} else {
		logLevel = log.LogLevel(logLevelVal)
	}
	logger := log.NewLogger(isAWSEnv, logLevel, nil)
	log.SetDefaultLogger(logger)
	handler := &Handler{isLambda: isAWSEnv, eventProcessorFunc: eventProcessorCreator, log: logger}
	handler.setAWSSession()
	return handler
}

func (h *Handler) setCorrelationParams(correlationParams map[string]string) {
	h.log.SetCorrelationParams(correlationParams)
}

var reduced = "<<<reduced>>>"
var reducedList = []string{reduced}

func (h *Handler) printReducedRequest(request *events.APIGatewayProxyRequest, isReduceBody bool) {

	body := request.Body
	if isReduceBody {
		request.Body = reduced
	}
	context := request.RequestContext
	request.RequestContext = events.APIGatewayProxyRequestContext{}
	authorization, aok := request.Headers["Authorization"]
	if aok {
		request.Headers["Authorization"] = reduced
	}
	mAuthorization, mok := request.MultiValueHeaders["Authorization"]
	if mok {
		request.MultiValueHeaders["Authorization"] = reducedList
	}
	h.log.Info("API Request", request)
	request.RequestContext = context
	request.Body = body
	if aok {
		request.Headers["Authorization"] = authorization
	}
	if mok {
		request.MultiValueHeaders["Authorization"] = mAuthorization
	}
}

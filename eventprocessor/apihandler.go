package eventprocessor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fatih/structs"
	"gobase-lambda/errornotification"
	"gobase-lambda/log"
	"gobase-lambda/utils"
)

type APIHandler func(headers interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error)

func (h *Handler) HandleAPIRequest(ctx context.Context, request events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
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
			h.log.Info("Response", res)
		} else {
			if err != nil {
				h.log.Error("Full Request", request)
				h.log.Error("API Error", err)
				custErr, ok := err.(*utils.Error)
				if ok {
					res.StatusCode = custErr.StatusCode
				} else {
					res.StatusCode = http.StatusInternalServerError
				}
				res.Body = fmt.Sprintf(`{"error":%v}`, err)
				h.log.Info("Response", res)
			} else {
				if res.StatusCode > 299 {
					h.printReducedRequest(&request, false)
					h.log.Info("Response", res)
				} else {
					h.log.Debug("Full Response", res)
					str := res.Body
					res.Body = "<<<redacted>>>"
					h.log.Info("Response", res)
					res.Body = str
				}
			}
		}
		err = nil
	}()
	h.setCorrelationParams(request.Headers)
	h.log.Debug("Full Request", request)
	eventProcessor := h.eventProcessorFunc(ctx, h.log, &request, EventAPI)
	h.printReducedRequest(&request, true)
	apiMap := eventProcessor.GetAPIHandler()
	resource, method := request.Resource, request.HTTPMethod
	methodMap, ok := apiMap[resource]
	matchFound := false
	var statusCode int
	var response interface{}
	if ok {
		handler, ok := methodMap[method]
		if ok {
			var headers, queryParams, pathParams interface{}
			var contentType string
			matchFound = true
			headers, queryParams, pathParams, contentType, jsonBody := extractAPIRequest(handler, &request)
			if method == http.MethodGet {
				statusCode, response, err = handler.ApiHandler(headers, pathParams, "", queryParams)
			} else {
				statusCode, response, err = handler.ApiHandler(headers, pathParams, jsonBody, nil)
			}
			if err != nil {
				return
			}
			resObj, ok := response.(events.APIGatewayProxyResponse)
			if ok {
				res = resObj
			} else {
				res.StatusCode = statusCode
				res.Body, res.IsBase64Encoded = ProcessResponse(contentType, response)
			}
			return
		}

	}
	if !matchFound {
		errorMessage := fmt.Sprintf("path not found %v, %v", resource, method)
		h.log.Alert(errorMessage, apiMap)
		panic(utils.NewHTTPNotFoundError(errorMessage, nil))
	}
	return
}

// func ProcessRequestBody(requestBody string) (body []byte) {
// 	body, err := base64.StdEncoding.DecodeString(requestBody)
// 	if err != nil {
// 		panic(utils.NewHTTPBadRequestError(fmt.Sprintf("request processing failed : %v", err), body))
// 	}
// 	return
// }

func ProcessResponse(ct string, response interface{}) (resBody string, base64Encoding bool) {
	switch contentType := strings.ToLower(ct); contentType {
	case "application/x-protobuf":
		u, err := response.([]byte)
		if !err {
			panic(utils.NewHTTPBadRequestError(fmt.Sprintf("response processing failed : %v", err), response))
		}
		resBody = base64.StdEncoding.EncodeToString(u)
		base64Encoding = true
	default:
		bodyBlob := bytes.NewBuffer([]byte{})
		jsonEncoder := json.NewEncoder(bodyBlob)
		jsonEncoder.SetEscapeHTML(false)
		jsonEncoder.Encode(response)
		// bodyBlob, err := json.Marshal(response)
		// if err != nil {
		// 	return
		// }
		resBody = string(bodyBlob.String())
		base64Encoding = false
	}
	return
}

type API struct {
	Resource    string
	Method      string
	ApiHandler  APIHandler
	Body        interface{}
	QueryParams interface{}
	PathParams  interface{}
}

func extractAPIRequest(apiMap *API, request *events.APIGatewayProxyRequest) (headers interface{}, queryParams interface{}, pathParams interface{}, contentType string, jsonBody string) {
	logger := log.GetDefaultLogger()
	contentType, notExists := request.Headers["Content-Type"]
	if notExists {
		contentType = "application/json"
	}
	headers = request.Headers
	if request.HTTPMethod == http.MethodGet {
		queryParams = getQueryStringParams(apiMap, request, logger)
	} else {
		jsonBody = request.Body
	}
	pathParams = getPathParams(apiMap, request, logger)
	return
}

func getBody(apiMap *API, request *events.APIGatewayProxyRequest, logger *log.Log, contentType string) (body interface{}) {
	body = request.Body
	if apiMap.Body != nil && request.Body != "" {
		switch ct := strings.ToLower(contentType); ct {
		case "application/x-protobuf":
			body, err := base64.StdEncoding.DecodeString(apiMap.Body.(string))
			if err != nil {
				panic(utils.NewHTTPBadRequestError(fmt.Sprintf("body unmarshal failed : %v", err), request.Body))
			}
			return body
		default:
			err := json.Unmarshal([]byte(request.Body), &apiMap.Body)
			if err != nil {
				panic(utils.NewHTTPBadRequestError(fmt.Sprintf("body unmarshal failed : %v", err), request.Body))
			}
			body = apiMap.Body
		}
	}
	return
}

func getQueryStringParams(apiMap *API, request *events.APIGatewayProxyRequest, logger *log.Log) (queryStringParams interface{}) {
	queryStringParams = request.MultiValueQueryStringParameters
	if apiMap.QueryParams != nil {
		queryParamMap := structs.Map(apiMap.QueryParams)
		queryParamValues := make(map[string]interface{}, len(queryParamMap))
		for key, value := range queryParamMap {
			switch v := value.(type) {
			case string:
				reqValue, ok := request.QueryStringParameters[key]
				if ok {
					queryParamValues[key] = reqValue
				}
			case []string:
				reqValue, ok := request.MultiValueQueryStringParameters[key]
				if ok {
					queryParamValues[key] = reqValue
				}
			case int:
				reqValue, ok := request.QueryStringParameters[key]
				if ok {
					intVal, err := strconv.Atoi(reqValue)
					if err != nil {
						panic(utils.NewHTTPBadRequestError(fmt.Sprintf("invalid value for int, %v-%v", key, reqValue), request.QueryStringParameters))
					}
					queryParamValues[key] = intVal
				}
			case []int:
				reqValue, ok := request.MultiValueQueryStringParameters[key]
				if ok {
					intList := make([]int, len(reqValue))
					for i, val := range reqValue {
						intVal, err := strconv.Atoi(val)
						if err != nil {
							panic(utils.NewHTTPBadRequestError(fmt.Sprintf("invalid value for int, %v-%v", key, reqValue), request.MultiValueQueryStringParameters))
						}
						intList[i] = intVal
					}
					queryParamValues[key] = intList
				}
			case bool:
				reqValue, ok := request.QueryStringParameters[key]
				if ok {
					var flag bool
					if reqValue == "true" {
						flag = true
					} else if reqValue == "false" {
						flag = false
					} else {
						panic(utils.NewHTTPBadRequestError(fmt.Sprintf("invalid value for bool %v - %v", key, v), request.QueryStringParameters))
					}
					queryParamValues[key] = flag
				}
			default:
				panic(utils.NewHTTPBadRequestError(fmt.Sprintf("%T is not supported for query string params", v), apiMap.QueryParams))
			}
		}
		jsonString, _ := json.Marshal(queryParamValues)
		json.Unmarshal(jsonString, apiMap.QueryParams)
		queryStringParams = apiMap.QueryParams
	}
	return
}

func getPathParams(apiMap *API, request *events.APIGatewayProxyRequest, logger *log.Log) (pathParams interface{}) {
	pathParams = request.PathParameters
	if apiMap.PathParams != nil {
		blob, err := json.Marshal(request.PathParameters)
		if err != nil {
			panic(fmt.Errorf("path param unmarshal failed : %v", err))
		}
		err = json.Unmarshal(blob, apiMap.PathParams)
		if err != nil {
			panic(fmt.Errorf("path param unmarshal failed : %v", err))
		}
		pathParams = apiMap.PathParams
	}
	return
}

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urllib "net/url"
	"strings"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"gobase-lambda/log"
	"gobase-lambda/utils"
)

const (
	HTTPMethodsGET  string = http.MethodGet
	HTTPMethodsPOST string = http.MethodPost
)

const (
	ContentTypeJSON string = "application/json"
	ContentTypeFORM string = "application/x-www-form-urlencoded"
)

type HTTP struct {
	client *http.Client
	ctx    context.Context
	log    *log.Log
}

func NewHTTPClient(ctx context.Context) *HTTP {
	return &HTTP{client: xray.Client(nil), ctx: ctx, log: log.GetDefaultLogger()}
}

func (h *HTTP) Get(url string, queryParams, header map[string]string, timeoutInSeconds int64) (response *http.Response, responseBody []byte, err error) {
	return h.process(HTTPMethodsGET, url, "", nil, header, timeoutInSeconds, queryParams)
}

func (h *HTTP) Post(url string, contentType string, body interface{}, header map[string]string, timeoutInSeconds int64) (response *http.Response, responseBody []byte, err error) {
	return h.process(HTTPMethodsPOST, url, contentType, body, header, timeoutInSeconds, nil)
}

func (h *HTTP) process(method string, url string, contentType string, body interface{}, headers map[string]string, timeoutInSeconds int64, queryParams map[string]string) (response *http.Response, responseBody []byte, err error) {
	if method == HTTPMethodsGET {
		response, responseBody, err = h.send(method, url, contentType, nil, headers, timeoutInSeconds, queryParams)
	} else if method == HTTPMethodsPOST {
		if contentType == ContentTypeJSON {
			var jsonContent []byte
			jsonContent, err = json.Marshal(body)
			if err == nil {
				response, responseBody, err = h.send(method, url, contentType, bytes.NewBuffer(jsonContent), headers, timeoutInSeconds, nil)
			}
		} else if contentType == ContentTypeFORM {
			valuesMap := body.(map[string][]string)
			values := make(urllib.Values)
			for key, value := range valuesMap {
				values[key] = value
			}
			response, responseBody, err = h.send(method, url, contentType, strings.NewReader(values.Encode()), headers, timeoutInSeconds, nil)
		} else {
			errorMessage := fmt.Sprintf("Invalid content type %v", contentType)
			h.log.Error(errorMessage, nil)
			err = utils.NewError(http.StatusInternalServerError, errorMessage, "INVALID_CONTENT_TYPE", nil)
		}
	} else {
		errorMessage := fmt.Sprintf("invalid http method %v", method)
		h.log.Error("Invalid http method", method)
		err = utils.NewError(http.StatusInternalServerError, errorMessage, "INVALID_METHOD", nil)
	}
	return
}

func (h *HTTP) send(method string, url string, contentType string, payload io.Reader, headers map[string]string, timeoutInSeconds int64, queryParams map[string]string) (*http.Response, []byte, error) {
	h.client.Timeout = time.Duration(timeoutInSeconds) * time.Second
	correlationParams := h.log.GetCorrelationParams()
	if headers == nil {
		headers = make(map[string]string)
	}
	for key, value := range correlationParams {
		headers[key] = value
	}
	req, err := http.NewRequestWithContext(h.ctx, method, url, payload)
	if method == http.MethodGet {
		query := req.URL.Query()
		for key, value := range queryParams {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", contentType)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	h.log.Debug("API request", map[string]interface{}{
		"headers": req.Header,
		"method":  method,
		"url":     url,
	})
	res, err := h.client.Do(req)
	if err != nil {
		h.log.Debug("API call error", err)
		return nil, nil, err
	}
	h.log.Debug("API response", map[string]interface{}{
		"statusCode": res.StatusCode,
		"headers":    res.Header,
	})
	var bodyBytes []byte
	xray.Capture(h.ctx, "ReadAPIResponseBody", func(ctx1 context.Context) error {
		bodyBytes, err = io.ReadAll(res.Body)
		return nil
	})
	if err != nil {
		h.log.Debug("API response body read error", err)
		return nil, nil, err
	}
	return res, bodyBytes, nil
}

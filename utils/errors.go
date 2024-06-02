package utils

import (
	"encoding/json"
	"net/http"

	"gobase-lambda/log"
)

type Error struct {
	StatusCode   int         `json:"-"`
	ErrorData    interface{} `json:"errorData"`
	ErrorMessage string      `json:"errorMessage"`
	ErrorCode    string      `json:"errorCode"`
}

func (e *Error) Error() string {
	blob, err := json.Marshal(e)
	if err != nil {
		logger := log.GetDefaultLogger()
		logger.Emergency("Error in marshalling of custom error type", err)
		logger.Emergency("Error data", e)
	}
	return string(blob)
}

func NewHTTPBadRequestError(errorMessage string, errorData interface{}) *Error {
	return NewError(http.StatusBadRequest, errorMessage, "BAD_REQUEST", errorData)
}

func NewHTTPNotFoundError(errorMessage string, errorData interface{}) *Error {
	return NewError(http.StatusNotFound, errorMessage, "NOT_FOUND", errorData)
}

func NewError(statusCode int, errorMessage string, errorCode string, errorData interface{}) *Error {
	return &Error{StatusCode: statusCode, ErrorMessage: errorMessage, ErrorData: errorData, ErrorCode: errorCode}
}

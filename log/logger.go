package log

import (
	"encoding/json"
	"github.com/google/uuid"
	"os"
)

type LogLevel int

const (
	EMERGENCY LogLevel = iota
	ALERT
	CRITICAL
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var logMap = map[LogLevel]string{
	EMERGENCY: "EMERGENCY",
	ALERT:     "ALERT",
	CRITICAL:  "CRITICAL",
	ERROR:     "ERROR",
	WARNING:   "WARNING",
	NOTICE:    "NOTICE",
	INFO:      "INFO",
	DEBUG:     "DEBUG",
}

type CorrelationParams struct {
	CorrelationId string `json:"x-correlation-id"`
	ScenarioId    string `json:"x-scenario-id"`
	SessionId     string `json:"x-session-id"`
	ScenarioName  string `json:"x-scenario-name"`
	ServiceName   string `json:"x-service-name"`
}

type Log struct {
	logLevel          LogLevel
	correlationParams *CorrelationParams
	printer           LogPrinter
}

var AuditEndpoint = map[string]string{
	"live":    "http://cam-prod.nm.wealth/prod/cam/v2/audit-events",
	"staging": "https://d6o0fhi2nl.execute-api.ap-south-1.amazonaws.com/dev/echo/staging/auditlog",
	"dev":     "https://d6o0fhi2nl.execute-api.ap-south-1.amazonaws.com/dev/echo/dev/auditlog",
}

var AuditHeaders = map[string]map[string]string{
	"live": {
		"Content-Type":   "application/json",
		"x-apigw-api-id": "i5p8zcz0pg",
	},
}

type Context struct {
	Entity   string `json:"entity"`
	EntityId string `json:"entity_id"`
}

type Target struct {
	CustomerId string `json:"customer_id"`
}

type AuditTemplate struct {
	Id            string  `json:"id"`
	Action        string  `json:"action"`
	Actor         string  `json:"actor"`
	Target        Target  `json:"target"`
	Application   string  `json:"application"`
	TenantId      string  `json:"tenant_id"`
	Timestamp     string  `json:"timestamp"`
	Context       Context `json:"context"`
	SchemaVersion string  `json:"schema_version"`
	CorrelationId string  `json:"correlation_id"`
}

var defaultLogger *Log

func NewLogger(isAWSEnv bool, logLevel LogLevel, correlationParams map[string]string) *Log {
	var printer LogPrinter
	logger := &Log{logLevel: logLevel}

	if correlationParams != nil {
		logger.SetCorrelationParams(correlationParams)
	} else {
		logger.correlationParams = &CorrelationParams{}
	}
	if isAWSEnv {
		printer = &CloudWatchPrinter{}
	} else {
		printer = &LocalPrinter{}
	}
	logger.printer = printer
	return logger
}

func SetDefaultLogger(log *Log) {
	defaultLogger = log
}

func GetDefaultLogger() *Log {
	return defaultLogger
}

func (l *Log) GetCorrelationParams() map[string]string {
	blob, _ := json.Marshal(l.correlationParams)
	val := make(map[string]string)
	json.Unmarshal(blob, &val)
	return val
}

func (l *Log) SetCorrelationParams(correlationParams map[string]string) {
	correlationObj := &CorrelationParams{}
	correlationObj.CorrelationId = correlationParams["x-correlation-id"]
	correlationObj.ScenarioId = correlationParams["x-scenario-id"]
	correlationObj.SessionId = correlationParams["x-session-id"]
	correlationObj.ScenarioName = correlationParams["x-scenario-name"]
	correlationObj.ServiceName = os.Getenv("service_name")
	if correlationObj.ServiceName == "" {
		correlationObj.ServiceName = "GOAPP"
	}
	l.correlationParams = correlationObj
	if correlationObj.CorrelationId == "" {
		id, _ := uuid.NewUUID()
		correlationObj.CorrelationId = id.String()
	}
}

func (l *Log) Debug(message string, object interface{}) {
	l.print(DEBUG, &message, object)
}

func (l *Log) Info(message string, object interface{}) {
	l.print(INFO, &message, object)
}

func (l *Log) Notice(message string, object interface{}) {
	l.print(NOTICE, &message, object)
}

func (l *Log) Warning(message string, object interface{}) {
	l.print(WARNING, &message, object)
}

func (l *Log) Error(message string, object interface{}) {
	l.print(ERROR, &message, object)
}

func (l *Log) Critical(message string, object interface{}) {
	l.print(CRITICAL, &message, object)
}

func (l *Log) Alert(message string, object interface{}) {
	l.print(ALERT, &message, object)
}

func (l *Log) Emergency(message string, object interface{}) {
	l.print(EMERGENCY, &message, object)
}

func (l *Log) print(logLevel LogLevel, messages *string, object interface{}) {
	if logLevel > l.logLevel {
		return
	}
	l.printer.Print(logLevel, messages, object, l.correlationParams)
}

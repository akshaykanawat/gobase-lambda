package log

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type LogPrinter interface {
	Print(logLevel LogLevel, message *string, object interface{}, correlationParams *CorrelationParams)
}

type CloudWatchPrinter struct{}

func (p *CloudWatchPrinter) Print(logLevel LogLevel, message *string, object interface{}, correlationParams *CorrelationParams) {
	log := map[string]interface{}{
		"version":           "1.1",
		"short_message":     *message,
		"full_message":      object,
		"timestamp":         time.Now().Unix(),
		"level":             logLevel,
		"_log_level_name":   logMap[logLevel],
		"_x_correlation_id": correlationParams.CorrelationId,
		"_x_scenario_id":    correlationParams.ScenarioId,
		"_x_session_id":     correlationParams.SessionId,
		"_x_scenario_name":  correlationParams.ScenarioName,
		"_service_name":     correlationParams.ServiceName,
		"_object_type":      reflect.TypeOf(object),
	}
	logBlob, err := json.Marshal(log)
	if err != nil {
		message := "Error occured during log print"
		p.Print(ERROR, &message, err, correlationParams)
		message = "Probably trying to print interface type"
		p.Print(NOTICE, &message, nil, correlationParams)
	}
	fmt.Println(string(logBlob))
}

type LocalPrinter struct {
}

func (p *LocalPrinter) Print(logLevel LogLevel, message *string, object interface{}, correlationParams *CorrelationParams) {
	if object != nil && reflect.ValueOf(object).Type().Kind() == reflect.Ptr {
		object = reflect.Indirect(reflect.ValueOf(object))
	}
	fmt.Printf("%v %v-%v-%+v \n", time.Now().Format("2006-01-02 15:04:05"), logMap[logLevel], *message, object)
}

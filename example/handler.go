package example

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/google/uuid"
	"gobase-lambda/aws"
	"gobase-lambda/eventprocessor"
	"gobase-lambda/log"
	"gobase-lambda/utils"
	"gobase-lambda/utils/http"
)

type Manager struct {
	log       *log.Log
	ctx       context.Context
	event     interface{}
	eventType eventprocessor.EventType
}

func (m *Manager) GetAPIHandler() map[string]map[string]*eventprocessor.API {
	resource := "/{customerId}"
	getAPI := &eventprocessor.API{
		Resource:    resource,
		Method:      "GET",
		ApiHandler:  m.Get,
		PathParams:  &PathParams{},
		QueryParams: &QueryParams{},
	}
	postAPI := &eventprocessor.API{
		Resource:    resource,
		Method:      "POST",
		ApiHandler:  m.Post,
		Body:        &Payload{},
		PathParams:  &PathParams{},
		QueryParams: &QueryParams{},
	}
	uploadDoc := &eventprocessor.API{
		Resource:   "/upload",
		Method:     "POST",
		ApiHandler: m.Upload,
		Body:       &Upload{},
	}
	uploadPiiDoc := &eventprocessor.API{
		Resource:   "/pii/upload",
		Method:     "POST",
		ApiHandler: m.UploadPII,
		Body:       &Upload{},
	}
	downloadDoc := &eventprocessor.API{
		Resource:   "/download",
		Method:     "POST",
		ApiHandler: m.Download,
		Body:       &Upload{},
	}
	downloadPiiDoc := &eventprocessor.API{
		Resource:   "/pii/download",
		Method:     "POST",
		ApiHandler: m.DownloadPII,
		Body:       &Upload{},
	}
	stepFuncInvocation := &eventprocessor.API{
		Resource:    "/step-func",
		Method:      "GET",
		ApiHandler:  m.StepFuncInvocation,
		PathParams:  nil,
		QueryParams: nil,
	}
	return map[string]map[string]*eventprocessor.API{
		resource: {
			getAPI.Method:  getAPI,
			postAPI.Method: postAPI,
		},
		uploadDoc.Resource: {
			uploadDoc.Method: uploadDoc,
		},
		uploadPiiDoc.Resource: {
			uploadPiiDoc.Method: uploadPiiDoc,
		},
		downloadDoc.Resource: {
			downloadDoc.Method: downloadDoc,
		},
		downloadPiiDoc.Resource: {
			downloadPiiDoc.Method: downloadPiiDoc,
		},
		"/step-func": {
			stepFuncInvocation.Method: stepFuncInvocation,
		},
	}
}

func (m *Manager) GetSNSHandler() map[string]map[string]*eventprocessor.SNS {
	topic := "dev_MFCORE_PAYMENT"
	event_rejected, event_confirmed := "transaction.rejected", "transaction.confirmed"
	return map[string]map[string]*eventprocessor.SNS{
		topic: {
			event_rejected: &eventprocessor.SNS{
				Topic:      topic,
				Event:      event_rejected,
				SnsHandler: m.TransactionRejected,
			},
			event_confirmed: &eventprocessor.SNS{
				Topic:      topic,
				Event:      event_confirmed,
				SnsHandler: m.TransactionConfirmed,
			},
		},
	}
}

func (m *Manager) GetLambdaHandler() map[string]*eventprocessor.LambdaInvocation {
	return nil
}

func (m *Manager) GetSQSEventHandler() map[string]*eventprocessor.SQS {
	t := &eventprocessor.SQS{
		SQSHandler: m.test_func,
	}
	new := map[string]*eventprocessor.SQS{"TEST_QUEUE": t}
	return new
}

func (m *Manager) test_func(body interface{}) error {
	m.log.Info("This is sqs test body. ", body)
	return nil
}

func (m *Manager) GetCronHandler() map[string]*eventprocessor.CronInvocation {
	return map[string]*eventprocessor.CronInvocation{
		"CRON_ACTION_1": {
			CronHandlerFunc: m.CronActionOne,
		}, "CRON_ACTION_2": {
			CronHandlerFunc: m.CronActionTwo,
		},
	}
}

func (m *Manager) Get(queryParam interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	m.log.Info("Queryparams", queryParam)
	m.log.Info("PathParams", pathParam)
	snsClient := aws.GetDefaultSNSClient(m.ctx)
	os.Setenv("snsTopicPrefix", "MFCORE")
	arn, err := aws.GetSNSARN("PAYMENT")
	if err != nil {
		return 500, nil, err
	}
	data := map[string]map[string]interface{}{
		"payment": {
			"id":     "pay_14341234",
			"amount": 123,
		},
		"bank": {
			"id":                "bank_fadsfas",
			"bankAccountNumber": "0000021312",
		},
		"customer": {
			"id": "cust_fasdfsa",
		},
	}
	message := snsClient.GetSNSDataTemplate("payment.success", data, "bank", "customer")
	err = snsClient.Publish(arn, nil, message, nil)
	if err != nil {
		return 500, nil, err
	}
	kms := aws.GetDefaultKMSClient(m.ctx, "arn:aws:kms:ap-south-1:490302598154:key/489a37f4-4ef0-408f-9ecf-7dd629505060")
	text := "asfasdfsaf"
	_, encryptedText, err := kms.Encrypt(&text)
	if err != nil {
		return 500, nil, err
	}
	plainText, err := kms.Decrypt(&encryptedText)
	if err != nil {
		return 500, nil, err
	}
	if plainText != text {
		m.log.Error("Texts are not matching", nil)
		return 500, nil, err
	}
	apiClient := http.NewHTTPClient(m.ctx)
	_, bodyBytes, err := apiClient.Post("https://api-gateway-nm.fnpaas.com/mfcore/dev/bedrockGobaseTest/cust_1241243", http.ContentTypeJSON, map[string]string{}, map[string]string{"x-mfcore-client-id": "44i9la6ck95ekll48df9it854k"}, 10)
	if err != nil {
		return 500, nil, err
	}
	m.log.Info("API Body", string(bodyBytes))
	_, bodyBytes, err = apiClient.Post("https://d6o0fhi2nl.execute-api.ap-south-1.amazonaws.com/dev/echo/golang/test", http.ContentTypeJSON, `{}`, nil, 10)
	if err != nil {
		return 500, nil, err
	}
	m.log.Info("API Body", string(bodyBytes))
	return 200, map[string]interface{}{
		"queryParams": queryParam,
		"pathParams":  pathParam,
	}, nil

}

func (m *Manager) Post(body interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	m.log.Info("Body", body)
	m.log.Info("PathParams", pathParam)
	queueUrl, err := aws.GetQueueURL("DIGIO_WEBHOOK", aws.GetAWSSQSClient(aws.GetDefaultAWSSession()), m.ctx)
	if err != nil {
		return 500, nil, err
	}
	sqsClient := aws.GetDefaultSQSClient(m.ctx, *queueUrl)
	err = sqsClient.SendMessage(map[string]string{
		"hello": "Fasdfasdf-" + uuid.New().String(),
	}, map[string]string{
		"id": uuid.NewString(),
	}, 1, nil, nil)
	if err != nil {
		return 500, nil, err
	}
	messageList, err := sqsClient.ReceiveMessage(10, 10, 3)
	if err != nil {
		return 500, nil, err
	}
	err = sqsClient.DeleteMessage(messageList[0].ReceiptHandle)
	if err != nil {
		return 500, nil, err
	}
	client := aws.GetDefaultSecretManagerClient(m.ctx)
	_, err = client.GetSecret("arn:aws:secretsmanager:ap-south-1:490302598154:secret:dev/signzy-XnLNDf")
	if err != nil {
		return 500, nil, err
	}
	return 200, map[string]interface{}{
		"body":       body,
		"pathParams": pathParam,
	}, nil

}

func (m *Manager) Upload(body interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	m.log.Info("Body", body)
	req, ok := body.(*Upload)
	if !ok {
		panic(utils.NewHTTPBadRequestError("Error in reading payload", body))
	}
	apiClient := http.NewHTTPClient(m.ctx)
	res, responseBody, err := apiClient.Get(req.URL, nil, nil, 10)
	if err != nil {
		return 500, nil, err
	}
	s3Client := aws.GetDefaultS3Client(m.ctx)
	s3Bucker := "mfcore-data"
	path := fmt.Sprintf("dev/temp/gobasetest/%v.pdf", uuid.NewString())
	err = s3Client.PutObject(s3Bucker, path, bytes.NewReader(responseBody), res.Header["Content-Type"][0])
	if err != nil {
		return 500, nil, err
	}
	return 200, map[string]interface{}{
		"uploadPath": path,
	}, nil

}

func (m *Manager) UploadPII(body interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	m.log.Info("Body", body)
	req, ok := body.(*Upload)
	if !ok {
		panic(utils.NewHTTPBadRequestError("Error in reading payload", body))
	}
	apiClient := http.NewHTTPClient(m.ctx)
	res, responseBody, err := apiClient.Get(req.URL, nil, nil, 10)
	if err != nil {
		return 500, nil, err
	}
	var path string
	err = xray.Capture(m.ctx, "HandlerUploadPII", func(ctx1 context.Context) error {
		kms := "arn:aws:kms:ap-south-1:490302598154:key/489a37f4-4ef0-408f-9ecf-7dd629505060"
		s3Client, err := aws.GetDefaultS3PIIClient(m.ctx, kms)
		if err != nil {
			return err
		}
		s3Bucker := "mfcore-data"
		path = fmt.Sprintf("dev/temp/gobasetest/%v.pdf", uuid.NewString())
		return s3Client.PutObject(s3Bucker, path, bytes.NewReader(responseBody), res.Header["Content-Type"][0])
	})
	if err != nil {
		return 500, nil, err
	}

	return 200, map[string]interface{}{
		"uploadPath": path,
	}, nil

}

func (m *Manager) DownloadPII(body interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	m.log.Info("Body", body)
	req, ok := body.(*Upload)
	if !ok {
		panic(utils.NewHTTPBadRequestError("Error in reading payload", body))
	}
	kms := "arn:aws:kms:ap-south-1:490302598154:key/489a37f4-4ef0-408f-9ecf-7dd629505060"
	s3Client, err := aws.GetDefaultS3PIIClient(m.ctx, kms)
	if err != nil {
		return 500, nil, err
	}
	s3Bucker := "mfcore-data"
	piiCache, err := s3Client.GetFileCache(s3Bucker, req.FileName, "gopiitest")
	if err != nil {
		return 500, nil, err
	}
	return 200, piiCache, nil

}

func (m *Manager) Download(body interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	m.log.Info("Body", body)
	req, ok := body.(*Upload)
	if !ok {
		panic(utils.NewHTTPBadRequestError("Error in reading payload", body))
	}
	s3Client := aws.GetDefaultS3Client(m.ctx)
	s3Bucker := "mfcore-data"
	piiCache, err := s3Client.CreatePresignedURLGET(s3Bucker, req.FileName, 60*15)
	if err != nil {
		return 500, nil, err
	}
	return 200, piiCache, nil

}

func (m *Manager) StepFuncInvocation(queryParam interface{}, pathParam interface{}, jsonBody string, queryParams interface{}) (int, interface{}, error) {
	sfnPayload := map[string]interface{}{
		"isLambdaInvocation": true,
		"actionName":         "ACTION_1",
		"payload": map[string]string{
			"hello": "world",
		},
	}
	sfnClient := aws.GetDefaultSFNClient(m.ctx)
	functionName := fmt.Sprintf("bedrockGobaseTest-%s-TestSFNGo", utils.Getenv("stage", ""))
	stateMachineArn := fmt.Sprintf("arn:aws:states:%s:%s:stateMachine:%s", utils.Getenv("region", ""), utils.Getenv("account_id", ""), functionName)
	executionName := fmt.Sprintf("test-exec-%s", uuid.NewString())
	err := sfnClient.StartExecution(stateMachineArn, executionName, sfnPayload)
	return 200, map[string]interface{}{
		"hello": "world",
	}, err
}

func (m *Manager) TransactionRejected(payload interface{}) error {
	m.log.Info("FuncPayloadRejected", payload)
	return nil
}

func (m *Manager) TransactionConfirmed(payload interface{}) error {
	m.log.Info("FuncPayloadConfirmed", payload)

	return nil
}

func (m *Manager) ActionOne(payload interface{}) (int, interface{}, error) {
	m.log.Info("LambdaInvocationActionOne", payload)
	return 200, map[string]interface{}{
		"isLambdaInvocation": true,
		"actionName":         "ACTION_2",
		"payload": map[string]string{
			"ACTION": "TWO",
		},
	}, nil
}

func (m *Manager) ActionTwo(payload interface{}) (int, interface{}, error) {
	m.log.Info("LambdaInvocationActionTwo", payload)
	return 200, map[string]string{"ACTION": "TWO"}, nil
}

func (m *Manager) CronActionOne(payload interface{}) (int, interface{}, error) {
	m.log.Info("CronActionOne", payload)
	return 200, map[string]string{"ACTION": "ONE"}, nil
}

func (m *Manager) CronActionTwo(payload interface{}) (int, interface{}, error) {
	m.log.Info("CronActionTwo", payload)
	return 200, map[string]string{"ACTION": "TWO"}, nil
}

func (m *Manager) GetS3EventHandler() map[string]map[string]map[string]*eventprocessor.S3Trigger {
	s3Map := &eventprocessor.S3Trigger{
		KeyPrefix:        "",
		S3TriggerHandler: m.TransactionConfirmed, //Update respective method
	}
	newMap := map[string]map[string]map[string]*eventprocessor.S3Trigger{
		utils.GetenvMust("s3_bucket_name"): {
			"ObjectCreated:Put": {"/temp": s3Map},
		},
	}
	return newMap
}

func NewManager(ctx context.Context, logger *log.Log, event interface{}, eventType eventprocessor.EventType) eventprocessor.EventProcessor {
	var obj eventprocessor.EventProcessor = &Manager{log: logger, ctx: ctx, event: event, eventType: eventType}
	return obj
}

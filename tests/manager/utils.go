package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"gobase-lambda/aws"
)

func uploadToS3(t *testing.T) string {
	s3Client := aws.GetDefaultS3Client(context.TODO())
	path := fmt.Sprintf("dev/temp/gobasetest/%v.pdf", uuid.NewString())
	s3Bucker := "mfcore-data"
	err := s3Client.PutFile(s3Bucker, path, "../testfile/sample_aadhaar.pdf")
	if err != nil {
		t.Fatal(err)
	}
	url, err := s3Client.CreatePresignedURLGET(s3Bucker, path, 10*60)
	if err != nil {
		t.Fatal(err)
	}
	return *url
}

func loadBody(t *testing.T, res *events.APIGatewayProxyResponse, body interface{}) {
	err := json.Unmarshal([]byte(res.Body), body)
	if err != nil {
		t.Fatal(err)
	}
}

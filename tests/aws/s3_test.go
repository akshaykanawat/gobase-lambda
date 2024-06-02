package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gobase-lambda/aws"
)

func TestS3(t *testing.T) {
	s3Client := aws.GetDefaultS3Client(context.TODO())
	path := fmt.Sprintf("dev/temp/gobasetest/%v.pdf", uuid.NewString())
	s3Bucker := "mfcore-data"
	err := s3Client.PutFile(s3Bucker, path, "../testfile/sample_aadhaar.pdf")
	if err != nil {
		t.Fatal(err)
	}
	err = s3Client.GetFile(s3Bucker, path, "../testfile/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s3Client.CreatePresignedURLGET(s3Bucker, path, 10*60)
	if err != nil {
		t.Fatal(err)
	}
	path = fmt.Sprintf("dev/temp/gobasetest/%v.pdf", uuid.NewString())
	_, err = s3Client.CreatePresignedURLPUT(s3Bucker, path, 10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestS3PII(t *testing.T) {
	keyArn := "arn:aws:kms:ap-south-1:490302598154:key/489a37f4-4ef0-408f-9ecf-7dd629505060"
	s3Client, err := aws.GetDefaultS3PIIClient(context.TODO(), keyArn)
	if err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("dev/temp/gobasetest/%v.pdf", uuid.NewString())
	s3Bucker := "mfcore-data"
	err = s3Client.PutFile(s3Bucker, path, "/home/user/Desktop/output.jpg")
	if err != nil {
		t.Fatal(err)
	}
	err = s3Client.GetFile(s3Bucker, path, "../testfile/testpii.pdf")
	if err != nil {
		t.Fatal(err)
	}
	// _, err = s3Client.GetFileCache(s3Bucker, path, "testCache")
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

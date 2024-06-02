package tests

import (
	"context"
	"testing"

	"gobase-lambda/aws"
)

func TestAWSKMS(t *testing.T) {
	kms := aws.GetDefaultKMSClient(context.TODO(), "arn:aws:kms:ap-south-1:490302598154:key/489a37f4-4ef0-408f-9ecf-7dd629505060")
	text := "asfasdfsaf"
	_, encryptedText, err := kms.Encrypt(&text)
	if err != nil {
		t.Fatal(err)
	}
	plainText, err := kms.Decrypt(&encryptedText)
	if err != nil {
		t.Fatal(err)
	}
	if plainText != text {
		t.Fatalf("Texts are not matching %v, %v", text, plainText)
	}
}

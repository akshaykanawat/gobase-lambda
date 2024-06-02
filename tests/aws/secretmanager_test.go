package tests

import (
	"context"
	"testing"

	"gobase-lambda/aws"
)

func TestSecretManager(t *testing.T) {
	client := aws.GetDefaultSecretManagerClient(context.TODO())
	_, err := client.GetSecret("arn:aws:secretsmanager:ap-south-1:490302598154:secret:dev/signzy-XnLNDf")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetSecret("arn:aws:secretsmanager:ap-south-1:490302598154:secret:dev/signzy-XnLNDf")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.GetSecretNonCache("arn:aws:secretsmanager:ap-south-1:490302598154:secret:dev/signzy-XnLNDf")
	if err != nil {
		t.Fatal(err)
	}
}

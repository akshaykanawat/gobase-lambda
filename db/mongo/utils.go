package mongo

import (
	"fmt"

	"gobase-lambda/aws"
)

func GetMongoURI(url, secretArn string, secretClient *aws.SecretManager) (string, error) {
	secretData, err := secretClient.GetSecret(secretArn)
	if err != nil {
		return "", err
	}
	mongoURI := fmt.Sprintf(url, secretData["username"], secretData["password"])
	return mongoURI, nil
}

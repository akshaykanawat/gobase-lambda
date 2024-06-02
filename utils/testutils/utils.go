package testutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	base "gobase-lambda/aws"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func setEnv(path string) {
	rawData, _ := ioutil.ReadFile(path)
	var envData map[string]string
	err := json.Unmarshal(rawData, &envData)
	if err != nil {
		wd, _ := os.Getwd()
		fmt.Println(wd)
		panic(err)
	}
	for key, value := range envData {
		os.Setenv(key, value)
	}

}

func getSTSToken() map[string]string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	filename := fmt.Sprintf("%v/.aws/sts.json", dirname)
	rawData, _ := ioutil.ReadFile(filename)
	var envData map[string]interface{}
	err = json.Unmarshal(rawData, &envData)
	if err != nil {
		panic(err)
	}
	rawCredentials := envData["Credentials"].(map[string]interface{})
	credentials := make(map[string]string)
	for key, value := range rawCredentials {
		credentials[key] = value.(string)
	}
	return credentials
}

func setAWSSession() {
	stsToken := getSTSToken()
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("region")),
		Credentials: credentials.NewStaticCredentials(stsToken["AccessKeyId"], stsToken["SecretAccessKey"], stsToken["SessionToken"]),
	}))
	base.SetDefaultAWSSession(awsSession)
}

func Initialize(path string) {
	os.Setenv("AWS_XRAY_SDK_DISABLED", "TRUE")
	setEnv(path)
	getSTSToken()
	setAWSSession()
}

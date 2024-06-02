package mongo

import (
	"gobase-lambda/aws"
	"gobase-lambda/log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/fatih/structs"
)

type awsKMSCredentials struct {
	AccessKeyID     string `structs:"accessKeyId"`
	SecretAccessKey string `structs:"secretAccessKey"`
	SessionToken    string `structs:"sessionToken"`
}

// The data key options used for this KMS provider.
// See https://github.com/mongodb/specifications/blob/master/source/client-side-encryption/client-side-encryption.rst#datakeyopts

type awsKMSDataKeyOpts struct {
	Region   string `bson:"region"`
	KeyARN   string `bson:"key"`                // the aws key arn
	Endpoint string `bson:"endpoint,omitempty"` // optional, defaults to kms.<region>.amazonaws.com
}

// KMSProvider holds the credentials and master key information to use this KMS
// Get an instance of this with the kms.AWSProvider() method
type KMSProvider struct {
	credentials awsKMSCredentials
	dataKeyOpts awsKMSDataKeyOpts
	name        string
}

// AWSProvider reads in the required environment variables for credentials and master key
// location to use AWS as a KMS

func GetAWSProvider(awsKeyARN string, session *session.Session, log *log.Log) (provider *KMSProvider, err error) {
	cred, err := session.Config.Credentials.Get()
	if err != nil {
		log.Error("AWS credential fetch", err)
		return
	}
	provider = AWSProvider(cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken, awsKeyARN, *session.Config.Region)
	log.Debug("Mongo AWS KMS provider", provider)
	return
}

func GetDefaultAWSProvider(awsKeyARN string) (*KMSProvider, error) {
	session := aws.GetDefaultAWSSession()
	return GetAWSProvider(awsKeyARN, session, log.GetDefaultLogger())
}

func AWSProvider(awsAccessKeyID, awsSecretAccessKey, sessionToken, awsKeyARN, awsKeyRegion string) *KMSProvider {

	return &KMSProvider{
		credentials: awsKMSCredentials{
			AccessKeyID:     awsAccessKeyID,
			SecretAccessKey: awsSecretAccessKey,
			SessionToken:    sessionToken,
		},

		dataKeyOpts: awsKMSDataKeyOpts{
			KeyARN: awsKeyARN,
			Region: awsKeyRegion,
		},
		name: "aws",
	}
}

// Name is the name of this provider
func (a *KMSProvider) Name() string {
	return a.name
}

// Credentials are the credentials for this provider returned in the format necessary
// to immediately pass to the driver
func (a *KMSProvider) Credentials() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{"aws": structs.Map(a.credentials)}
}

// DataKeyOpts are the data key options for this provider returned in the format necessary
// to immediately pass to the driver
func (a *KMSProvider) DataKeyOpts() interface{} {
	return a.dataKeyOpts
}

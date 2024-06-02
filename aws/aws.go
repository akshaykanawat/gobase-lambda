package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var defaultAWSSession *session.Session

func SetDefaultAWSSession(sess *session.Session) {
	xray.AWSSession(sess)
	defaultAWSSession = sess
}

func GetDefaultAWSSession() *session.Session {
	return defaultAWSSession
}

func GetRegionalDefaultAWSSession(region string) *session.Session {
	return GetRegionalAWSSession(defaultAWSSession, region)
}

func GetRegionalAWSSession(awsSession *session.Session, region string) *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: awsSession.Config.Credentials,
	}))
	xray.AWSSession(sess)
	return sess
}

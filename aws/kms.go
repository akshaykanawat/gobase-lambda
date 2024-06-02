package aws

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"gobase-lambda/log"
)

type KMS struct {
	_      struct{}
	Client *kms.KMS
	keyArn *string
	log    *log.Log
	ctx    context.Context
}

var defaultKMSClient *kms.KMS

func GetAWSKMSClient(awsSession *session.Session) *kms.KMS {
	client := kms.New(awsSession)
	return client
}

func GetDefaultKMSClient(ctx context.Context, keyArn string) *KMS {
	if defaultKMSClient == nil {
		defaultKMSClient = GetAWSKMSClient(defaultAWSSession)
	}
	return GetKMSClient(ctx, defaultKMSClient, keyArn)
}

func GetKMSClient(ctx context.Context, client *kms.KMS, keyArn string) *KMS {
	return &KMS{Client: client, keyArn: &keyArn, log: log.GetDefaultLogger(), ctx: ctx}
}

func (k *KMS) Encrypt(plainText *string) (cipherTextBlob []byte, b64EncodedText string, err error) {
	req := &kms.EncryptInput{
		KeyId:     k.keyArn,
		Plaintext: []byte(*plainText),
	}
	k.log.Debug("KMS encryption request", fmt.Sprint(req))
	res, err := k.Client.EncryptWithContext(k.ctx, req)
	if err != nil {
		k.log.Error("KMS encryption error", err)
		return
	}
	k.log.Debug("KMS encryption response", res)
	cipherTextBlob = res.CiphertextBlob
	b64EncodedText = base64.StdEncoding.EncodeToString(cipherTextBlob)
	return
}

func (k *KMS) Decrypt(b64EncodedText *string) (plainText string, err error) {
	data, err := base64.StdEncoding.DecodeString(*b64EncodedText)
	if err != nil {
		return
	}
	req := &kms.DecryptInput{
		KeyId:          k.keyArn,
		CiphertextBlob: []byte(data),
	}
	k.log.Debug("KMS decryption request", req)
	res, err := k.Client.DecryptWithContext(k.ctx, req)
	if err != nil {
		k.log.Error("KMS decryption error", err)
		return
	}
	k.log.Debug("KMS decryption response", fmt.Sprint(res))
	plainText = string(res.Plaintext)
	return
}

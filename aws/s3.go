package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"gobase-lambda/log"
)

type S3 struct {
	_      struct{}
	Client *s3.S3
	log    *log.Log
	ctx    context.Context
}

var defaultS3Client *s3.S3

func GetAWSS3Client(awsSession *session.Session) *s3.S3 {
	return s3.New(awsSession)
}

func GetDefaultS3Client(ctx context.Context) *S3 {
	if defaultS3Client == nil {
		defaultS3Client = GetAWSS3Client(defaultAWSSession)
	}
	return GetS3Client(ctx, defaultS3Client)
}

func GetS3Client(ctx context.Context, client *s3.S3) *S3 {
	return &S3{Client: client, log: log.GetDefaultLogger(), ctx: ctx}
}

func (s *S3) PutObject(s3Bucket, s3Key string, body io.ReadSeeker, mimeType string) error {
	req := &s3.PutObjectInput{Bucket: &s3Bucket, Key: &s3Key, Body: body, ContentType: &mimeType}
	s.log.Debug("S3 put object request", req)
	res, err := s.Client.PutObjectWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("S3 put object error", err)
		return err
	}
	s.log.Debug("S3 put object response", res)
	return nil
}

func (s *S3) PutFile(s3Bucket, s3Key, localFilPath string) error {
	fp, err := os.Open(localFilPath)
	if err != nil {
		s.log.Error("Error opening file", localFilPath)
		return err
	}
	mime, err := mimetype.DetectFile(localFilPath)
	if err != nil {
		s.log.Error("Error detecting mime type", err)
	}
	s.log.Debug("File mimetype", mime)
	defer fp.Close()
	return s.PutObject(s3Bucket, s3Key, fp, mime.String())
}

func (s *S3) GetObject(s3Bucket, s3Key string) ([]byte, error) {
	req := &s3.GetObjectInput{Bucket: &s3Bucket, Key: &s3Key}
	s.log.Debug("S3 get object request", req)
	res, err := s.Client.GetObjectWithContext(s.ctx, req)
	if err != nil {
		s.log.Error("S3 get object error", err)
		return nil, err
	}
	s.log.Debug("S3 get object response", res)
	blob, err := io.ReadAll(res.Body)
	if err != nil {
		s.log.Error("S3 get object read error", err)
		return nil, err
	}
	return blob, nil
}

func (s *S3) GetFile(s3Bucket, s3Key, localFilePath string) error {
	blob, err := s.GetObject(s3Bucket, s3Key)
	if err != nil {
		return err
	}
	fp, err := os.Create(localFilePath)
	if err != nil {
		s.log.Error("S3 get file - file creation error", err)
		return err
	}
	defer fp.Close()
	n, err := fp.Write(blob)
	if err != nil {
		s.log.Error("S3 get file - file writing error", err)
		return err
	}
	if n != len(blob) {
		err := fmt.Errorf("total bytes %v, written bytes %v", len(blob), n)
		s.log.Error("S3 get file - file writing error", err)
		return err
	}
	return nil
}

func (s *S3) CreatePresignedURLGET(s3Bucket, s3Key string, expireTimeInSeconds int) (*string, error) {
	req, _ := s.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &s3Bucket,
		Key:    &s3Key,
	})
	urlStr, err := req.Presign(time.Duration(expireTimeInSeconds) * time.Second)
	if err != nil {
		s.log.Error("S3 failed to sign GET request", err)
		return nil, err
	}
	s.log.Debug("S3 presigned GET url", urlStr)
	return &urlStr, nil
}

func (s *S3) CreatePresignedURLPUT(s3Bucket, s3Key string, expireTimeInSeconds int) (*string, error) {
	req, _ := s.Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: &s3Bucket,
		Key:    &s3Key,
	})
	urlStr, err := req.Presign(time.Duration(expireTimeInSeconds) * time.Second)
	if err != nil {
		s.log.Error("S3 failed to sign PUT request", err)
		return nil, err
	}
	s.log.Debug("S3 presigned PUT url", urlStr)
	return &urlStr, nil
}

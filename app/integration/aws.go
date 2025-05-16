package integration

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"go-aws-s3-bucket/app/helper"
	"io"
	"path/filepath"
)

type AWS struct {
	BucketName  string
	Region      string
	S3URLPrefix string
	AccessKey   string
	SecretKey   string
}

type awsConfigChecklist struct {
	Key   string
	Valid bool
}

func NewAWSInstance(v *viper.Viper) (*AWS, error) {
	awsConfig := &AWS{
		BucketName:  v.GetString("AWS_BUCKET"),
		Region:      v.GetString("AWS_DEFAULT_REGION"),
		S3URLPrefix: v.GetString("AWS_URL_API"),
		AccessKey:   v.GetString("AWS_ACCESS_KEY"),
		SecretKey:   v.GetString("AWS_SECRET_KEY"),
	}

	checks := []awsConfigChecklist{
		{"AWS_BUCKET", awsConfig.BucketName != ""},
		{"AWS_DEFAULT_REGION", awsConfig.Region != ""},
		{"AWS_URL_API", awsConfig.S3URLPrefix != ""},
		{"AWS_ACCESS_KEY", awsConfig.AccessKey != ""},
		{"AWS_SECRET_KEY", awsConfig.SecretKey != ""},
	}

	for _, check := range checks {
		if !check.Valid {
			return nil, fmt.Errorf("missing or invalid required AWS configuration: %s", check.Key)
		}
	}

	return awsConfig, nil
}

func (h *AWS) ListObjects() (*[]s3.Object, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	}))
	svc := s3.New(sess)

	resp, err := svc.ListObjectsV2WithContext(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(h.BucketName),
	})
	if err != nil {
		return nil, err
	}

	var list []s3.Object

	for _, obj := range resp.Contents {
		list = append(list, *obj)
	}

	return &list, nil
}

func (h *AWS) Upload(apiCallID, folder, filename string, fileData []byte) (string, error) {
	path := filepath.Join(folder, helper.GenerateUniqueFilename()+filepath.Ext(filename))

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	}))
	svc := s3.New(sess)

	params := &s3.PutObjectInput{
		Bucket:      aws.String(h.BucketName),
		Key:         aws.String(path),
		Body:        bytes.NewReader(fileData),
		ACL:         aws.String(s3.BucketCannedACLPublicRead),
		ContentType: aws.String(filepath.Ext(filename)),
	}

	_, err := svc.PutObject(params)
	if err != nil {
		return "", err
	}

	helper.LogInfo(apiCallID, "Uploaded file successfully: "+path)

	return path, nil
}

func (h *AWS) Download(apiCallID, filePath string) (io.ReadCloser, string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	}))
	svc := s3.New(sess)

	output, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(h.BucketName),
		Key:    aws.String(filePath),
	})

	if err != nil {
		return nil, "", err
	}

	helper.LogInfo(apiCallID, "Downloaded file successfully")

	return output.Body, *output.ContentType, nil
}

package integration

import (
	"errors"
	"os"
)

type AWS struct {
	bucketName  string
	region      string
	S3URLPrefix string
}

func NewAWSInstance() (*AWS, error) {
	bucketName := os.Getenv("AWS_BUCKET")
	region := os.Getenv("AWS_DEFAULT_REGION")
	urlPrefix := os.Getenv("AWS_URL_API")

	if bucketName == "" || region == "" || urlPrefix == "" {
		return nil, errors.New("missing required AWS environment variables")
	}

	return &AWS{
		bucketName:  bucketName,
		region:      region,
		S3URLPrefix: urlPrefix,
	}, nil
}

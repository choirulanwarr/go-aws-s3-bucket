package integration

import (
	"bytes"
	"context"
	"fmt"
	"go-aws-s3-bucket/app/helper"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
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
	// Create AWS session with proper error handling
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize S3 client
	svc := s3.New(sess)

	// Prepare list objects parameters
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(h.BucketName),
	}

	// List objects from S3 with context
	resp, err := svc.ListObjectsV2WithContext(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects from S3: %w", err)
	}

	// Initialize result slice with capacity
	list := make([]s3.Object, 0, len(resp.Contents))

	// Append objects to result slice
	for _, obj := range resp.Contents {
		list = append(list, *obj)
	}

	return &list, nil
}

func (h *AWS) Upload(apiCallID, folder, filename string, fileData []byte) (string, error) {
	// Generate unique file path
	path := filepath.Join(folder, helper.GenerateUniqueFilename()+filepath.Ext(filename))

	// Create AWS session with proper error handling
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize S3 client
	svc := s3.New(sess)

	// Get content type from file extension
	contentType := helper.DefaultMIME(mime.TypeByExtension(filepath.Ext(filename)))

	// Prepare upload parameters
	params := &s3.PutObjectInput{
		Bucket:      aws.String(h.BucketName),
		Key:         aws.String(path),
		Body:        bytes.NewReader(fileData),
		ACL:         aws.String(s3.BucketCannedACLPublicRead),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"original-filename": aws.String(filename),
			"upload-timestamp":  aws.String(time.Now().UTC().Format(time.RFC3339)),
		},
	}

	// Upload file to S3
	_, err = svc.PutObjectWithContext(context.Background(), params)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Log successful upload
	helper.LogInfo(apiCallID, fmt.Sprintf("Successfully uploaded file to S3: %s (original: %s)", path, filename))

	return path, nil
}

func (h *AWS) GeneratePresignedURL(apiCallID, filePath string, expiration time.Duration) (string, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize S3 client
	svc := s3.New(sess)

	// Create GetObject request
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(h.BucketName),
		Key:    aws.String(filePath),
	})

	// Generate presigned URL
	urlStr, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	helper.LogInfo(apiCallID, fmt.Sprintf("Generated presigned URL for: %s (expires in %s)", filePath, expiration))

	return urlStr, nil
}

func (h *AWS) Download(apiCallID, filePath string) (io.ReadCloser, string, error) {
	// Create AWS session with proper error handling
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize S3 client
	svc := s3.New(sess)

	// Prepare download parameters
	params := &s3.GetObjectInput{
		Bucket: aws.String(h.BucketName),
		Key:    aws.String(filePath),
	}

	// Download file from S3
	output, err := svc.GetObjectWithContext(context.Background(), params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file from S3: %w", err)
	}

	// Log successful download
	helper.LogInfo(apiCallID, fmt.Sprintf("Successfully downloaded file from S3: %s", filePath))

	return output.Body, *output.ContentType, nil
}

func (h *AWS) MoveObject(apiCallID, sourcePath, destPath string) error {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(h.Region),
		Credentials: credentials.NewStaticCredentials(
			h.AccessKey,
			h.SecretKey,
			"",
		),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize S3 client
	svc := s3.New(sess)

	// Step 1: Copy object from source to destination
	copyInput := &s3.CopyObjectInput{
		Bucket:     aws.String(h.BucketName),
		CopySource: aws.String(fmt.Sprintf("%s/%s", h.BucketName, sourcePath)),
		Key:        aws.String(destPath),
		ACL:        aws.String(s3.BucketCannedACLPublicRead),
	}

	_, err = svc.CopyObjectWithContext(context.Background(), copyInput)
	if err != nil {
		return fmt.Errorf("failed to copy object from %s to %s: %w", sourcePath, destPath, err)
	}

	helper.LogInfo(apiCallID, fmt.Sprintf("Copied object from %s to %s", sourcePath, destPath))

	// Step 2: Delete the source object
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(h.BucketName),
		Key:    aws.String(sourcePath),
	}

	_, err = svc.DeleteObjectWithContext(context.Background(), deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete source object %s after copy: %w", sourcePath, err)
	}

	helper.LogInfo(apiCallID, fmt.Sprintf("Deleted source object: %s", sourcePath))
	helper.LogInfo(apiCallID, fmt.Sprintf("Successfully moved object from %s to %s", sourcePath, destPath))

	return nil
}

package integration_test

import (
	"fmt"
	"github.com/spf13/viper"
	"go-aws-s3-bucket/app/integration"
	"os"
	"testing"
)

// hasAWSConfig checks if AWS environment variables are configured
func hasAWSConfig() bool {
	return os.Getenv("AWS_ACCESS_KEY") != "" &&
		os.Getenv("AWS_SECRET_KEY") != "" &&
		os.Getenv("AWS_BUCKET") != "" &&
		os.Getenv("AWS_DEFAULT_REGION") != "" &&
		os.Getenv("AWS_URL_API") != ""
}

func setupTestViper() *viper.Viper {
	v := viper.New()
	v.Set("AWS_ACCESS_KEY", os.Getenv("AWS_ACCESS_KEY"))
	v.Set("AWS_SECRET_KEY", os.Getenv("AWS_SECRET_KEY"))
	v.Set("AWS_BUCKET", os.Getenv("AWS_BUCKET"))
	v.Set("AWS_DEFAULT_REGION", os.Getenv("AWS_DEFAULT_REGION"))
	v.Set("AWS_URL_API", os.Getenv("AWS_URL_API"))
	v.Set("AWS_ACL", "public-read")
	return v
}

func TestNewAWSInstance_MissingConfig(t *testing.T) {
	v := viper.New()
	// Don't set any AWS values — should fail validation

	_, err := integration.NewAWSInstance(v)
	if err == nil {
		t.Error("Expected error for missing AWS config, got nil")
	}
}

func TestNewAWSInstance_MissingIndividualKeys(t *testing.T) {
	requiredKeys := []string{
		"AWS_BUCKET",
		"AWS_DEFAULT_REGION",
		"AWS_URL_API",
		"AWS_ACCESS_KEY",
		"AWS_SECRET_KEY",
	}

	for _, missingKey := range requiredKeys {
		t.Run(fmt.Sprintf("missing_%s", missingKey), func(t *testing.T) {
			v := viper.New()
			// Set all keys except the one being tested
			for _, key := range requiredKeys {
				if key != missingKey {
					v.Set(key, "test-value")
				}
			}

			_, err := integration.NewAWSInstance(v)
			if err == nil {
				t.Errorf("Expected error when %s is missing, got nil", missingKey)
			}
		})
	}
}

func TestNewAWSInstance_ValidConfig(t *testing.T) {
	if !hasAWSConfig() {
		t.Skip("Skipping integration test: AWS environment variables not configured")
	}

	v := setupTestViper()
	awsInstance, err := integration.NewAWSInstance(v)
	if err != nil {
		t.Fatalf("Failed to create AWS instance: %v", err)
	}

	if awsInstance.BucketName != os.Getenv("AWS_BUCKET") {
		t.Errorf("BucketName = %q; want %q", awsInstance.BucketName, os.Getenv("AWS_BUCKET"))
	}
	if awsInstance.Region != os.Getenv("AWS_DEFAULT_REGION") {
		t.Errorf("Region = %q; want %q", awsInstance.Region, os.Getenv("AWS_DEFAULT_REGION"))
	}
}

func TestListObjects(t *testing.T) {
	if !hasAWSConfig() {
		t.Skip("Skipping integration test: AWS environment variables not configured")
	}

	v := setupTestViper()
	awsInstance, err := integration.NewAWSInstance(v)
	if err != nil {
		t.Fatalf("Failed to create AWS instance: %v", err)
	}

	objects, err := awsInstance.ListObjects()
	if err != nil {
		t.Fatalf("ListObjects() failed: %v", err)
	}

	if objects == nil {
		t.Error("ListObjects() returned nil slice")
	}

	t.Logf("Found %d objects in bucket %s", len(*objects), awsInstance.BucketName)
}

func TestGeneratePresignedURL(t *testing.T) {
	if !hasAWSConfig() {
		t.Skip("Skipping integration test: AWS environment variables not configured")
	}

	v := setupTestViper()
	awsInstance, err := integration.NewAWSInstance(v)
	if err != nil {
		t.Fatalf("Failed to create AWS instance: %v", err)
	}

	// First list objects to find a file to generate presigned URL for
	objects, err := awsInstance.ListObjects()
	if err != nil || len(*objects) == 0 {
		t.Skip("No objects in bucket to test presigned URL")
	}

	testKey := *(*objects)[0].Key
	url, err := awsInstance.GeneratePresignedURL("TEST_API_CALL", testKey, 60*1e9) // 1 minute in nanoseconds
	if err != nil {
		t.Fatalf("GeneratePresignedURL() failed: %v", err)
	}

	if url == "" {
		t.Error("GeneratePresignedURL() returned empty URL")
	}

	t.Logf("Presigned URL for %s: %s", testKey, url[:min(len(url), 100)]+"...")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

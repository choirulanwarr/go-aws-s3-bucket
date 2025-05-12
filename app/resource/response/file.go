package response

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"go-aws-s3-bucket/app/helper"
	"time"
)

type GetFileResponse struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	LastModified string `json:"last_modified"`
}

func GetFileResponseFormatter(listFile *[]s3.Object) []GetFileResponse {
	var result []GetFileResponse

	for _, file := range *listFile {
		result = append(result, GetFileResponse{
			Name:         *file.Key,
			Size:         helper.FormatFileSize(*file.Size),
			LastModified: file.LastModified.Format(time.RFC3339),
		})
	}

	return result
}

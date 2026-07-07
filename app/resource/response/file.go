package response

import (
	"go-aws-s3-bucket/app/helper"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
)

type GetFileResponse struct {
	URL          string `json:"url"`
	Name         string `json:"name"`
	Size         string `json:"size"`
	LastModified string `json:"last_modified"`
}

func GetFileResponseFormatter(listFile *[]s3.Object, s3URLPrefix string) []GetFileResponse {
	var result []GetFileResponse

	for _, file := range *listFile {
		result = append(result, GetFileResponse{
			URL:          s3URLPrefix + "/" + *file.Key,
			Name:         *file.Key,
			Size:         helper.FormatFileSize(*file.Size),
			LastModified: file.LastModified.Format(time.RFC3339),
		})
	}

	return result
}

type UploadFileResponse struct {
	Path string `json:"path"`
}

type PresignedURLResponse struct {
	URL       string `json:"url"`
	Path      string `json:"path"`
	ExpiresAt string `json:"expires_at"`
}

type MoveFileResponse struct {
	SourcePath string `json:"source_path"`
	DestPath   string `json:"dest_path"`
}

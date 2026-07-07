package service

import (
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/helper"
	"go-aws-s3-bucket/app/integration"
	"go-aws-s3-bucket/app/resource/response"
	"io"
	"time"

	"github.com/spf13/viper"
)

type FileService struct {
	Viper *viper.Viper
}

func NewFileService(viper *viper.Viper) *FileService {
	return &FileService{
		viper,
	}
}

func (f *FileService) GetAllFile(apiCallID string) (*[]response.GetFileResponse, constant.ResponseMap) {
	awsConfig, err := integration.NewAWSInstance(f.Viper)
	if err != nil {
		helper.LogError(apiCallID, "Error creating AWS configuration: "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}
	listFile, err := awsConfig.ListObjects()
	if err != nil {
		helper.LogError(apiCallID, "Error list file : "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}

	formatted := response.GetFileResponseFormatter(listFile, awsConfig.S3URLPrefix)

	return &formatted, constant.Res200Get
}

func (f *FileService) UploadFile(apiCallID, folder, filename string, file []byte) (*response.UploadFileResponse, constant.ResponseMap) {
	gcs, err := integration.NewAWSInstance(f.Viper)
	if err != nil {
		helper.LogError(apiCallID, "Error creating AWS configuration: "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}
	uploadedPath, err := gcs.Upload(apiCallID, folder, filename, file)
	if err != nil {
		helper.LogError(apiCallID, "Error upload file : "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}

	return &response.UploadFileResponse{Path: uploadedPath}, constant.Res200Save

}

func (f *FileService) GeneratePresignedURL(apiCallID, filePath string, expiresInMinutes int) (*response.PresignedURLResponse, constant.ResponseMap) {
	awsConfig, err := integration.NewAWSInstance(f.Viper)
	if err != nil {
		helper.LogError(apiCallID, "Error creating AWS configuration: "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}

	// Default expiration to 15 minutes if not specified
	if expiresInMinutes <= 0 {
		expiresInMinutes = 15
	}
	expiration := time.Duration(expiresInMinutes) * time.Minute
	expiresAt := time.Now().UTC().Add(expiration)

	presignedURL, err := awsConfig.GeneratePresignedURL(apiCallID, filePath, expiration)
	if err != nil {
		helper.LogError(apiCallID, "Error generating presigned URL: "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}

	return &response.PresignedURLResponse{
		URL:       presignedURL,
		Path:      filePath,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, constant.Res200Get
}

func (f *FileService) MoveFile(apiCallID, sourcePath, destPath string) (*response.MoveFileResponse, constant.ResponseMap) {
	awsConfig, err := integration.NewAWSInstance(f.Viper)
	if err != nil {
		helper.LogError(apiCallID, "Error creating AWS configuration: "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}

	// Validate source and destination are different
	if sourcePath == destPath {
		helper.LogError(apiCallID, "Source and destination paths are the same")
		return nil, constant.Res400InvalidPayload
	}

	err = awsConfig.MoveObject(apiCallID, sourcePath, destPath)
	if err != nil {
		helper.LogError(apiCallID, "Error moving file: "+err.Error())
		return nil, constant.Res422SomethingWentWrong
	}

	return &response.MoveFileResponse{
		SourcePath: sourcePath,
		DestPath:   destPath,
	}, constant.Res200Save
}

func (f *FileService) DownloadFile(apiCallID, filePath string) (io.ReadCloser, string, constant.ResponseMap) {
	gcs, err := integration.NewAWSInstance(f.Viper)
	if err != nil {
		helper.LogError(apiCallID, "Error creating AWS configuration: "+err.Error())
		return nil, "", constant.Res422SomethingWentWrong
	}
	fileStream, contentType, err := gcs.Download(apiCallID, filePath)
	if err != nil {
		helper.LogError(apiCallID, "Error upload file : "+err.Error())
		return nil, "", constant.Res422SomethingWentWrong
	}

	return fileStream, contentType, constant.Res200Get
}

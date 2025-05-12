package service

import (
	"github.com/spf13/viper"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/helper"
	"go-aws-s3-bucket/app/integration"
	"go-aws-s3-bucket/app/resource/response"
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

	formatted := response.GetFileResponseFormatter(listFile)

	return &formatted, constant.Res200Get
}

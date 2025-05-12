package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/helper"
	"go-aws-s3-bucket/app/service"
)

type FileHandler struct {
	Service   *service.FileService
	Validator *validator.Validate
}

func NewFileHandler(service *service.FileService, validator *validator.Validate) *FileHandler {
	return &FileHandler{
		service,
		validator,
	}
}

func (f *FileHandler) GetAllFile(ctx *gin.Context) {
	apiCallID := ctx.GetString(constant.RequestIDKey)
	result, response := f.Service.GetAllFile(apiCallID)
	helper.ResponseAPI(ctx, response, result)
}

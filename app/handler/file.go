package handler

import (
	"github.com/go-playground/validator/v10"
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

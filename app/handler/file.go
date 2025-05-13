package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/helper"
	"go-aws-s3-bucket/app/resource/request"
	"go-aws-s3-bucket/app/service"
	"io"
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

func (f *FileHandler) UploadFile(ctx *gin.Context) {
	apiCallID := ctx.GetString(constant.RequestIDKey)

	var req request.UploadFileRequest
	if err := ctx.ShouldBind(&req); err != nil {
		helper.LogError(apiCallID, "Failed to bind request: "+err.Error())
		helper.ResponseAPI(ctx, constant.Res400InvalidPayload)
		return
	}

	if err := f.Validator.Struct(req); err != nil {
		helper.LogError(apiCallID, "Payload validation failed: "+err.Error())
		formattedErrors := helper.ErrorValidationFormatter(err.(validator.ValidationErrors))
		helper.ResponseAPI(ctx, constant.Res400InvalidPayload, formattedErrors)
		return
	}

	formFile, err := ctx.FormFile("file")
	if err != nil {
		helper.LogError(apiCallID, "Failed to retrieve uploaded file: "+err.Error())
		helper.ResponseAPI(ctx, constant.Res400InvalidPayload)
		return
	}

	uploadedFile, err := formFile.Open()
	if err != nil {
		helper.LogError(apiCallID, "Failed to open uploaded file: "+err.Error())
		helper.ResponseAPI(ctx, constant.Res400InvalidPayload)
		return
	}
	defer uploadedFile.Close()

	if !helper.IsAllowedFileType(apiCallID, uploadedFile) {
		helper.LogError(apiCallID, "Rejected file: unsupported content type")
		helper.ResponseAPI(ctx, constant.Res400InvalidPayload)
		return
	}

	fileBytes, err := io.ReadAll(uploadedFile)
	if err != nil {
		helper.LogError(apiCallID, "Failed to read uploaded file: "+err.Error())
		helper.ResponseAPI(ctx, constant.Res400InvalidPayload)
		return
	}

	result, response := f.Service.UploadFile(apiCallID, req.Folder, formFile.Filename, fileBytes)
	helper.ResponseAPI(ctx, response, result)
}

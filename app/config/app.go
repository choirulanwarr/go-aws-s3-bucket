package config

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go-aws-s3-bucket/app/handler"
	"go-aws-s3-bucket/app/router"
	"go-aws-s3-bucket/app/service"
	"gorm.io/gorm"
)

type AppConfig struct {
	Config    *viper.Viper
	DB        *gorm.DB
	Server    *gin.Engine
	Validator *validator.Validate
	Logger    *logrus.Logger
}

func InitConfig(config *AppConfig) {
	// File
	fileService := service.NewFileService(config.Config)
	fileHandler := handler.NewFileHandler(fileService, config.Validator)

	// Routers
	routeConfig := router.Config{
		Server:      config.Server,
		Config:      config.Config,
		DB:          config.DB,
		Logger:      config.Logger,
		FileHandler: fileHandler,
	}

	routeConfig.Init()
}

package service

import "github.com/spf13/viper"

type FileService struct {
	Viper *viper.Viper
}

func NewFileService(viper *viper.Viper) *FileService {
	return &FileService{
		viper,
	}
}

package router

func initGlobalRoutes(config *Config) {
	globalApi := config.Server.Group("/api/v1")

	// File
	globalApiFile := globalApi.Group("/")
	globalApiFile.GET("/list", config.FileHandler.GetAllFile)
	globalApiFile.POST("/upload", config.FileHandler.UploadFile)
	globalApiFile.GET("/download", config.FileHandler.DownloadFile)
}

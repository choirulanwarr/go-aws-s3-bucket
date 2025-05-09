package middleware

import (
	"github.com/gin-gonic/gin"
	"go-aws-s3-bucket/app/constant"
	"go-aws-s3-bucket/app/helper"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiID := helper.GenerateApiCallID()
		c.Set(constant.RequestIDKey, apiID)

		c.Next()
	}
}

package server

import (
	"github.com/gin-gonic/gin"
)

func InitControllers(r *gin.Engine) {
	r.NoRoute(NoRoute)
	r.GET("/api/:component/:channel/:os/:arch/json", GetUpdateJson)
	r.POST("/api/:component/:channel/:os/:arch/:version/uploadbinary", UploadBinary)

	r.GET("/", Index)
	r.GET("/files/*content", Files)
}

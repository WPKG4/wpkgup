package server

import (
	"github.com/gin-gonic/gin"
)

func InitControllers(r *gin.Engine) {
	r.NoRoute(NoRoute)
	r.GET("/ping", Ping)
	r.GET("/:component/:channel/:os/:arch/json", GetUpdateJson)
	r.POST("/:component/:channel/:os/:arch/:version/uploadbinary", UploadBinary)
}

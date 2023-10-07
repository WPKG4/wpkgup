package server

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StartServer(ip string, port int) {
	r := gin.Default()
	InitControllers(r)
	fmt.Println("Starting HTTP Server at http://" + ip + ":" + strconv.Itoa(port))
	r.Run(ip + ":" + strconv.Itoa(port))
}

func InitControllers(r *gin.Engine) {
	r.NoRoute(NoRoute)
	r.GET("/api/:component/:channel/:os/:arch/json", GetUpdateJson)
	r.GET("/api/:component/:channel/:os/:arch/:version/getbinary", GetBinary)
	r.POST("/api/:component/:channel/:os/:arch/:version/uploadbinary", UploadBinary)

	r.GET("/", Index)
	r.GET("/files/*content", Files)
	r.PUT("/api/keys/add", AddPublicKey)
}

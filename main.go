package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/config"
)

func main() {
	println("WpkgUp2", config.Version)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.Run(":8080")
}

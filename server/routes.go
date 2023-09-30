package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/utils"
)

func NoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 page not found"})
}

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func GetUpdateJson(c *gin.Context) {
	component := c.Param("component")
	channel := c.Param("channel")
	Os := c.Param("os")
	arch := c.Param("arch")

	path := filepath.Join(config.WorkDir, config.CONTENT_DIR, component, channel, Os, arch, "version.json")

	if !utils.FileExists(path) {
		c.JSON(404, gin.H{"error": "INVALID_COMPONENT"})
		return
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	json := string(buf)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, json)
}

func UploadBinary(c *gin.Context) {
	component := c.Param("component")
	channel := c.Param("channel")
	Os := c.Param("os")
	version := c.Param("version")
	arch := c.Param("arch")

	//Process path
	savepath := filepath.Join(config.WorkDir, config.CONTENT_DIR, component, channel, Os, arch, version)

	if err := os.MkdirAll(savepath, os.ModeSticky|os.ModePerm); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//Get file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//Save
	err = c.SaveUploadedFile(file, filepath.Join(savepath, file.Filename))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	checksum, err := utils.Sha256File(filepath.Join(savepath, file.Filename))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	jsonMap := VersionJson{
		Version:  version,
		Checksum: checksum,
		Url:      "/" + component + "/" + channel + "/" + Os + "/" + arch + "/" + version + "/version.json",
	}

	//Generate JSON
	err = GenerateJson(filepath.Join(config.WorkDir, config.CONTENT_DIR, component, channel, Os, arch, "version.json"), jsonMap)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(200)
}

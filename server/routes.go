package server

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/utils"
)

type Href struct {
	Href string
	Name string
}

func Index(c *gin.Context) {
	c.Redirect(http.StatusFound, "/files")
}

func Files(c *gin.Context) {
	log.SetPrefix("[FILE SERVER] ")

	path := c.Param("content")

	list := []Href{}

	file := filepath.Join(config.WorkDir, config.CONTENT_DIR, path)

	if utils.IsDir(file) {
		files, err := ioutil.ReadDir(file)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		for _, file := range files {
			list = append(list, Href{
				Href: filepath.Clean("/" + "files" + path + "/" + file.Name()),
				Name: file.Name(),
			})
		}

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, ProcessTemplate("file", FILES_TEMPLATE, gin.H{
			"path": path,
			"list": list,
		}))
	} else {
		f, err := os.Open(file)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer f.Close()

		mime, err := utils.GetMimeType(file)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		filesize, err := utils.FileSize(file)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", mime)
		c.Header("Content-Length", strconv.Itoa(int(filesize)))
		c.Stream(func(w io.Writer) bool {
			buffer := make([]byte, 1024)

			for {
				n, err := f.Read(buffer)
				if err != nil && err != io.EOF {
					log.Println("Error while reading file:", err)
					return false
				}

				if n > 0 {
					if _, err := w.Write(buffer[:n]); err != nil {
						log.Println("Error while sending data to client:", err)
						return false
					}
				}

				if err == io.EOF {
					break
				}
			}

			return false
		})
	}
}

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
	log.SetPrefix("[API] ")

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
	log.Println("Receiving new binary for component " + component + " | channel: " + channel + " version: " + version)

	err = c.SaveUploadedFile(file, filepath.Join(savepath, file.Filename))
	log.Println("Saving " + file.Filename + "...")
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

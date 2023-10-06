package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/crypto"
	"wpkg.dev/wpkgup/keystore"
	"wpkg.dev/wpkgup/utils"
)

type Href struct {
	Href string
	Name string
}

func Index(c *gin.Context) {
	c.Redirect(http.StatusFound, "/files")
}

func NoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 page not found"})
}

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func Files(c *gin.Context) {
	log.SetPrefix("[FILE SERVER] ")

	path := c.Param("content")

	var list []Href

	file := filepath.Join(config.WorkDir, config.ContentDir, path)

	if utils.IsDir(file) {
		files, err := os.ReadDir(file)
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
		c.String(http.StatusOK, ProcessTemplate("file", FilesTemplate, gin.H{
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

func GetUpdateJson(c *gin.Context) {
	component := c.Param("component")
	channel := c.Param("channel")
	Os := c.Param("os")
	arch := c.Param("arch")

	path := filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, "version.json")

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
	savePath := filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, version)

	if err := os.MkdirAll(savePath, os.ModeSticky|os.ModePerm); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	//Process multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	//Get file
	file := form.File["file"][0]
	sign := form.File["sign"][0]
	log.Println("Receiving new binary for component " + component + " | channel: " + channel + " | version: " + version)

	log.Println("Saving signature...")
	signaturePath := filepath.Join(savePath, "signature.der")
	err = c.SaveUploadedFile(sign, signaturePath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Println("Saving binary...")
	binaryPath := filepath.Join(savePath, file.Filename)
	err = c.SaveUploadedFile(file, binaryPath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	allKeys, err := keystore.GetAllKeys()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	verified := false
	for _, key := range allKeys {
		ecdsaKey, err := crypto.ParsePublicKeyFromString(key)
		if err != nil {
			log.Println("Error while parsing key:", err)
			continue
		}
		verifyResult, err := crypto.VerifyFromSignFile(ecdsaKey, binaryPath, signaturePath)
		if err != nil {
			log.Println("Error while verifying:", err)
			continue
		}

		if verifyResult {
			log.Println("Verified for ", key)
			verified = true
			break
		}
	}

	if verified {
		checksum, err := utils.Sha256File(filepath.Join(savePath, file.Filename))
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
		err = GenerateJson(filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, "version.json"), jsonMap)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusCreated)
	} else {
		log.Println("Sign is not valid, removing files...")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sign is not valid"})
	}
}

func AddPublicKey(c *gin.Context) {
	password := c.GetHeader("Password")
	key := c.GetHeader("Key")

	if password != config.LoadedConfig.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err := keystore.AddKey(key)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

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

func GetBinary(c *gin.Context) {
	log.SetPrefix("[API] ")

	component := c.Param("component")
	channel := c.Param("channel")
	Os := c.Param("os")
	version := c.Param("version")
	arch := c.Param("arch")

	var jsonMap VersionJson
	var err error

	log.Println("Requesting binary for component " + component + " | channel: " + channel + " | os: " + Os + " | arch: " + arch + " | version: " + version)
	//Process path
	if version == "latest" {
		log.Println("Version is latest")
		jsonMap, err = ReadVersionJson(filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, "version.json"))
	} else {
		jsonMap, err = ReadVersionJson(filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, version, "version.json"))
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	binaryPath := filepath.Join(config.WorkDir, config.ContentDir, jsonMap.Path)

	c.FileAttachment(filepath.Base(binaryPath), binaryPath)
}

func UploadBinary(c *gin.Context) {
	log.SetPrefix("[API] ")

	component := c.Param("component")
	channel := c.Param("channel")
	Os := c.Param("os")
	version := c.Param("version")
	arch := c.Param("arch")

	//Process path
	tempSavePath, err := os.MkdirTemp("", "wpkgup2_*")
	if err != nil {
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
	signaturePath := filepath.Join(tempSavePath, "signature.der")
	err = c.SaveUploadedFile(sign, signaturePath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Println("Saving binary...")
	binaryPath := filepath.Join(tempSavePath, file.Filename)
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
		//save to content dir
		savePath := filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, version)

		if err := os.MkdirAll(savePath, os.ModeSticky|os.ModePerm); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		//copy binary
		err := utils.CopyFile(binaryPath, filepath.Join(savePath, file.Filename))
		if err != nil {
			log.Println("Copy binary error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		//copy signature
		err = utils.CopyFile(signaturePath, filepath.Join(savePath, "signature.der"))
		if err != nil {
			log.Println("Copy signature error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		//generate checksum
		checksum, err := utils.Sha256File(filepath.Join(tempSavePath, file.Filename))
		if err != nil {
			log.Println("Checksum error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		jsonMap := VersionJson{
			Version:  version,
			Checksum: checksum,
			Path:     "/" + component + "/" + channel + "/" + Os + "/" + arch + "/" + version + "/" + file.Filename,
		}

		//Generate JSON
		err = GenerateVersionJson(filepath.Join(config.WorkDir, config.ContentDir, component, channel, Os, arch, "version.json"), jsonMap)
		if err != nil {
			log.Println("JSON generate error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		//Generate JSON in version folder
		err = GenerateVersionJson(filepath.Join(savePath, "version.json"), jsonMap)
		if err != nil {
			log.Println("JSON generate error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		//Removing temp files
		err = os.RemoveAll(tempSavePath)
		if err != nil {
			log.Println("Remove temp error:", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusCreated)
	} else {
		log.Println("Signature verification failed, removing files...")
		err := os.RemoveAll(tempSavePath)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Signature verification failed valid"})
		}
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

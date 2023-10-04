package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"wpkg.dev/wpkgup/utils"
)

var WorkDir string

func FindAppDataFolder(folderName string) string {
	var appDataFolder string

	switch Os := runtime.GOOS; Os {
	case "windows":
		appDataFolder = os.Getenv("APPDATA")
	case "darwin":
		appDataFolder = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default:
		appDataFolder = filepath.Join(os.Getenv("HOME"), ".config")
	}

	return filepath.Join(appDataFolder, folderName)
}

func InitDirs(workdir string) {
	if !utils.FileExists(workdir) {
		//create directories
		if err := os.MkdirAll(workdir, os.ModeSticky|os.ModePerm); err != nil {
			fmt.Println("Error creating directories:", err)
			return
		}
		//create directories
		if err := os.MkdirAll(filepath.Join(workdir, "content"), os.ModeSticky|os.ModePerm); err != nil {
			fmt.Println("Error creating directories:", err)
			return
		}
	}

	WorkDir = workdir
}

package server

import (
	"encoding/json"
	"os"
)

type VersionJson struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
	Url      string `json:"url"`
}

func GenerateJson(path string, jsonMap VersionJson) error {
	buf, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0664)
}

package server

import (
	"encoding/json"
	"os"
)

type VersionJson struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
	Path     string `json:"path"`
}

func GenerateVersionJson(path string, jsonMap VersionJson) error {
	buf, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0664)
}

func ReadVersionJson(path string) (VersionJson, error) {
	var jsonMap VersionJson

	buf, err := os.ReadFile(path)
	if err != nil {
		return jsonMap, err
	}

	err = json.Unmarshal(buf, &jsonMap)
	if err != nil {
		return jsonMap, err
	}
	return jsonMap, nil
}

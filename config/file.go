package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

var LoadedConfig Config

type Config struct {
	Password string
}

func Init() error {
	configFilePath := filepath.Join(WorkDir, ConfigFile)
	b, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	return toml.Unmarshal(b, &LoadedConfig)
}

func Parse(path string) (Config, error) {
	var ConfigSettings Config
	b, err := os.ReadFile(path)
	if err != nil {
		return ConfigSettings, err
	}

	toml.Unmarshal(b, &ConfigSettings)
	return ConfigSettings, nil
}

func Save(config Config, path string) error {
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, b, 0664)
	if err != nil {
		return err
	}
	return nil
}

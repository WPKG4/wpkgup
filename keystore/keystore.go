package keystore

import (
	"encoding/json"
	"os"
	"path/filepath"

	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/crypto"
	"wpkg.dev/wpkgup/utils"
)

var KeystorePath string

type AuthorizedKeys struct {
	Keys []string `json:"authorized_keys"`
}

func Init() error {
	KeystorePath = filepath.Join(config.WorkDir, config.KeystoreFile)
	if !utils.FileExists(KeystorePath) {
		keys := AuthorizedKeys{
			Keys: []string{},
		}
		return saveJson(keys, KeystorePath)
	}
	return nil
}

func saveJson(keys AuthorizedKeys, path string) error {
	b, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, b, 0664)
	if err != nil {
		return err
	}
	return nil
}

func readJson(path string) (AuthorizedKeys, error) {
	var authorizedKeys AuthorizedKeys

	buf, err := os.ReadFile(path)
	if err != nil {
		return authorizedKeys, err
	}

	err = json.Unmarshal(buf, &authorizedKeys)
	if err != nil {
		return authorizedKeys, err
	}
	return authorizedKeys, nil
}

func AddKey(key string) error {
	keys, err := readJson(KeystorePath)
	if err != nil {
		return err
	}
	keys.Keys = append(keys.Keys, key)

	_, err = crypto.ParsePublicKeyFromString(key)
	if err != nil {
		return err
	}

	err = saveJson(keys, KeystorePath)
	if err != nil {
		return err
	}
	return nil
}

func GetAllKeys() ([]string, error) {
	keys, err := readJson(KeystorePath)
	if err != nil {
		return nil, err
	}
	return keys.Keys, nil
}

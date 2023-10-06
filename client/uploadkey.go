package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/crypto"
)

func UploadKey(address, password string) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/keys/add", address), nil)
	if err != nil {
		return err
	}

	privateKey, err := crypto.ParsePublicKeyFromFile(filepath.Join(config.WorkDir, config.KeyringDir, "public.pem"))
	if err != nil {
		return err
	}
	privateKeyString, err := crypto.PublicKeyToBase64(privateKey)
	if err != nil {
		return err
	}

	req.Header.Set("Password", password)
	req.Header.Set("Key", privateKeyString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		var m map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&m)
		return fmt.Errorf("server response error: %s", m["error"])
	}

	return nil
}

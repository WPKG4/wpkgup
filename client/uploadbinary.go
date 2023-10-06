package client

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"wpkg.dev/wpkgup/crypto"
)

func addToForm(writer *multipart.Writer, name, filep string) error {
	file, err := os.Open(filep)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(name, file.Name())
	if err != nil {
		fmt.Println("Błąd tworzenia formularza pliku:", err)
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	return nil
}

func generateSign(privateKey *ecdsa.PrivateKey, filename, output string) error {
	signBuffer, err := crypto.Sign(privateKey, filename)
	if err != nil {
		return err
	}

	return os.WriteFile(output, signBuffer, 0664)
}

func UploadBinary(component, channel, Os, arch, version, address, filename string, privateKey *ecdsa.PrivateKey) error {

	temp, err := os.MkdirTemp("", "wpkgup2_*")
	if err != nil {
		return fmt.Errorf("mkdir temp error: %s", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	signPath := filepath.Join(temp, "sign.der")
	err = generateSign(privateKey, filename, signPath)
	if err != nil {
		return fmt.Errorf("sign error: %s", err)
	}

	//adding file
	addToForm(writer, "file", filename)
	addToForm(writer, "sign", signPath)

	writer.Close()

	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s/api/%s/%s/%s/%s/%s/uploadbinary", address, component, channel, Os, arch, version), &requestBody)
	if err != nil {

		return fmt.Errorf("http error: %s", err)
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {

		return fmt.Errorf("http request error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		var m map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&m)
		return fmt.Errorf("server response error: %s", m["error"])
	}

	return nil
}

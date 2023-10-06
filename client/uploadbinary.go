package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"wpkg.dev/wpkgup/config"
	"wpkg.dev/wpkgup/crypto"
)

func addToForm(writer *multipart.Writer, filename string) error {
	part, err := writer.CreateFormFile("file", "file.txt")
	if err != nil {
		fmt.Println("Błąd tworzenia formularza pliku:", err)
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	return nil
}

func generateSign(filename, output string) {
	privateKey, err := crypto.ParsePrivateKeyFromFile(filepath.Join(config.WorkDir, config.KeyringDir, "private.pem"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	signBuffer, err := crypto.Sign(privateKey, filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = os.WriteFile(output, signBuffer, 0664)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func UploadBinary(component, channel, Os, arch, version, address, filename string) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	addToForm(writer, filename)

	writer.Close()
}

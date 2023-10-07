package utils

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func ScanDefault(defaultInput string) string {
	var input string

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	if text := strings.TrimSpace(scanner.Text()); text != "" {
		input = text
	} else {
		input = defaultInput
	}
	return input
}

func ScanRequired() string {
	var input string

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	if text := strings.TrimSpace(scanner.Text()); text != "" {
		input = text
	} else {
		fmt.Print("Value must be set: ")
		return ScanRequired()
	}

	return input
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func GetMimeType(filePath string) (string, error) {
	mtype, err := mimetype.DetectFile(filePath)
	return mtype.String(), err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileSize(path string) (int64, error) {
	f, err := os.Stat(path)
	return f.Size(), err
}

func Sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func Sha256FileByte(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func Sha256(buff []byte) (string, error) {
	h := sha256.New()
	if _, err := h.Write(buff); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

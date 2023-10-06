package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func ParsePrivateKeyFromString(privateKeyStr string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}

	privateKey, err := x509.ParseECPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func ParsePublicKeyFromString(privateKeyStr string) (*ecdsa.PublicKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}

	publicKey, err := x509.ParsePKIXPublicKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("want ECDSA key")
	}

	return ecdsaPublicKey, nil
}

func ParsePrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	pemData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("invalid pem file")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func ParsePublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	pemData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("invalid pem file")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("want ECDSA key")
	}

	return ecdsaPublicKey, nil
}

func PrivateKeyToBase64(key *ecdsa.PrivateKey) (string, error) {
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", err
	}

	base64Key := base64.StdEncoding.EncodeToString(keyBytes)

	return base64Key, nil
}

func PublicKeyToBase64(key *ecdsa.PublicKey) (string, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	base64Key := base64.StdEncoding.EncodeToString(keyBytes)

	return base64Key, nil
}

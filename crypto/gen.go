package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	curve := elliptic.P256()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}

func SavePrivateKeyToFile(privateKey *ecdsa.PrivateKey, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	key, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	privateKeyPEM := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: key,
	}
	return pem.Encode(file, privateKeyPEM)
}

func SavePublicKeyToFile(publicKey *ecdsa.PublicKey, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return pem.Encode(file, publicKeyPEM)
}

func GeneratePublicFromPrivate(privateKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	return &privateKey.PublicKey
}

func GenKeys(privateKeyPath, publicKeyPath string) error {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		return err
	}

	err = SavePrivateKeyToFile(privateKey, privateKeyPath)
	if err != nil {
		return err
	}

	err = SavePublicKeyToFile(publicKey, publicKeyPath)
	if err != nil {
		return err
	}
	return nil
}

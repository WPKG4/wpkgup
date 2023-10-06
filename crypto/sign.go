package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"
	"os"

	"wpkg.dev/wpkgup/utils"
)

func Sign(privateKey *ecdsa.PrivateKey, filename string) ([]byte, error) {
	hash, err := utils.Sha256FileByte(filename)
	if err != nil {
		return nil, fmt.Errorf("hash generate error: %v", err)
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign error: %v", err)
	}

	return asn1.Marshal(struct {
		R, S *big.Int
	}{r, s})
}

func Verify(publicKey *ecdsa.PublicKey, filename string, signature []byte) (bool, error) {
	hash, err := utils.Sha256FileByte(filename)
	if err != nil {
		return false, fmt.Errorf("hash generate error: %v", err)
	}

	var sig struct {
		R, S *big.Int
	}
	_, err = asn1.Unmarshal(signature, &sig)
	if err != nil {
		return false, err
	}

	valid := ecdsa.Verify(publicKey, hash[:], sig.R, sig.S)
	return valid, nil
}

func VerifyFromSignFile(publicKey *ecdsa.PublicKey, filename string, signfile string) (bool, error) {
	buf, err := os.ReadFile(signfile)
	if err != nil {
		return false, fmt.Errorf("reading file error: %v", err)
	}
	return Verify(publicKey, filename, buf)
}

package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var (
	// SecretKey Define aes secret key 2^5
	SecretKey = []byte("2985BCFDB5FE43129843DB59825F8647")
)

func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	latest := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, latest...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unfading := int(origData[length-1])
	return origData[:(length - unfading)]
}

func Encrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypto := make([]byte, len(origData))
	blockMode.CryptBlocks(crypto, origData)
	return crypto, nil
}

func EncryptToBase64(origData, key []byte) (string, error) {
	crypto, err := Encrypt(origData, key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(crypto), nil
}

func Decrypt(crypto, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypto))
	blockMode.CryptBlocks(origData, crypto)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func DecryptFromBase64(data string, key []byte) ([]byte, error) {
	crypto, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return Decrypt(crypto, key)
}

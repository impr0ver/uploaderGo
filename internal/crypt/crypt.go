package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/mergermarket/go-pkcs7"
)

// Cipher key must be 32 chars long because block size is 16 bytes!
func getMD5Hash(text string) []byte {
	hash := md5.Sum([]byte(text))
	return hash[:]
}

// Encrypt plain text string into cipher text string
func AES256CBCEncode(plainText []byte, key string) ([]byte, error) {
	bKey := getMD5Hash(key)

	plainText, err := pkcs7.Pad(plainText, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("error in Pad function: %w", err)
	}
	if len(plainText)%aes.BlockSize != 0 {
		err := fmt.Errorf("plainText has the wrong block size")
		return nil, err
	}

	block, err := aes.NewCipher(bKey)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)

	return cipherText, nil
}

// Decrypt cipher text string into plain text string
func AES256CBCDecode(cipherText []byte, key string) ([]byte, error) {
	bKey := getMD5Hash(key)

	block, err := aes.NewCipher(bKey)
	if err != nil {
		panic(err)
	}

	if len(cipherText) < aes.BlockSize {
		panic("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	if len(cipherText)%aes.BlockSize != 0 {
		panic("cipherText is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	cipherText, err = pkcs7.Unpad(cipherText, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("error in Unpad function: %w", err)
	}
	return cipherText, nil
}

func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAES(t *testing.T) {

	t.Run("Encrypts and decrypts", func(t *testing.T) {

		plainTexts := []string{"1234567890", "123456789012345678901234567890123456789012345678901234567890", "1", "MySecretText", ""}
		key := "mytestkey"
		for _, plainText := range plainTexts {

			encrypted, err := AES256CBCEncode([]byte(plainText), key)
			if err != nil {
				t.Fatalf("Failed to encrypt: %s - %s", []byte(plainText), err.Error())
			}

			decrypted, err := AES256CBCDecode(encrypted, key)
			if err != nil {
				t.Fatalf("Failed to decrypt: %s - %s", []byte(plainText), err.Error())
			}

			assert.Equal(t, []byte(plainText), decrypted)
			assert.Equal(t, len([]byte(plainText)), len(decrypted))
		}
	})
}

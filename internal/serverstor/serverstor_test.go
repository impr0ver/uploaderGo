package serverstor_test

import (
	"crypto/rand"
	"log"
	"os"
	"testing"

	"github.com/impr0ver/uploaderGo/internal/serverstor"
	"github.com/stretchr/testify/assert"
)

const (
	_  = iota //ignore first value by assigning to blank identifier
	KB = 1 << (10 * iota)
	MB
)

func TestFilePathWalkDir(t *testing.T) {
	t.Run("===FilePathWalkDir===", func(t *testing.T) {
		tempDir, _ := os.MkdirTemp("../../public/uploads/", "TMP_")
		defer os.RemoveAll(tempDir)

		for i := 0; i < 9; i++ {
			file, err := os.CreateTemp(tempDir, "TMP_")
			if err != nil {
				log.Fatal(err)
			}

			token := make([]byte, i*MB)
			rand.Read(token)
			if _, err := file.Write(token); err != nil {
				log.Fatal(err)
			}
			file.Close()
		}

		res, err := serverstor.FilePathWalkDir(tempDir)
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(t, len(res), 10)

		for _, item := range res{
			assert.Equal(t, item.FileName[:3], "TMP", item.FileName)
		}
		
	})
}

package clientwork_test

import (
	"log"
	"os"
	"testing"

	"github.com/impr0ver/uploaderGo/internal/clientwork"
	"github.com/stretchr/testify/assert"
)

const (
	_  = iota //ignore first value by assigning to blank identifier
	KB = 1 << (10 * iota)
	MB
)

//func FilePathWalkDir(root string, fileMaxSize int64, key string) ([]string, string, error) {

func TestFilePathWalkDir(t *testing.T) {
	t.Run("===FilePathWalkDir===", func(t *testing.T) {
		res, tmpPath, err := clientwork.FilePathWalkDir("../../cmd/client/defaultfolder/", 1, "")
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(t, len(res), 11, "count files test")

		assert.Equal(t, tmpPath[:31], "../../cmd/client/defaultfolder/", tmpPath)
		os.RemoveAll(tmpPath)
	})
}

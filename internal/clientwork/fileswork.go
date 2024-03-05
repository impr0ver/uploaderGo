package clientwork

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/impr0ver/uploaderGo/internal/crypt"
	"go.uber.org/zap"
)

const (
	_  = iota //ignore first value by assigning to blank identifier
	KB = 1 << (10 * iota)
	MB
)

func FilePathWalkDir(root string, fileMaxSize int64, key string) ([]string, string, error) {
	var files []string

	tempDir, err := os.MkdirTemp(root, "TMP_")
	if err != nil {
		return nil, "", fmt.Errorf("error in MkdirTemp: %w", err)
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Mode()&os.ModeSymlink != os.ModeSymlink && info.Size() <= fileMaxSize*MB {

			newFilePath := filepath.Join(tempDir, fmt.Sprintf("%d", time.Now().UnixNano())+"_"+info.Name())

			if key != "" {
				_, err := fileCopyWithCrypt(path, newFilePath, key)
				if err != nil {
					return fmt.Errorf("error in fileCopyWithCrypt: %w", err)
				}
			} else {
				_, err := fileCopy(path, newFilePath)
				if err != nil {
					return fmt.Errorf("error in fileCopy: %w", err)
				}
			}
			files = append(files, newFilePath)
		}
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("error in filepath.Walk: %w", err)
	}

	return files, tempDir, nil
}

func fileCopy(src string, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)

	return nBytes, err
}

func fileCopyWithCrypt(src string, dst string, key string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	bData := crypt.StreamToByte(source)
	cipherData, err := crypt.AES256CBCEncode(bData, key)
	if err != nil {
		return 0, err
	}

	r := bytes.NewReader(cipherData)

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, r)

	return nBytes, err
}

func deleteTMPFolder(tmpDirPath string, sLogger *zap.SugaredLogger) {
	err := os.RemoveAll(tmpDirPath)
	if err != nil {
		sLogger.Error(err)
		return
	}
}

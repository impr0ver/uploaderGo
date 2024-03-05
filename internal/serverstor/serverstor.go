package serverstor

import (
	"fmt"
	"os"
	"path/filepath"
)

// struct fot marshal JSON-data list
type FileInfo struct {
	FileName string `json:"name"`
	FilePath string `json:"path"`
	FileSize int64  `json:"size"`
}

func FilePathWalkDir(root string) ([]FileInfo, error) {
	var fileInfos []FileInfo
	var fileInfo FileInfo
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		fileInfo.FileName = info.Name()
		fileInfo.FilePath = path
		fileInfo.FileSize = info.Size()

		fileInfos = append(fileInfos, fileInfo)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error in filepath.Walk: %w", err)
	}
	return fileInfos, nil
}

func DeleteFile(workDir string, filePath string) error {
	err := os.Remove(filepath.Join(workDir, filePath))
	if err != nil {
		return fmt.Errorf("error in os.Remove to delete file: %w", err)
	}
	return nil
}

package serverstor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/impr0ver/uploaderGo/internal/logger"
	"github.com/impr0ver/uploaderGo/internal/servconfig"
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
		if !info.IsDir() {
			fileInfo.FileName = info.Name()
			fileInfo.FilePath = path
			fileInfo.FileSize = info.Size()

			fileInfos = append(fileInfos, fileInfo)
		}

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

type MemoryStoragerInterface interface {
	AddNewFileInfo(ctx context.Context, fileName string, filePath string, fileSize int64) error
	GetFilePathByName(ctx context.Context, fileName string) (string, error)
	DeleteFileInfoByFilePath(ctx context.Context, filePath string) error
	RegisterNewUser(ctx context.Context, userName string, hash string) error
	GetUserByName(ctx context.Context, userName string) (DBUser, error)
	GetAllFileInfo(ctx context.Context) ([]DBData, error)
}

func NewStorage(cfg *servconfig.ServerConfig) MemoryStoragerInterface {
	var sLogger = logger.NewLogger()
	var memStor MemoryStoragerInterface

	ctxTimeOut, cancel := context.WithTimeout(context.Background(), servconfig.DefaultCtxTimeout)
	defer cancel()

	db, err := ConnectDB(ctxTimeOut, cfg.DatabaseDSN)
	if err != nil {
		sLogger.Fatalf("error DB: %v", err)
	}
	memStor = &DBStorage{DB: db.DB}

	return memStor
}

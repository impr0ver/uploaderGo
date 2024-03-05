package clientwork

import (
	"fmt"
	"time"

	"github.com/impr0ver/uploaderGo/internal/clientconfig"
	"github.com/impr0ver/uploaderGo/internal/clientreq"
	"go.uber.org/zap"
)

// work with cloud storage - DropBox
func CloudWork(cfg *clientconfig.Config, sLogger *zap.SugaredLogger) error {

	if cfg.RefreshToken != "" {
		sLogger.Info("Work with cloud storage DropBox...")

		//get new access token
		accessToken, err := clientreq.RefreshAccessToken(cfg.RefreshToken)
		if err != nil {
			return fmt.Errorf("error in RefreshAccessToken: %w", err)
		}
		sLogger.Info("New accessToken: ", accessToken)

		//delete file in cloud - DropBox
		//200 - Delete, 409 - Not found
		if cfg.DeletePath != "" {
			status, err := clientreq.DeleteDataInCloud(accessToken, cfg.DeletePath)
			if err != nil {
				return fmt.Errorf("error in DeleteDataInCloud: %w", err)
			}
			sLogger.Infof("Delete file \"%s\" status code: %s", cfg.DeletePath, status)
			return nil
		}

		//list data in cloud - DropBox
		if cfg.ListCloudData /*!= ""*/ {
			status, dataList, err := clientreq.ListDataInCloud(accessToken /*, cfg.ListCloudData*/)
			if err != nil {
				return fmt.Errorf("error in ListDataInCloud: %w", err)
			}

			sLogger.Infof("List folder status code: %s", status)

			for _, item := range dataList.Entries {
				sLogger.Infof("File name: %s, Cloud path: %s, size: %d", item.Name, item.PathDisplay, item.FileSize)
			}
			return nil
		}

		//work with target directory
		files, tmpDirPath, err := FilePathWalkDir(cfg.Root, cfg.FileMaxSize, cfg.Key)
		if err != nil {
			sLogger.Error(err)
			deleteTMPFolder(tmpDirPath, sLogger)
			return fmt.Errorf("error in FilePathWalkDir: %w", err)
		}
		defer deleteTMPFolder(tmpDirPath, sLogger)

		//upload files in cloud - DropBox
		//for cloud work do not use routines, because we get status code from cloud: 429 Too Many Requests
		for _, file := range files {
			status, err := clientreq.UploadDataInCloud(accessToken, file)
			if err != nil {
				sLogger.Error(err)
				continue
			}
			sLogger.Infof("Upload file \"%s\" status code: %s", file, status)
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

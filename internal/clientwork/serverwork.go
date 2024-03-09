package clientwork

import (
	"fmt"
	"sync"

	"github.com/impr0ver/uploaderGo/internal/clientconfig"
	"github.com/impr0ver/uploaderGo/internal/clientreq"
	"go.uber.org/zap"
)

// work with file server
func ServerWork(wg *sync.WaitGroup, cfg *clientconfig.Config, sLogger *zap.SugaredLogger) error {
	if cfg.RefreshToken == "" {
		//delete file from file server
		if cfg.DeletePath != "" {
			status, err := clientreq.DeleteDataFromServer(cfg.Address, cfg.DeletePath)
			if err != nil {
				return fmt.Errorf("error in DeleteDataFromServer: %w", err)
			}
			sLogger.Infof("Delete file \"%s\" status code: %s", cfg.DeletePath, status)
			return nil
		}

		//list data on file server
		if cfg.ListServerData {
			status, dataList, err := clientreq.ListDataFromServer(cfg.Address, cfg.Key)
			if err != nil {
				return fmt.Errorf("error in ListDataFromServer: %w", err)
			}
			sLogger.Infof("List folder status code: %s", status)

			for _, item := range *dataList {
				sLogger.Infof("File name: %s, Server path: %s, size: %d", item.FileName, item.FilePath, item.FileSize)
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

		if cfg.Mode { //upload files with parts
			chunkSlices := chunkSlice(files, cfg.FilesCount)

			sLogger.Info("Sending files parts: ", chunkSlices)
			err := clientreq.UploadDataMulti(cfg.Address, cfg.Key, chunkSlices)
			if err != nil {
				return fmt.Errorf("error in UploadDataMulti: %w", err)
			}

		} else { //upload files single
			cStatus := make(chan string, 1)
			cError := make(chan error, 1)

			for _, file := range files {
				file := file //ツツツ we know this feature!
				wg.Add(1)
				go func() {
					defer wg.Done()
					status, err := clientreq.UploadDataSingle(cfg.Address, cfg.Key, file)
					if err != nil {
						cError <- err
					}
					cStatus <- "Upload file \"" + file + "\" status code: " + status
				}()
			}

			for range files {
				select {
				case msgStatus := <-cStatus:
					sLogger.Info("received: ", msgStatus)
				case msgError := <-cError:
					sLogger.Info("received: ", msgError)
				}
			}

			sLogger.Info("Wait for complete all routines...")
			wg.Wait()
		}
	}
	return nil
}

func chunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string

	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}
	return chunks
}

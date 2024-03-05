package main

import (
	"sync"

	"github.com/impr0ver/uploaderGo/internal/clientconfig"
	"github.com/impr0ver/uploaderGo/internal/clientwork"
	"github.com/impr0ver/uploaderGo/internal/logger"
)

func main() {
	var sLogger = logger.NewLogger()
	cfg := clientconfig.InitConfig(sLogger)
	var wg sync.WaitGroup

	//work with cloud storage - DropBox
	err := clientwork.CloudWork(cfg, sLogger)
	if err != nil {
		sLogger.Error(err)
		return
	}

	//work with file server
	err = clientwork.ServerWork(&wg, cfg, sLogger)
	if err != nil {
		sLogger.Error(err)
		return
	}
}
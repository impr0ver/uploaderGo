package main

import (
	"github.com/impr0ver/uploaderGo/internal/handlers"
	"github.com/impr0ver/uploaderGo/internal/logger"
	"github.com/impr0ver/uploaderGo/internal/servconfig"
	"github.com/impr0ver/uploaderGo/internal/serverstor"
)

func main() {
	var sLogger = logger.NewLogger()
	cfg := servconfig.InitConfig(sLogger)

	memStor := serverstor.NewStorage(cfg)

	e := handlers.EchoRouter(memStor, cfg)

	//e.Logger.Fatal(e.Start("localhost:8080")) //HTTP
	e.Logger.Fatal(e.StartTLS(cfg.Address, "./.cert/cert.pem", "./.cert/key.pem")) //HTTPS
}

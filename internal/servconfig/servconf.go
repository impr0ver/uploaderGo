package servconfig

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

const (
	ROOTDIR = "../../public/uploads"
)

type ServerConfig struct {
	Address  string `env:"ADDRESS"`
	WorkDir  string `env:"WORKDIR"`
	Key      string `env:"KEY"`
	FullPath string
}

func InitConfig(sLogger *zap.SugaredLogger) *ServerConfig {
	cfg := &ServerConfig{}

	err := env.Parse(cfg)
	if err != nil {
		sLogger.Fatal(err)
	}

	flag.StringVar(&cfg.Address, "a", "localhost:8443", "Server address and port.")
	flag.StringVar(&cfg.WorkDir, "workdir", "/", "Path to store files.")
	flag.StringVar(&cfg.Key, "key", "", "Secret key for crypt/decrypt data with AES-256-CBC cipher algoritm.")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Address = envAddr
	}

	if envWorkDir := os.Getenv("WORKDIR"); envWorkDir != "" {
		cfg.WorkDir = envWorkDir
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	resPath := filepath.Join(ROOTDIR, cfg.WorkDir)
	checkDir(sLogger, resPath)
	cfg.FullPath = resPath

	return cfg
}

func checkDir(sLogger *zap.SugaredLogger, path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			sLogger.Fatal(err)
		}
	}
}

package clientconfig

import (
	"flag"
	"os"
	"strconv"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

const (
	_  = iota //ignore first value by assigning to blank identifier
	KB = 1 << (10 * iota)
	MB
)

// struct for get new access token
type Refresh struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// struct for unmarshal JSON-data list from cloud
type Entry struct {
	Tag            string `json:".tag"`
	Name           string `json:"name"`
	PathLower      string `json:"path_lower"`
	PathDisplay    string `json:"path_display"`
	ID             string `json:"id"`
	ClientModified string `json:"client_modified"`
	ServerModified string `json:"server_modified"`
	Rev            string `json:"rev"`
	FileSize       int64  `json:"size"`
	IsDownloadable bool   `json:"is_downloadable"`
	ContentHash    string `json:"content_hash"`
}

// struct for unmarshal JSON-data list from cloud
type DboxFolder struct {
	Entries Entries `json:"entries"`
	Cursor  string  `json:"cursor"`
	HasMore bool    `json:"has_more"`
}
type Entries []Entry

// struct for upload file in cloud
type DropboxUploadAPIArg struct {
	Autorename bool   `json:"autorename"`
	Mode       string `json:"mode"`
	Mute       bool   `json:"mute"`
	Path       string `json:"path"`
	Strict     bool   `json:"strict_conflict"`
}

// struct for request list data to cloud
type DropboxListAPIArg struct {
	IncDeleted        bool   `json:"include_deleted"`
	IncExpSharMembers bool   `json:"include_has_explicit_shared_members"`
	IncMediaInfo      bool   `json:"include_media_info"`
	IncMountedFolders bool   `json:"include_mounted_folders"`
	IncNonDwnFls      bool   `json:"include_non_downloadable_files"`
	Path              string `json:"path"`
	Recursive         bool   `json:"recursive"`
}

// struct for unmarshal JSON-data from server
type FServerFolder struct {
	FileName string `json:"name"`
	FilePath string `json:"path"`
	FileSize int64  `json:"size"`
}
type FServerFolders []FServerFolder

// main client struct
type Config struct {
	Address        string `env:"ADDRESS"`
	Root           string `env:"TARGETPATH"`
	Mode           bool   `env:"SENDMODE"`
	Key            string `env:"KEY"`
	FilesCount     int    `env:"FILESCOUNT"`
	RefreshToken   string `env:"REFRESHTOKEN"`
	DeletePath     string `env:"DELETEPATH"`
	ListCloudData  bool   `env:"LISTCLOUDDATA"`
	ListServerData bool   `env:"LISTSERVERDATA"`
	FileMaxSize    int64  `env:"FILEMAXSIZE"`
}

func InitConfig(sLogger *zap.SugaredLogger) *Config {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		sLogger.Fatal(err)
	}

	flag.StringVar(&cfg.Address, "a", "localhost:8443", "Server address and port.")
	flag.StringVar(&cfg.Root, "path", "defaultfolder", "Path to find files for sending.")
	flag.BoolVar(&cfg.Mode, "mode", true, "Upload files mode: multipart (true) or single (false).")
	flag.StringVar(&cfg.Key, "key", "", "Secret key for crypt data with AES-256-CBC cipher algoritm.")
	flag.IntVar(&cfg.FilesCount, "fcount", 3, "Files count in one multipart/form-data body in POST request.")
	flag.StringVar(&cfg.RefreshToken, "token", "", "Refresh token for get access token and work with cloud storage.")
	flag.StringVar(&cfg.DeletePath, "delete", "", "Delete file path.")
	flag.BoolVar(&cfg.ListCloudData, "listcloud", false, "List data in cloud DropBox.") 
	flag.BoolVar(&cfg.ListServerData, "listserver", false, "List data in file server.")
	flag.Int64Var(&cfg.FileMaxSize, "maxsize", 16, "Max size of send file in MB (<=16MB).")
	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.Address = envAddr
	}

	if envTargetPath := os.Getenv("TARGETPATH"); envTargetPath != "" {
		cfg.Root = envTargetPath
	}

	if envSendMode := os.Getenv("SENDMODE"); envSendMode != "" {
		boolValue, err := strconv.ParseBool(envSendMode)
		if err != nil {
			sLogger.Fatal(err)
		}
		cfg.Mode = boolValue
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}

	if envFileCount := os.Getenv("FILESCOUNT"); envFileCount != "" {
		intVar, err := strconv.Atoi(envFileCount)
		if err != nil {
			sLogger.Fatal(err)
		}
		cfg.FilesCount = intVar
	}

	if envRefreshToken := os.Getenv("REFRESHTOKEN"); envRefreshToken != "" {
		cfg.RefreshToken = envRefreshToken
	}

	if envDeletePath := os.Getenv("DELETEPATH"); envDeletePath != "" {
		cfg.DeletePath = envDeletePath
	}

	if envListCloudData := os.Getenv("LISTCLOUDDATA"); envListCloudData != "" {
		boolValue, err := strconv.ParseBool(envListCloudData)
		if err != nil {
			sLogger.Fatal(err)
		}
		cfg.ListCloudData = boolValue
	}

	if envListServerData := os.Getenv("LISTSERVERDATA"); envListServerData != "" {
		boolValue, err := strconv.ParseBool(envListServerData)
		if err != nil {
			sLogger.Fatal(err)
		}
		cfg.ListServerData = boolValue
	}

	if envFileMaxSize := os.Getenv("FILEMAXSIZE"); envFileMaxSize != "" {
		int64Var, err := strconv.ParseInt(envFileMaxSize, 10, 64)
		if err != nil {
			sLogger.Fatal(err)
		}
		cfg.FileMaxSize = int64Var
	}

	//check some program logic
	if cfg.DeletePath != "" && (cfg.ListCloudData || cfg.ListServerData) {
		sLogger.Fatal("error set flags arguments: flags delete and list can not be used together!")
	}

	//check if max filesize <= 16MB
	if cfg.FileMaxSize*MB > 16*MB {
		sLogger.Fatal("error set flag maxsize: must be <= 16MB!")
	}

	return cfg
}

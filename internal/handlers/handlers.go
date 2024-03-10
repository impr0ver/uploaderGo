package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/impr0ver/uploaderGo/internal/crypt"
	"github.com/impr0ver/uploaderGo/internal/logger"
	"github.com/impr0ver/uploaderGo/internal/servconfig"
	"github.com/impr0ver/uploaderGo/internal/serverstor"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/impr0ver/uploaderGo/internal/auth"
)

const (
	defaultCtxTimeout = servconfig.DefaultCtxTimeout
)

// TemplateRenderer is a custom html/template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func displayFormMain(c echo.Context, memStor serverstor.MemoryStoragerInterface) error {
	loginUser := c.Get("user")

	ctx, cancel := context.WithTimeout(c.Request().Context(), defaultCtxTimeout)
	defer cancel()

	dbData, _ := memStor.GetAllFileInfo(ctx)

	return c.Render(http.StatusOK, "upload.html", map[string]interface{}{
		"title":     "Загрузка файлов",
		"user":      loginUser,
		"message":   "",
		"filesdata": dbData,
	})
}

func displayFormRegister(c echo.Context) error {
	return c.Render(http.StatusOK, "register.html", map[string]interface{}{
		"title":     "Загрузка файлов",
	})
}

func displayFormLogin(c echo.Context) error {
	//kill cookie
	auth.WriteCookie(c, "Authorization", "", time.Now().Add(-1*time.Hour), "/", c.Request().URL.Hostname(), false, false) //for logout
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"title":     "Загрузка файлов",
	})
}

func uploadFilesMultiple(c echo.Context, cfg *servconfig.ServerConfig, memStor serverstor.MemoryStoragerInterface) error {
	var sLogger = logger.NewLogger()

	ctx, cancel := context.WithTimeout(c.Request().Context(), defaultCtxTimeout)
	defer cancel()

	err := c.Request().ParseMultipartForm(32 << 20)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	for _, files := range c.Request().MultipartForm.File {
		// Get the file from the request
		for _, file := range files {
			dst, err := os.Create(filepath.Join(cfg.FullPath, file.Filename))
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			defer dst.Close()

			multiPartFile, err := file.Open()
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			defer multiPartFile.Close()

			//check for decrypt data if need
			if cfg.Key != "" {
				bCryptData := crypt.StreamToByte(multiPartFile)
				decryptData, err := crypt.AES256CBCDecode(bCryptData, cfg.Key)
				if err != nil {
					return c.String(http.StatusInternalServerError, err.Error())
				}
				r := bytes.NewReader(decryptData)
				// Copy the contents of the file to the new file
				_, err = io.Copy(dst, r)
				if err != nil {
					return c.String(http.StatusInternalServerError, err.Error())
				}
			} else { //work with plaintext
				// Copy the contents of the file to the new file
				_, err = io.Copy(dst, multiPartFile)
				if err != nil {
					return c.String(http.StatusInternalServerError, err.Error())
				}
			}

			//work with DB: add new file data
			memStor.AddNewFileInfo(ctx, file.Filename, filepath.Join(cfg.FullPath, file.Filename), file.Size)

			sLogger.Info("File: ", file.Filename, "size: ", file.Size, " was uploaded successfully!")
		}
		c.String(http.StatusOK, "Files uploaded successfully\n")
	}
	return nil
}

func uploadFilesSingle(c echo.Context, cfg *servconfig.ServerConfig, memStor serverstor.MemoryStoragerInterface) error {
	var sLogger = logger.NewLogger()

	ctx, cancel := context.WithTimeout(c.Request().Context(), defaultCtxTimeout)
	defer cancel()

	err := c.Request().ParseMultipartForm(32 << 20) // 32 MB is the maximum file size
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Get the file from the request
	multiPartFile, handler, err := c.Request().FormFile("file")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer multiPartFile.Close()

	//check if send via web-from then add file timestamp prefix
	if c.Request().FormValue("send-from-web") != ""{
		handler.Filename = fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
	}

	dst, err := os.OpenFile(filepath.Join(cfg.FullPath, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	//check for decrypt data if need
	if cfg.Key != "" &&  c.Request().FormValue("send-from-web") == "" {
		bCryptData := crypt.StreamToByte(multiPartFile)
		decryptData, err := crypt.AES256CBCDecode(bCryptData, cfg.Key)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		r := bytes.NewReader(decryptData)
		// Copy the contents of the file to the new file
		_, err = io.Copy(dst, r)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	} else { //work with plaintext
		// Copy the contents of the file to the new file
		_, err = io.Copy(dst, multiPartFile)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	//work with DB: add new file data
	memStor.AddNewFileInfo(ctx, handler.Filename, filepath.Join(cfg.FullPath, handler.Filename), handler.Size)

	sLogger.Info("File: ", handler.Filename, " size: ", handler.Size, " was uploaded successfully!")

	//check if send via web-from 
	if c.Request().FormValue("send-from-web") != ""{
		return c.Redirect(http.StatusFound, "/index")
	} 
		
	return c.String(http.StatusOK, "File uploaded successfully\n")
}

func getFilesList(c echo.Context, cfg *servconfig.ServerConfig) error {
	filesInfo, err := serverstor.FilePathWalkDir(cfg.FullPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if cfg.Key != "" {
		jData, err := json.Marshal(filesInfo)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		cipherData, err := crypt.AES256CBCEncode(jData, cfg.Key)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		c.Blob(http.StatusOK, "text/html; charset=utf-8", cipherData)
	} else {
		c.JSON(http.StatusOK, filesInfo)
	}

	return nil
}

func deleteFileSingle(c echo.Context, cfg *servconfig.ServerConfig, memStor serverstor.MemoryStoragerInterface) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), defaultCtxTimeout)
	defer cancel()

	queryStr := c.QueryParam("filename")

	err := serverstor.DeleteFile(cfg.FullPath, queryStr)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	//work with DB: delete file data
	memStor.DeleteFileInfoByFilePath(ctx, filepath.Join(cfg.FullPath, queryStr))

	c.String(http.StatusOK, "File delete successfully\n")
	return nil
}

func EchoRouter(memStor serverstor.MemoryStoragerInterface, cfg *servconfig.ServerConfig) *echo.Echo {
	e := echo.New()

	e.Static("static", "../../internal/assets")
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("../../internal/templates/*.html")),
	}
	e.Renderer = renderer

	//middleware logger and decompress gzip
	e.Use(middleware.Logger())
	e.Use(middleware.Decompress())

	e.POST("/multiupload", func(c echo.Context) error {
		return uploadFilesMultiple(c, cfg, memStor)
	})
	e.POST("/upload", func(c echo.Context) error {
		return uploadFilesSingle(c, cfg, memStor)
	})

	e.GET("/list", func(c echo.Context) error {
		return getFilesList(c, cfg)
	})

	e.DELETE("/delete", func(c echo.Context) error {
		return deleteFileSingle(c, cfg, memStor)
	})

	//frontend: "register", "login"
	e.GET("/register", displayFormRegister)
	e.GET("/login", displayFormLogin)

	e.POST("/register", func(c echo.Context) error {
		return auth.RegisterUser(c, cfg, memStor)
	})
	e.POST("/login", func(c echo.Context) error {
		return auth.GenerateToken(c, cfg, memStor)
	})

	securedGroup := e.Group("")
	//middleware auth
	securedGroup.Use(auth.Auth)
	securedGroup.GET("/index", func(c echo.Context) error {
		return displayFormMain(c, memStor)
	})

	return e
}

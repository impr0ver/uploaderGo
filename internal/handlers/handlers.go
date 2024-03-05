package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/impr0ver/uploaderGo/internal/crypt"
	"github.com/impr0ver/uploaderGo/internal/logger"
	"github.com/impr0ver/uploaderGo/internal/servconfig"
	"github.com/impr0ver/uploaderGo/internal/serverstor"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer is a custom html/template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func displayForm(c echo.Context) error {
	return c.Render(http.StatusOK, "upload.html", nil)
}

func uploadFilesMultiple(c echo.Context, cfg *servconfig.ServerConfig) error {
	var sLogger = logger.NewLogger()
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

			sLogger.Info("File: ", file.Filename, "size: ", file.Size, " was uploaded successfully!")
			c.String(http.StatusOK, "File uploaded successfully\n")
		}
	}
	return nil
}

func uploadFilesSingle(c echo.Context, cfg *servconfig.ServerConfig) error {
	var sLogger = logger.NewLogger()
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

	dst, err := os.OpenFile(filepath.Join(cfg.FullPath, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

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

	sLogger.Info("File: ", handler.Filename, "size: ", handler.Size, " was uploaded successfully!")
	c.String(http.StatusOK, "File uploaded successfully\n")
	return nil
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

func deleteFileSingle(c echo.Context, cfg *servconfig.ServerConfig) error {
	queryStr := c.QueryParam("filename")

	err := serverstor.DeleteFile(cfg.FullPath, queryStr)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	c.String(http.StatusOK, "File delete successfully\n")
	return nil
}

func EchoRouter(cfg *servconfig.ServerConfig) *echo.Echo {
	e := echo.New()

	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("../../internal/templates/*.html")),
	}

	e.Use(middleware.Logger())

	e.POST("/multiupload", func(c echo.Context) error {
		return uploadFilesMultiple(c, cfg)
	})
	e.POST("/upload", func(c echo.Context) error {
		return uploadFilesSingle(c, cfg)
	})

	e.GET("/index", displayForm)

	e.GET("/list", func(c echo.Context) error {
		return getFilesList(c, cfg)
	})
	e.DELETE("/delete", func(c echo.Context) error {
		return deleteFileSingle(c, cfg)
	})

	return e
}

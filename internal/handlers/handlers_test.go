package handlers_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/impr0ver/uploaderGo/internal/handlers"
	"github.com/impr0ver/uploaderGo/internal/logger"
	"github.com/impr0ver/uploaderGo/internal/servconfig"
	"github.com/impr0ver/uploaderGo/internal/serverstor"
	"github.com/stretchr/testify/assert"
)


func Test_uploadFilesSingle(t *testing.T) {
	type want struct {
		httpStatus int
	}

	tests := []struct {
		name  string
		value string
		want  want
	}{
		{"test #1", "../../cmd/client/defaultfolder/myfile.txt", want{http.StatusOK}},
		{"test #2", "../../cmd/client/defaultfolder/myfile3.txt", want{http.StatusOK}},
		{"test #3", "../../cmd/client/defaultfolder/myfile9.txt", want{http.StatusOK}},
		{"test #4", "../../cmd/client/defaultfolder/ТЗ_выпускная_работа_1_v4.pdf", want{http.StatusOK}},	
	}

	var sLogger = logger.NewLogger()
	cfg := servconfig.InitConfig(sLogger)
	memStor := serverstor.NewStorage(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := handlers.EchoRouter(memStor, cfg)

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", tt.value)
			assert.NoError(t, err)
			sample, err := os.Open(tt.value)
			assert.NoError(t, err)
			defer sample.Close()

			_, err = io.Copy(part, sample)
			assert.NoError(t, err)
			assert.NoError(t, writer.Close())
			
			request := httptest.NewRequest(http.MethodPost, "/upload", body)
			request.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()

			//check status code
			if res.StatusCode != tt.want.httpStatus {
				t.Errorf("expected status code %d, got %d", tt.want.httpStatus, res.StatusCode)
			}

			bBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()

			assert.Equal(t, tt.want.httpStatus, res.StatusCode)
			assert.Equal(t, "File uploaded successfully\n", string(bBody))
		})
	}
}

func Test_uploadFilesMultiple(t *testing.T) {
	type want struct {
		httpStatus int
	}

	tests := []struct {
		name  string
		value []string
		want  want
	}{
		{"test #1", []string{"../../cmd/client/defaultfolder/myfile.txt", "../../cmd/client/defaultfolder/myfile10.txt", "../../cmd/client/defaultfolder/myfile2.txt"}, want{http.StatusOK}},
		{"test #2", []string{"../../cmd/client/defaultfolder/myfile3.txt", "../../cmd/client/defaultfolder/myfile4.txt", "../../cmd/client/defaultfolder/myfile5.txt"}, want{http.StatusOK}},
		{"test #3", []string{"../../cmd/client/defaultfolder/myfile6.txt", "../../cmd/client/defaultfolder/myfile7.txt", "../../cmd/client/defaultfolder/myfile8.txt"}, want{http.StatusOK}},
		{"test #4", []string{"../../cmd/client/defaultfolder/myfile9.txt", "../../cmd/client/defaultfolder/ТЗ_выпускная_работа_1_v4.pdf"}, want{http.StatusOK}},	
	}

	var sLogger = logger.NewLogger()
	cfg := servconfig.InitConfig(sLogger)
	memStor := serverstor.NewStorage(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := handlers.EchoRouter(memStor, cfg)

			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			///

			for _, file := range tt.value {
				fl, err := os.Open(file)
				assert.NoError(t, err)
				defer fl.Close()
	
				fw, err := writer.CreateFormFile("file", file)
				assert.NoError(t, err)
				
				_, err = io.Copy(fw, fl)
				assert.NoError(t, err)
			}
			assert.NoError(t, writer.Close())
			
			request := httptest.NewRequest(http.MethodPost, "/multiupload", body)
			request.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()

			//check status code
			if res.StatusCode != tt.want.httpStatus {
				t.Errorf("expected status code %d, got %d", tt.want.httpStatus, res.StatusCode)
			}

			bBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()

			assert.Equal(t, tt.want.httpStatus, res.StatusCode)
			assert.Equal(t, "Files uploaded successfully\n", string(bBody))
		})
	}
}
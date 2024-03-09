package clientreq

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/impr0ver/uploaderGo/internal/clientconfig"
	"github.com/impr0ver/uploaderGo/internal/crypt"
	"github.com/impr0ver/uploaderGo/internal/gzip"
	"github.com/impr0ver/uploaderGo/internal/logger"
)

func UploadDataSingle(address string, key string, filePath string) (string, error) {
	var sLogger = logger.NewLogger()
	var buf bytes.Buffer
	var gzipBuffer bytes.Buffer
	var req *http.Request

	w := multipart.NewWriter(&buf)
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error to open file: %w", err)
	}
	defer file.Close()

	fw, err := w.CreateFormFile("file", filePath)
	if err != nil {
		return "", fmt.Errorf("error create form file: %w", err)
	}
	if _, err := io.Copy(fw, file); err != nil {
		return "", fmt.Errorf("error io copy file: %w", err)
	}
	w.Close()

	//add Gzip compress if do not use binary crypto alghoritm
	if key == "" {
		gzip.CompressMultipart(&gzipBuffer, &buf)
		req, err = http.NewRequest("POST", fmt.Sprintf("https://%s/upload", address), &gzipBuffer) //also send on https://%s/multiupload works too
		if err != nil {
			return "", fmt.Errorf("error in NewRequest: %w", err)
		}
	} else {
		req, err = http.NewRequest("POST", fmt.Sprintf("https://%s/upload", address), &buf) //also send on https://%s/multiupload works too
		if err != nil {
			return "", fmt.Errorf("error in NewRequest: %w", err)
		}
	}

	//add gzip header if do not use binary crypto alghoritm
	if key == "" {
		req.Header.Set("Content-Encoding", "gzip")
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	sLogger.Info("Try to send file: ", filePath)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in send request to upload file in file server: %w", err)
	}
	defer resp.Body.Close()

	return resp.Status, nil
}

func UploadDataMulti(address string, key string, chunks [][]string) error {
	var sLogger = logger.NewLogger()
	var resp *http.Response

	for _, chunk := range chunks {
		var buf bytes.Buffer
		var gzipBuffer bytes.Buffer
		var req *http.Request
		var err error

		w := multipart.NewWriter(&buf)

		for _, file := range chunk {
			fl, err := os.Open(file)
			if err != nil {
				sLogger.Error(err)
				continue
			}
			defer fl.Close()

			fw, err := w.CreateFormFile("file", file)
			if err != nil {
				sLogger.Error(err)
				continue
			}
			if _, err := io.Copy(fw, fl); err != nil {
				sLogger.Error(err)
				continue
			}
		}
		w.Close()

		
		//add Gzip compress if do not use binary crypto alghoritm
		if key == "" {
			gzip.CompressMultipart(&gzipBuffer, &buf)
			req, err = http.NewRequest("POST", fmt.Sprintf("https://%s/multiupload", address), &gzipBuffer)
			if err != nil {
				return fmt.Errorf("error in NewRequest: %w", err)
			}
		} else {
			req, err = http.NewRequest("POST", fmt.Sprintf("https://%s/multiupload", address), &buf)
			if err != nil {
				return fmt.Errorf("error in NewRequest: %w", err)
			}
		}

		//add gzip header if do not use binary crypto alghoritm
		if key == "" {
			req.Header.Set("Content-Encoding", "gzip")
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Content-Type", w.FormDataContentType())

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Second * 5,
		}

		sLogger.Info("Try to send part files: ", chunk)

		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("error in send request to upload file: %w", err)
		}
		defer resp.Body.Close()

		sLogger.Info("Upload part files status code: ", resp.Status)
	}
	return nil
}

func DeleteDataFromServer(address string, deleteFile string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	baseURL := fmt.Sprintf("https://%s", address)
	resource := "delete"
	data := url.Values{}
	data.Add("filename", deleteFile)
	URI, _ := url.ParseRequestURI(baseURL)
	URI.Path = resource
	URI.RawQuery = data.Encode()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v", URI), nil)
	if err != nil {
		return "", fmt.Errorf("error in NewRequest to delete file from server: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in send request to delete file from server: %w", err)
	}
	defer resp.Body.Close()

	return resp.Status, nil
}

func ListDataFromServer(address string, key string) (string, *clientconfig.FServerFolders, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/list", address), nil)
	if err != nil {
		return "", nil, fmt.Errorf("error in NewRequest to list files from server: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("error in send request to list files from server: %w", err)
	}
	defer resp.Body.Close()

	fileServerListFolder := &clientconfig.FServerFolders{}

	//check for decrypt data if need
	if key != "" {
		bCryptData, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", nil, fmt.Errorf("error in ReadAll list files from server: %w", err)
		}

		decryptData, err := crypt.AES256CBCDecode(bCryptData, key)
		if err != nil {
			return "", nil, fmt.Errorf("error in AES256CBCDecode list file from server: %w", err)
		}

		err = json.Unmarshal(decryptData, &fileServerListFolder)
		if err != nil {
			return "", nil, fmt.Errorf("error in Unmarshal list file from server: %w", err)
		}

	} else { //work with plain text
		err = json.NewDecoder(resp.Body).Decode(&fileServerListFolder)
		if err != nil {
			return "", nil, fmt.Errorf("error in NewDecoder to list files from server: %w", err)
		}
	}

	return resp.Status, fileServerListFolder, nil
}

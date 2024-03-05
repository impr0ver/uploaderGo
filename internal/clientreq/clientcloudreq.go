package clientreq

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/impr0ver/uploaderGo/internal/clientconfig"
)

const (
	grantType    string = "refresh_token"
	clientID     string = "vxohuxe4tabtkza" //this id creates DropBox when you configure AppConsole in cloud storage
	clientSecret string = "2twf8tn33kck6vn" //this secret creates DropBox when you configure AppConsole in cloud storage
)

// cloud storage functions (DropBox): RefreshAccessToken
func RefreshAccessToken(refreshToken string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	data := url.Values{}
	data.Set("grant_type", grantType)
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	encodedData := data.Encode()

	req, err := http.NewRequest(http.MethodPost, "https://api.dropbox.com/oauth2/token", strings.NewReader(encodedData))
	if err != nil {
		return "", fmt.Errorf("error in NewRequest to get access token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in send request to get access token: %w", err)
	}
	defer resp.Body.Close()

	refresh := &clientconfig.Refresh{}
	err = json.NewDecoder(resp.Body).Decode(&refresh)
	if err != nil {
		return "", fmt.Errorf("error in NewDecoder to get access token: %w", err)
	}

	return refresh.AccessToken, nil
}

// cloud storage functions (DropBox): UploadDataInCloud
func UploadDataInCloud(accessToken string, filePath string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error to open file: %w", err)
	}
	fi, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("error to get file size: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest(http.MethodPost, "https://content.dropboxapi.com/2/files/upload", file)
	if err != nil {
		return "", fmt.Errorf("error in NewRequest: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken) //Set access token to DropBox

	dropBoxUploadAPI := &clientconfig.DropboxUploadAPIArg{
		Autorename: false,
		Mode:       "add",
		Mute:       false,
		Path:       "/data/" + fi.Name(),
		Strict:     false,
	}

	jData, err := json.Marshal(dropBoxUploadAPI)
	if err != nil {
		return "", fmt.Errorf("error in JSON-encoding data: %w", err)
	}

	req.Header.Add("Dropbox-API-Arg", string(jData))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Add("Content-Length", strconv.FormatInt(fi.Size(), 10))
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in send request to upload file in cloud: %w", err)
	}
	defer resp.Body.Close()

	return resp.Status, nil
}

// cloud storage functions (DropBox): DeleteDataInCloud
func DeleteDataInCloud(accessToken string, filePath string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	deletePath := map[string]string{"path": filePath}
	jData, err := json.Marshal(deletePath)
	if err != nil {
		return "", fmt.Errorf("error in JSON-encoding data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.dropboxapi.com/2/files/delete_v2", bytes.NewBuffer(jData))
	if err != nil {
		return "", fmt.Errorf("error in NewRequest: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken) //Set access token to DropBox
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(jData)))
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in send request to delete file from cloud: %w", err)
	}
	defer resp.Body.Close()

	return resp.Status, nil
}

// cloud storage functions (DropBox): ListDataInCloud
func ListDataInCloud(accessToken string/*, path string*/) (string, *clientconfig.DboxFolder, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	dropBoxListAPI := &clientconfig.DropboxListAPIArg{
		IncDeleted:        false,
		IncExpSharMembers: false,
		IncMediaInfo:      false,
		IncMountedFolders: true,
		IncNonDwnFls:      true,
		Path:              "/data/",/*path,*/
		Recursive:         false,
	}

	jData, err := json.Marshal(dropBoxListAPI)
	if err != nil {
		return "", nil, fmt.Errorf("error in JSON-encoding data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.dropboxapi.com/2/files/list_folder", bytes.NewBuffer(jData))
	if err != nil {
		return "", nil, fmt.Errorf("error in NewRequest: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken) //Set access token to DropBox
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(jData)))
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("error in send request to list files from cloud: %w", err)
	}
	defer resp.Body.Close()

	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	dropBoxListFolder := &clientconfig.DboxFolder{}
	err = json.NewDecoder(resp.Body).Decode(&dropBoxListFolder)
	if err != nil {
		return "", nil, fmt.Errorf("error in NewDecoder to list files from cloud: %w", err)
	}

	return resp.Status, dropBoxListFolder, nil
}


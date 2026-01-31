package filepathgetter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const getFileURL = "https://api.telegram.org/bot%s/getFile"

type FilepathGetter interface {
	GetFilepath(fileID string) (string, error)
}

type filepathGetter struct {
	client *http.Client
	token  string
}

func NewFilepathGetter(client *http.Client, token string) FilepathGetter {
	return &filepathGetter{
		client: client,
		token:  token,
	}
}

func (f filepathGetter) GetFilepath(fileID string) (string, error) {
	u := fmt.Sprintf(getFileURL, f.token)
	q := url.Values{}
	q.Set("file_id", fileID)

	req, _ := http.NewRequest(http.MethodGet, u+"?"+q.Encode(), nil)
	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var r getFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if !r.Ok || r.Result.FilePath == "" {
		return "", errors.New("getFile failed")
	}
	return r.Result.FilePath, nil
}

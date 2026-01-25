package updatesgetter

import (
	"encoding/json"
	"errors"
	"financer/pkg/telegram/entities"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const getUpdatesURL = "https://api.telegram.org/bot%s/getUpdates"

type UpdatesGetter interface {
	GetUpdates(offset int) ([]entities.Update, error)
}

type updatesGetter struct {
	client *http.Client

	token string
}

func NewUpdatesGetter(client *http.Client, token string) UpdatesGetter {
	return &updatesGetter{
		client: client,
		token:  token,
	}
}

func (ug *updatesGetter) GetUpdates(offset int) ([]entities.Update, error) {
	u := fmt.Sprintf(getUpdatesURL, ug.token)
	q := url.Values{}
	q.Set("timeout", strconv.Itoa(int(ug.client.Timeout)))
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	// Можно ограничить типы апдейтов:
	// q.Set("allowed_updates", `["message"]`)

	req, _ := http.NewRequest(http.MethodGet, u+"?"+q.Encode(), nil)
	resp, err := ug.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(b))
	}

	var r UpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if !r.Ok {
		return nil, errors.New("telegram ok=false")
	}
	return r.Result, nil
}

package menu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type SetMyCommandsRequest struct {
	Commands []BotCommand `json:"commands"`
}

func setMyCommands(client *http.Client, token string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", token)

	body := SetMyCommandsRequest{
		Commands: []BotCommand{
			{Command: "start", Description: "Запуск"},
			{Command: "ui", Description: "Меню"},
			{Command: "list", Description: "Последние загруженные файлы"},
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("setMyCommands status %d: %s", resp.StatusCode, string(rb))
	}
	return nil
}

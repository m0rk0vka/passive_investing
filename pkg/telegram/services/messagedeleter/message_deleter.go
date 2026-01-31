package messagedeleter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const deleteMessageURL = "https://api.telegram.org/bot%s/deleteMessage"

type MessageDeleter interface {
	DeleteMessage(chatID int64, messageID int) error
}

type messageDeleter struct {
	client *http.Client
	token  string
}

func NewMessageDeleter(client *http.Client, token string) MessageDeleter {
	return &messageDeleter{
		client: client,
		token:  token,
	}
}

func (md *messageDeleter) DeleteMessage(chatID int64, messageID int) error {
	endpoint := fmt.Sprintf(deleteMessageURL, md.token)

	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("message_id", strconv.Itoa(messageID))

	resp, err := md.client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// deleteMessage возвращает ok=true/false, но для MVP достаточно проверить HTTP status
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deleteMessage status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

package messagesender

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const sendMessageURL = "https://api.telegram.org/bot%s/sendMessage"

type MessageSender interface {
	SendMessage(message Message) error
}

type messageSender struct {
	client *http.Client
	token  string
}

func NewMessageSender(client *http.Client, token string) MessageSender {
	return &messageSender{
		client: client,
		token:  token,
	}
}

func (s messageSender) SendMessage(message Message) error {
	u := fmt.Sprintf(sendMessageURL, s.token)
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(message.chatID, 10))
	form.Set("text", message.text)
	if message.inlineKeyboard.InlineKeyboard != nil {
		kbJSON, err := json.Marshal(message.inlineKeyboard)
		if err != nil {
			return err
		}

		form.Set("reply_markup", string(kbJSON))
	}

	resp, err := s.client.PostForm(u, form)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

package messagesender

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const sendMessageURL = "https://api.telegram.org/bot%s/sendMessage"

type MessageSender interface {
	SendMessage(message Message) (massageID int, err error)
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

func (s messageSender) SendMessage(message Message) (int, error) {
	u := fmt.Sprintf(sendMessageURL, s.token)
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(message.chatID, 10))
	form.Set("text", message.text)
	if message.inlineKeyboard.InlineKeyboard != nil {
		kbJSON, err := json.Marshal(message.inlineKeyboard)
		if err != nil {
			return -1, fmt.Errorf("marshal inline keyboard: %w", err)
		}

		form.Set("reply_markup", string(kbJSON))
	}

	resp, err := s.client.PostForm(u, form)
	if err != nil {
		return -1, fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	var messageResponse sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&messageResponse); err != nil {
		return -1, fmt.Errorf("decode send message response: %w", err)
	}

	if !messageResponse.Ok {
		return -1, errors.New("send message ok = false")
	}

	return messageResponse.Result.MessageID, nil
}

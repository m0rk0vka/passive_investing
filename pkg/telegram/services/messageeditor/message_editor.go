package messageeditor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

const editMessageTextURL = "https://api.telegram.org/bot%s/editMessageText"

type MessageEditor interface {
	EditMessage(chatID int64, messageID int, text string, parseMode string, kb entities.InlineKeyboardMarkup) error
}

type messageEditor struct {
	client *http.Client
	token  string
}

func NewMessageEditor(client *http.Client, token string) MessageEditor {
	return &messageEditor{
		client: client,
		token:  token,
	}
}

func (m *messageEditor) EditMessage(
	chatID int64,
	messageID int,
	text string,
	parseMode string,
	kb entities.InlineKeyboardMarkup,
) error {
	endpoint := fmt.Sprintf(editMessageTextURL, m.token)

	kbJSON, err := json.Marshal(kb)
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("message_id", strconv.Itoa(messageID))
	form.Set("text", text)
	if parseMode != "" {
		form.Set("parse_mode", parseMode)
	}
	form.Set("reply_markup", string(kbJSON))

	resp, err := m.client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("editMessageText status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

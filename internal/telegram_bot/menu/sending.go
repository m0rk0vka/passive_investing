package menu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func sendMessageWithKeyboard(client *http.Client, token string, chatID int64, text string, kb InlineKeyboardMarkup) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	kbJSON, _ := json.Marshal(kb)
	form := url.Values{}
	form.Set("chat_id", strconv.FormatInt(chatID, 10))
	form.Set("text", text)
	// form.Set("parse_mode", "Markdown") // или убрать и слать plain text
	form.Set("reply_markup", string(kbJSON))

	resp, err := client.PostForm(u, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

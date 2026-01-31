package callbackqueryanswerer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const answerCallbackQueryURL = "https://api.telegram.org/bot%s/answerCallbackQuery"

type CallbackQueryAnswerer interface {
	AnswerCallbackQuery(callbackQueryID string, text string, showAlert bool) error
}

type callbackQueryAnswerer struct {
	client *http.Client
	token  string
}

func NewCallbackQueryAnswerer(client *http.Client, token string) CallbackQueryAnswerer {
	return &callbackQueryAnswerer{
		client: client,
		token:  token,
	}
}

func (c *callbackQueryAnswerer) AnswerCallbackQuery(callbackQueryID string, text string, showAlert bool) error {
	endpoint := fmt.Sprintf(answerCallbackQueryURL, c.token)

	form := url.Values{}
	form.Set("callback_query_id", callbackQueryID)
	if text != "" {
		form.Set("text", text)
	}
	if showAlert {
		form.Set("show_alert", "true")
	}

	resp, err := c.client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("answerCallbackQuery status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

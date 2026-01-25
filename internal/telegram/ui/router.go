package ui

import (
	"context"
	"net/http"

	messagesender "github.com/m0rk0vka/passive_investing/pkg/telegram/message_sender"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/visualizer/entities"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/visualizer/renderers"
)

type TelegramBotVisualizer interface {
	Visualize(chatID int64) error
}

type telegramBotVisualizer struct {
	client *http.Client
	token  string

	sessionStore SessionStore

	messageSender messagesender.MessageSender
}

func NewTelegramBotVisualizer(client *http.Client, token string) TelegramBotVisualizer {
	return &telegramBotVisualizer{
		client:        client,
		token:         token,
		sessionStore:  NewSessionStore(),
		messageSender: messagesender.NewMessageSender(client, token),
	}
}

func (t *telegramBotVisualizer) Visualize(chatID int64) error {
	return t.RenderHomeScreen(chatID)
}

func (t telegramBotVisualizer) RenderHomeScreen(chatID int64) error {
	hr := renderers.HomeRenderer{}
	data, _ := hr.Render(context.TODO(), 0, entities.UIState{})
	return t.messageSender.SendMessage(messagesender.NewInlineKeyboardMessage(chatID, data.Text, data.Kb))
}

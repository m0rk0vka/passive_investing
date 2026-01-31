package ui

import (
	"context"
	"fmt"
	"net/http"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/renderers"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagedeleter"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagesender"
)

type TelegramBotVisualizer interface {
	Visualize(chatID int64) error
}

type telegramBotVisualizer struct {
	client *http.Client
	token  string

	sessionStore SessionStore

	messageSender  messagesender.MessageSender
	messageDeleter messagedeleter.MessageDeleter
}

func NewTelegramBotVisualizer(client *http.Client, token string) TelegramBotVisualizer {
	return &telegramBotVisualizer{
		client:         client,
		token:          token,
		sessionStore:   NewSessionStore(),
		messageSender:  messagesender.NewMessageSender(client, token),
		messageDeleter: messagedeleter.NewMessageDeleter(client, token),
	}
}

func (t *telegramBotVisualizer) Visualize(chatID int64) error {
	session, ok := t.sessionStore.Get(chatID)
	if ok {
		if err := t.messageDeleter.DeleteMessage(chatID, session.MessageID()); err != nil {
			fmt.Println("WARNING: failed to delete old messsage %w", err)
		}
	}
	session = NewSession(chatID)
	t.sessionStore.Put(chatID, session)

	err := t.RenderHomeScreen(session)
	if err != nil {
		return fmt.Errorf("failed to render home screen: %w", err)
	}

	return nil
}

func (t *telegramBotVisualizer) RenderHomeScreen(session Session) error {
	hr := renderers.HomeRenderer{}
	data, _ := hr.Render(context.TODO(), 0, entities.UIState{})
	messageID, err := t.messageSender.SendMessage(
		messagesender.NewInlineKeyboardMessage(session.ChatID(), data.Text, data.Kb))
	if err != nil {
		return err
	}
	session.SetMessageID(messageID)
	t.sessionStore.Put(session.ChatID(), session)
	return nil
}

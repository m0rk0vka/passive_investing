package ui

import (
	"context"
	"fmt"
	"net/http"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/renderers"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagedeleter"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messageeditor"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagesender"
	"go.uber.org/zap"
)

type TelegramBotVisualizer interface {
	Visualize(chatID int64) error
	ProcessCallbackQuery(callbackQuery *domainEntities.CallbackQuery) error
}

type telegramBotVisualizer struct {
	ctx    context.Context
	logger *zap.Logger

	client *http.Client
	token  string

	sessionStore SessionStore

	visualizer *Visualizer

	messageSender  messagesender.MessageSender
	messageEditor  messageeditor.MessageEditor
	messageDeleter messagedeleter.MessageDeleter
}

func NewTelegramBotVisualizer(ctx context.Context, client *http.Client, token string, logger *zap.Logger) TelegramBotVisualizer {
	return &telegramBotVisualizer{
		ctx:    ctx,
		logger: logger,

		client: client,
		token:  token,

		sessionStore: NewSessionStore(),

		visualizer: NewVisualizer(renderers.Renderers),

		messageSender:  messagesender.NewMessageSender(client, token),
		messageDeleter: messagedeleter.NewMessageDeleter(client, token),
		messageEditor:  messageeditor.NewMessageEditor(client, token),
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

func (t *telegramBotVisualizer) ProcessCallbackQuery(callbackQuery *domainEntities.CallbackQuery) error {
	session, ok := t.sessionStore.Get(callbackQuery.Message.Chat.ID)
	if !ok {
		if err := t.processCallbackForOldSession(callbackQuery); err != nil {
			return fmt.Errorf("failed to process callbackQuery for old session %w", err)
		}
		return nil
	}

	if err := t.processCallbackQuery(session, callbackQuery); err != nil {
		return fmt.Errorf("failed to process callbackQuery for existed session %w", err)
	}
	return nil
}

func (t *telegramBotVisualizer) processCallbackForOldSession(callbackQuery *domainEntities.CallbackQuery) error {
	chatID := callbackQuery.Message.Chat.ID
	const sorryMsg = "please run new /ui command"
	_, err := t.messageSender.SendMessage(messagesender.NewSimpleMessage(chatID, sorryMsg))
	if err != nil {
		fmt.Println("ERROR: failed to send sorry message: %w", err)
	}

	err = t.messageDeleter.DeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	if err != nil {
		fmt.Println("ERROR: failed to delete message: %w", err)
	}

	return nil
}

func (t *telegramBotVisualizer) processCallbackQuery(session Session, callbackQuery *domainEntities.CallbackQuery) error {
	switch callbackQuery.Data {
	case entities.CBClose:
		err := t.messageDeleter.DeleteMessage(session.ChatID(), session.MessageID())
		if err != nil {
			return fmt.Errorf("failed to delete message: %w", err)
		}
		t.sessionStore.Delete(session.ChatID())
		return nil
	case entities.CBBack:
		session.PopOrHome()
	case entities.CBNavHome:
		session = NewSession(session.ChatID())
		session.SetMessageID(callbackQuery.Message.MessageID)
		session.SetState(entities.UIState{
			Screen: entities.ScreenHome,
		})
		t.sessionStore.Put(session.ChatID(), session)
	case entities.CBNavPortfolios:
		session.PushCurrentState()
		session.SetState(entities.UIState{
			Screen: entities.ScreenPortfolioList,
		})
	default:
		return fmt.Errorf("unknown callback query: %s", callbackQuery.Data)
	}

	rendered, err := t.visualizer.Render(t.ctx, session.ChatID(), session.State)
	if err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	t.logger.Info("visualiser result", zap.Object("rendered", rendered))

	const parseMode = ""

	return t.messageEditor.EditMessage(session.ChatID(), session.MessageID(), rendered.Text, parseMode, rendered.Kb)
}

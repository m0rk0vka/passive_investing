package ui

import (
	"context"
	"fmt"
	"net/http"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/mocks"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/renderers"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/repos"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagedeleter"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messageeditor"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagesender"
	"go.uber.org/zap"
)

const (
	msgSessionExpired = "please run new /ui command"
	parseModeDefault  = ""
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

	renderer *Renderer

	messageSender  messagesender.MessageSender
	messageEditor  messageeditor.MessageEditor
	messageDeleter messagedeleter.MessageDeleter

	repo repos.PortfolioRepo
}

func NewTelegramBotVisualizer(ctx context.Context, client *http.Client, token string, logger *zap.Logger) TelegramBotVisualizer {
	return &telegramBotVisualizer{
		ctx:    ctx,
		logger: logger,

		client: client,
		token:  token,

		sessionStore: NewSessionStore(),

		renderer: NewRenderer(renderers.Renderers),

		messageSender:  messagesender.NewMessageSender(client, token),
		messageDeleter: messagedeleter.NewMessageDeleter(client, token),
		messageEditor:  messageeditor.NewMessageEditor(client, token),

		repo: &mocks.MockPortfolioRepo{},
	}
}

func (t *telegramBotVisualizer) Visualize(chatID int64) error {
	session, ok := t.sessionStore.Get(chatID)
	if ok {
		if err := t.messageDeleter.DeleteMessage(chatID, session.MessageID()); err != nil {
			t.logger.Warn("failed to delete old message", zap.Error(err))
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
	_, err := t.messageSender.SendMessage(messagesender.NewSimpleMessage(chatID, msgSessionExpired))
	if err != nil {
		t.logger.Error("failed to send sorry message", zap.Error(err))
	}

	err = t.messageDeleter.DeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	if err != nil {
		t.logger.Error("failed to delete message", zap.Error(err))
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
	case entities.CBNavPortfolios:
		session.PushCurrentState()
		session.SetState(entities.UIState{
			Screen: entities.ScreenPortfolioList,
		})
	case entities.CBNavPositions:
		session.PushCurrentState()
		session.SetState(entities.UIState{
			Screen:      entities.ScreenPortfolioPositions,
			PortfolioID: session.State.PortfolioID,
			Period:      session.State.Period,
		})
	case entities.CBPeriodNext:
		nextPeriod, err := t.repo.GetNextPeriod(t.ctx, session.ChatID(), session.State.PortfolioID, session.State.Period)
		if err != nil {
			return fmt.Errorf("failed to get next period: %w", err)
		}
		session.SetState(entities.UIState{
			Screen:      entities.ScreenPortfolioPositions,
			PortfolioID: session.State.PortfolioID,
			Period:      nextPeriod,
		})
	case entities.CBPeriodPrev:
		prevPeriod, err := t.repo.GetPrevPeriod(t.ctx, session.ChatID(), session.State.PortfolioID, session.State.Period)
		if err != nil {
			return fmt.Errorf("failed to get prev period: %w", err)
		}
		session.SetState(entities.UIState{
			Screen:      entities.ScreenPortfolioPositions,
			PortfolioID: session.State.PortfolioID,
			Period:      prevPeriod,
		})
	default:
		portfolioId, ok := entities.IsOpenPortfolio(callbackQuery.Data)
		if !ok {
			return fmt.Errorf("unknown callback query: %s", callbackQuery.Data)
		}
		lastPeriod, err := t.repo.GetLastPeriod(t.ctx, session.ChatID(), session.State.PortfolioID)
		if err != nil {
			return fmt.Errorf("failed to get last period: %w", err)
		}
		session.PushCurrentState()
		session.SetState(entities.UIState{
			Screen:      entities.ScreenPortfolioSum,
			PortfolioID: portfolioId,
			Period:      lastPeriod,
		})
	}

	t.sessionStore.Put(session.ChatID(), session)

	rendered, err := t.renderer.Render(t.ctx, session.ChatID(), session.State)
	if err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	t.logger.Info("visualiser result", zap.Object("rendered", rendered))

	return t.messageEditor.EditMessage(session.ChatID(), session.MessageID(), rendered.Text, parseModeDefault, rendered.Kb)
}

package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/callbackqueryanswerer"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/filedownloader"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/messagesender"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/poller"
	"go.uber.org/zap"
)

const (
	msgWelcome          = "Пришли мне XLSX отчёт документом, я сохраню его."
	msgWrongFileFormat  = "Пришли как .xlsx, пожалуйста."
	msgFileDownloadFail = "Не смог скачать файл: "
	msgFileSaved        = "Файл сохранен.\nИмя: %s\nSHA256: %s"
)

var _ poller.UpdatesProcessor = (*updatesProcessor)(nil)

type updatesProcessor struct {
	ctx context.Context

	client  *http.Client
	token   string
	dataDir string
	logger  *zap.Logger

	messageSender         messagesender.MessageSender
	fileDownloader        filedownloader.FileDownloader
	callbackQueryAnswerer callbackqueryanswerer.CallbackQueryAnswerer

	visualizer ui.TelegramBotVisualizer
}

func NewUpdatesProcessor(ctx context.Context, client *http.Client, token string, dataDir string, logger *zap.Logger) (poller.UpdatesProcessor, error) {
	u := &updatesProcessor{
		ctx:     ctx,
		client:  client,
		token:   token,
		dataDir: dataDir,
		logger:  logger,
	}
	if err := u.init(); err != nil {
		return nil, err
	}
	return u, nil
}

func (u *updatesProcessor) init() error {
	fileDownloader, err := filedownloader.NewFileDownloader(u.client, u.token, u.dataDir)
	if err != nil {
		return err
	}
	u.fileDownloader = fileDownloader
	u.messageSender = messagesender.NewMessageSender(u.client, u.token)
	u.visualizer = ui.NewTelegramBotVisualizer(u.ctx, u.client, u.token, u.logger)
	u.callbackQueryAnswerer = callbackqueryanswerer.NewCallbackQueryAnswerer(u.client, u.token)
	return nil
}

func (u *updatesProcessor) ProcessUpdates(updates []entities.Update) int {
	offset := -1
	for _, update := range updates {
		offset = u.processUpdate(update)
	}

	return offset
}

func (u *updatesProcessor) processUpdate(update entities.Update) (offset int) {
	offset = update.UpdateID + 1

	if (update.Message == nil || update.Message.From == nil) && update.CallbackQuery == nil {
		return
	}

	// Логируем непустой апдейт
	u.logger.Info("processing update", zap.Object("update", update))

	if update.CallbackQuery != nil {
		err := u.processCallbackQuery(update.CallbackQuery)
		if err != nil {
			u.logger.Error("failed to process callback query", zap.Error(err))
			return
		}
		return
	}

	// Команды
	if strings.HasPrefix(strings.TrimSpace(update.Message.Text), "/start") {
		_, err := u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, msgWelcome))
		if err != nil {
			u.logger.Error("failed to send message", zap.Error(err))
			return
		}
		return
	}

	if strings.HasPrefix(strings.TrimSpace(update.Message.Text), "/ui") {
		err := u.visualizer.Visualize(update.Message.Chat.ID)
		if err != nil {
			u.logger.Error("failed to render home screen", zap.Error(err))
			return
		}
		return
	}

	// Документ
	if update.Message.Document != nil {
		doc := update.Message.Document
		// простая фильтрация по расширению
		if !strings.HasSuffix(strings.ToLower(doc.FileName), ".xlsx") {
			_, err := u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, msgWrongFileFormat))
			if err != nil {
				u.logger.Error("failed to send message", zap.Error(err))
				return
			}
			return
		}

		sha, err := u.fileDownloader.DownloadFile(doc.FileID)
		if err != nil {
			_, err := u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, msgFileDownloadFail+err.Error()))
			if err != nil {
				u.logger.Error("failed to send message", zap.Error(err))
				return
			}
			return
		}

		msg := fmt.Sprintf(msgFileSaved, doc.FileName, sha)
		_, err = u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, msg))
		if err != nil {
			u.logger.Error("failed to send message", zap.Error(err))
			return
		}
	}

	return
}

func (u *updatesProcessor) processCallbackQuery(callbackQuery *entities.CallbackQuery) error {
	err := u.callbackQueryAnswerer.AnswerCallbackQuery(callbackQuery.ID, "", false)
	if err != nil {
		u.logger.Error("failed to answer on callback query", zap.Error(err))
	}
	return u.visualizer.ProcessCallbackQuery(callbackQuery)
}

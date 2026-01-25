package core

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	filedownloader "github.com/m0rk0vka/passive_investing/pkg/telegram/services/file_downloader"
	messagesender "github.com/m0rk0vka/passive_investing/pkg/telegram/services/message_sender"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/poller"
)

var _ poller.UpdatesProcessor = (*updatesProcessor)(nil)

type updatesProcessor struct {
	client  *http.Client
	token   string
	dataDir string

	messageSender  messagesender.MessageSender
	fileDownloader filedownloader.FileDownloader
}

func NewUpdatesProcessor(client *http.Client, token string, dataDir string) (poller.UpdatesProcessor, error) {
	u := &updatesProcessor{
		client:  client,
		token:   token,
		dataDir: dataDir,
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

	if update.Message == nil || update.Message.From == nil {
		return
	}

	// Команды
	if strings.HasPrefix(strings.TrimSpace(update.Message.Text), "/start") {
		_ = u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, "Пришли мне XLSX отчёт документом, я сохраню его."))
		return
	}

	// if strings.HasPrefix(strings.TrimSpace(update.Message.Text), "/ui") {
	// 	_ = u.visualizer.Visualize(update.Message.Chat.ID)
	// 	continue
	// }

	// Документ
	if update.Message.Document != nil {
		doc := update.Message.Document
		// простая фильтрация по расширению
		if !strings.HasSuffix(strings.ToLower(doc.FileName), ".xlsx") {
			_ = u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, "Пришли как .xlsx, пожалуйста."))
			return
		}

		sha, err := u.fileDownloader.DownloadFile(doc.FileID)
		if err != nil {
			_ = u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, "Не смог скачать файл: "+err.Error()))
			return
		}

		msg := fmt.Sprintf("Файл сохранен.\nИмя: %s\nSHA256: %s", doc.FileName, sha)
		_ = u.messageSender.SendMessage(messagesender.NewSimpleMessage(update.Message.Chat.ID, msg))
	}

	return
}

package poller

import (
	"context"
	"financer/pkg/telegram/entities"
	updatesGetter "financer/pkg/telegram/services/updates_getter"
	"fmt"
	"net/http"
	"time"
)

type TelegramBotPoller interface {
	Polling() error
}

type UpdatesProcessor interface {
	ProcessUpdates(updates []entities.Update) (offset int)
}

type OffsetKeepper interface {
	GetOffset() (int, error)
	SetOffset(offset int) error
}

type telegramBotPoller struct {
	ctx context.Context

	updatesGetter    updatesGetter.UpdatesGetter
	updatesProcessor UpdatesProcessor

	offsetKeepper OffsetKeepper
}

func NewTelegramBotPoller(
	ctx context.Context,
	client *http.Client,
	token string,
	updatesProcessor UpdatesProcessor,
	offsetKeepper OffsetKeepper,
) TelegramBotPoller {
	return &telegramBotPoller{
		ctx:              ctx,
		updatesGetter:    updatesGetter.NewUpdatesGetter(client, token),
		updatesProcessor: updatesProcessor,
		offsetKeepper:    offsetKeepper,
	}
}

func (t *telegramBotPoller) Polling() error {
	offset, err := t.offsetKeepper.GetOffset()
	if err != nil {
		return fmt.Errorf("get offset error: %w", err)
	}
	fmt.Println("Bot started. offset=", offset)

	for {
		select {
		case <-t.ctx.Done():
			return nil
		default:
		}
		updates, err := t.updatesGetter.GetUpdates(offset)
		if err != nil {
			fmt.Println("getUpdates error:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Println("got updates:", updates)

		processedOffset := t.updatesProcessor.ProcessUpdates(updates)

		if processedOffset == -1 {
			continue
		}

		offset = processedOffset
		t.offsetKeepper.SetOffset(offset)
	}
}

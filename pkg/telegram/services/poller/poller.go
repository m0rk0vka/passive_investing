package poller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/updatesgetter"
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

	updatesGetter    updatesgetter.UpdatesGetter
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
		updatesGetter:    updatesgetter.NewUpdatesGetter(client, token),
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
		if len(updates) == 0 {
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

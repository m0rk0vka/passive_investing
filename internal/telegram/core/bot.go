package core

import (
	"context"
	"financer/pkg/environment"
	"financer/pkg/telegram/services/poller"
	"net/http"
	"time"
)

type TelegramBot struct {
	ctx context.Context

	poller poller.TelegramBotPoller
}

func NewTelegramBot(ctx context.Context) (*TelegramBot, error) {
	tb := &TelegramBot{ctx: ctx}
	if err := tb.init(); err != nil {
		return nil, err
	}
	return tb, nil
}

func (t *TelegramBot) init() error {
	client := &http.Client{Timeout: 90 * time.Second}
	token := environment.MustEnv("BOT_TOKEN")
	dataDir := environment.MustEnv("RAW_DATA_DIR")

	updatesProcessor, err := NewUpdatesProcessor(client, token, dataDir)
	if err != nil {
		return err
	}

	poller := poller.NewTelegramBotPoller(
		t.ctx, client, token, updatesProcessor, NewOffsetKeepper())
	t.poller = poller
	return nil
}

func (t *TelegramBot) Start() error {
	return t.poller.Polling()
}

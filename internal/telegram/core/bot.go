package core

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/m0rk0vka/passive_investing/pkg/environment"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/services/poller"
	"go.uber.org/zap"
)

type TelegramBot struct {
	ctx    context.Context
	logger *zap.Logger
	db     *sql.DB

	poller poller.TelegramBotPoller
}

func NewTelegramBot(ctx context.Context, db *sql.DB, logger *zap.Logger) (*TelegramBot, error) {
	tb := &TelegramBot{
		ctx:    ctx,
		db:     db,
		logger: logger,
	}
	if err := tb.init(); err != nil {
		return nil, err
	}
	return tb, nil
}

func (t *TelegramBot) init() error {
	client := &http.Client{Timeout: 90 * time.Second}
	token := environment.MustEnv("BOT_TOKEN")
	dataDir := environment.MustEnv("RAW_DATA_DIR")

	updatesProcessor, err := NewUpdatesProcessor(t.ctx, client, token, dataDir, t.db, t.logger)
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

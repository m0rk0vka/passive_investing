package main

import (
	"context"
	"fmt"
	"os"

	"github.com/m0rk0vka/passive_investing/internal/telegram/core"
	"go.uber.org/zap"
)

func main() {
	// Создаем production логгер
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("Error creating logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	telegramBot, err := core.NewTelegramBot(context.Background(), logger)
	if err != nil {
		logger.Fatal("Error creating telegram bot", zap.Error(err))
	}

	if err := telegramBot.Start(); err != nil {
		logger.Fatal("Error starting telegram bot", zap.Error(err))
	}
}

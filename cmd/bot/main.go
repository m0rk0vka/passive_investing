package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/m0rk0vka/passive_investing/internal/telegram/core"
	"github.com/m0rk0vka/passive_investing/internal/worker"
	"github.com/m0rk0vka/passive_investing/pkg/environment"
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

	// Подключаемся к БД
	dbURL := environment.MustEnv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("Error opening database", zap.Error(err))
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		logger.Fatal("Error pinging database", zap.Error(err))
	}
	logger.Info("Database connected successfully")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем воркер-парсер
	parserWorker := worker.NewParserWorker(db, logger)
	go parserWorker.Start(ctx)

	// Запускаем Telegram бота
	telegramBot, err := core.NewTelegramBot(ctx, db, logger)
	if err != nil {
		logger.Fatal("Error creating telegram bot", zap.Error(err))
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := telegramBot.Start(); err != nil {
			logger.Fatal("Error starting telegram bot", zap.Error(err))
		}
	}()

	<-sigChan
	logger.Info("Shutting down gracefully...")
	cancel()
}

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/m0rk0vka/passive_investing/internal/telegram/core"
)

func main() {
	telegramBot, err := core.NewTelegramBot(context.Background())
	if err != nil {
		fmt.Println("Error creating telegram bot:", err)
		os.Exit(1)
	}

	if err := telegramBot.Start(); err != nil {
		fmt.Println("Error starting telegram bot:", err)
		os.Exit(1)
	}
}

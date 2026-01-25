package ui

import (
	"time"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/visualizer/entities"
)

type Session struct {
	MessageID int
	State     entities.UIState
	Stack     []entities.UIState
	UpdatedAt time.Time
}

func NewSession(messageID int) Session {
	return Session{
		MessageID: messageID,
	}
}

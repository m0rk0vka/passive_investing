package ui

import (
	"financer/pkg/telegram/visualizer/entities"
	"time"
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

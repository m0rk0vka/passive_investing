package ui

import (
	"time"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
)

type Session struct {
	chatID    int64
	messageID int
	State     entities.UIState
	Stack     []entities.UIState
	UpdatedAt time.Time
}

func NewSession(chatID int64) Session {
	return Session{
		chatID: chatID,
	}
}

func (s Session) ChatID() int64 {
	return s.chatID
}

func (s Session) MessageID() int {
	return s.messageID
}

func (s *Session) SetMessageID(messageID int) {
	s.messageID = messageID
}

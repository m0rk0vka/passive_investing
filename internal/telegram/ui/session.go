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

func (s *Session) PushCurrentState() {
	s.Stack = append(s.Stack, s.State)
}

func (s *Session) PopOrHome() {
	n := len(s.Stack)
	if n == 0 {
		s.State = entities.UIState{
			Screen: entities.ScreenHome,
		}
		return
	}

	s.State = s.Stack[n-1]
	s.Stack = s.Stack[:n-1]
}

func (s *Session) SetState(state entities.UIState) {
	s.State = state
}

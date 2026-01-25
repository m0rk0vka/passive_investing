package messagesender

import "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"

type Message struct {
	chatID         int64
	text           string
	inlineKeyboard entities.InlineKeyboardMarkup
}

func NewSimpleMessage(chatID int64, text string) Message {
	return Message{
		chatID: chatID,
		text:   text,
	}
}

func NewInlineKeyboardMessage(chatID int64, text string, inlineKeyboard entities.InlineKeyboardMarkup) Message {
	return Message{
		chatID:         chatID,
		text:           text,
		inlineKeyboard: inlineKeyboard,
	}
}

package entities

import "go.uber.org/zap/zapcore"

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

func NewInlineKeyboardButton(text string, callbackData string) InlineKeyboardButton {
	return InlineKeyboardButton{
		Text:         text,
		CallbackData: callbackData,
	}
}

func NewInlineKeyboardRow(buttons ...InlineKeyboardButton) []InlineKeyboardButton {
	return buttons
}

func NewInlineKeyboardMarkup(rows ...[]InlineKeyboardButton) InlineKeyboardMarkup {
	return InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (m InlineKeyboardMarkup) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("inline_keyboard", zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
		for _, row := range m.InlineKeyboard {
			inner.AppendArray(zapcore.ArrayMarshalerFunc(func(inner zapcore.ArrayEncoder) error {
				for _, button := range row {
					inner.AppendObject(button)
				}
				return nil
			}))
		}
		return nil
	}))
	return nil
}

func (b InlineKeyboardButton) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("text", b.Text)
	encoder.AddString("callback_data", b.CallbackData)
	return nil
}

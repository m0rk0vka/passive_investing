package entities

import "go.uber.org/zap/zapcore"

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler for structured logging
func (u Update) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("update_id", u.UpdateID)

	if u.Message != nil {
		enc.AddObject("message", u.Message)
	}

	if u.CallbackQuery != nil {
		enc.AddObject("callback_query", u.CallbackQuery)
	}

	return nil
}

type Message struct {
	MessageID int       `json:"message_id"`
	From      *User     `json:"from,omitempty"`
	Chat      *Chat     `json:"chat,omitempty"`
	Text      string    `json:"text,omitempty"`
	Document  *Document `json:"document,omitempty"`
	Date      int64     `json:"date,omitempty"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler for structured logging
func (m Message) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("message_id", m.MessageID)
	enc.AddString("text", m.Text)
	enc.AddInt64("date", m.Date)

	if m.From != nil {
		enc.AddObject("from", m.From)
	}

	if m.Chat != nil {
		enc.AddObject("chat", m.Chat)
	}

	if m.Document != nil {
		enc.AddObject("document", m.Document)
	}

	return nil
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler for structured logging
func (u User) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("id", u.ID)
	enc.AddString("username", u.Username)
	enc.AddString("first_name", u.FirstName)
	enc.AddString("last_name", u.LastName)
	return nil
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type,omitempty"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler for structured logging
func (c Chat) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("id", c.ID)
	enc.AddString("type", c.Type)
	return nil
}

type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler for structured logging
func (d Document) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("file_id", d.FileID)
	enc.AddString("file_name", d.FileName)
	enc.AddString("mime_type", d.MimeType)
	enc.AddInt64("file_size", d.FileSize)
	return nil
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler for structured logging
func (cq CallbackQuery) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("id", cq.ID)
	enc.AddString("data", cq.Data)

	if cq.From != nil {
		enc.AddObject("from", cq.From)
	}

	if cq.Message != nil {
		enc.AddObject("message", cq.Message)
	}

	return nil
}

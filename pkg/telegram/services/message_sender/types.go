package messagesender

type sendMessageResponse struct {
	Ok     bool    `json:"ok"`
	Result message `json:"result"`
}

type message struct {
	MessageID int `json:"message_id"`
}

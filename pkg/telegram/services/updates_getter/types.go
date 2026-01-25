package updatesgetter

import "financer/pkg/telegram/entities"

type UpdateResponse struct {
	Ok     bool              `json:"ok"`
	Result []entities.Update `json:"result"`
}

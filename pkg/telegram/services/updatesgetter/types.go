package updatesgetter

import "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"

type UpdateResponse struct {
	Ok     bool              `json:"ok"`
	Result []entities.Update `json:"result"`
}

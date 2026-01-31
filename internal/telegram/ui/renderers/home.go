package renderers

import (
	"context"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

type HomeRenderer struct{}

func (r *HomeRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	text := "FINANCER\n\nВыбери действие:"

	kb := domainEntities.NewInlineKeyboardMarkup(
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Портфели", entities.CBNavPortfolios)),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Создать виртуальный портфель", "todo:vportfolio_create")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Правила пополнения", "todo:funding_rules")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Правила портфеля", "todo:portfolio_rules")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose)),
	)

	return entities.Rendered{Text: text, Kb: kb}, nil
}

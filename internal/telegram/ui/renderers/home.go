package renderers

import (
	"context"

	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/m0rk0vka/passive_investing/pkg/telegram/visualizer/entities"
)

type HomeRenderer struct{}

func (r *HomeRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	text := "FINANCER\n\nВыбери действие:"

	kb := domainEntities.NewInlineKeyboardMarkup(
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Портфели", "todo:portfolios")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Создать виртуальный портфель", "todo:vportfolio_create")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Правила пополнения", "todo:funding_rules")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("Правила портфеля", "todo:portfolio_rules")),
		domainEntities.NewInlineKeyboardRow(domainEntities.NewInlineKeyboardButton("✖️ Закрыть", "todo:close")),
	)

	return entities.Rendered{Text: text, Kb: kb}, nil
}

package renderers

import (
	"context"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/repos"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

type PortfolioSumRenderer struct {
	Repo repos.PortfolioRepo
}

func (r *PortfolioSumRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	summary, err := r.Repo.GetSummary(ctx, userID, st.PortfolioID, st.Period)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to get summary: %w", err)
	}

	var rows [][]domainEntities.InlineKeyboardButton

	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("Позиции", entities.CBNavPositions)))

	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("Периоды", entities.CBNavPeriods)))

	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
		domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
	))

	return entities.Rendered{
		Text: fmt.Sprintf("Сумма портфеля за период %s: %.2f", summary.Period, summary.Total),
		Kb:   domainEntities.NewInlineKeyboardMarkup(rows...),
	}, nil
}

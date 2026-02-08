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
	periods, err := r.Repo.ListPeriods(ctx, userID, st.PortfolioID)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to get periods: %w", err)
	}

	if len(periods) == 0 {
		return entities.Rendered{}, fmt.Errorf("portfolio is empty")
	}

	summary, err := r.Repo.GetSummary(ctx, userID, st.PortfolioID, periods[len(periods)-1])
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to get summary: %w", err)
	}

	var rows [][]domainEntities.InlineKeyboardButton

	if len(periods) != 1 {
		rows = append(rows, domainEntities.NewInlineKeyboardRow(
			domainEntities.NewInlineKeyboardButton("⬅️ Предыдущий период", entities.CBPeriodPrev)))
	}

	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
		domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
	))

	return entities.Rendered{
		Text: fmt.Sprintf("Сумма портфеля за период %s: %.2f", summary.Period, summary.Total),
		Kb:   domainEntities.NewInlineKeyboardMarkup(rows...),
	}, nil
}

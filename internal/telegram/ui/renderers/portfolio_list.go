package renderers

import (
	"context"
	"fmt"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/repos"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

type PortfolioListRenderer struct {
	Repo repos.PortfolioRepo
}

func (r *PortfolioListRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	ps, err := r.Repo.ListPortfolios(ctx, userID)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to list portfolios: %w", err)
	}

	var rows [][]domainEntities.InlineKeyboardButton
	for _, p := range ps {
		title := p.Name
		if strings.TrimSpace(p.Kind) != "" {
			title = title + " (" + p.Kind + ")"
		}
		rows = append(rows, domainEntities.NewInlineKeyboardRow(
			domainEntities.NewInlineKeyboardButton(title, entities.CBOpenPortfolio(p.ID))))
	}

	// nav row
	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
		domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
	))

	text := "Портфели:"
	return entities.Rendered{Text: text, Kb: domainEntities.NewInlineKeyboardMarkup(rows...)}, nil
}

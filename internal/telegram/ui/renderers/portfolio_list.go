package renderers

import (
	"context"
	"fmt"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

// portfolioListService определяет контракт для получения списка портфелей
type portfolioListService interface {
	ListPortfolios(ctx context.Context, userID int64) ([]entities.Portfolio, error)
}

type PortfolioListRenderer struct {
	listService portfolioListService
}

// NewPortfolioListRenderer creates a new portfolio list renderer
func NewPortfolioListRenderer(listService portfolioListService) *PortfolioListRenderer {
	return &PortfolioListRenderer{
		listService: listService,
	}
}

func (r *PortfolioListRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	ps, err := r.listService.ListPortfolios(ctx, userID)
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

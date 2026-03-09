package renderers

import (
	"context"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

// portfolioSummaryService определяет контракт для получения summary портфеля
type portfolioSummaryService interface {
	GetSummary(ctx context.Context, userID int64, portfolioID string, period string) (entities.PortfolioSummary, error)
}

type PortfolioSumRenderer struct {
	summaryService portfolioSummaryService
}

// NewPortfolioSumRenderer creates a new portfolio sum renderer
func NewPortfolioSumRenderer(summaryService portfolioSummaryService) *PortfolioSumRenderer {
	return &PortfolioSumRenderer{
		summaryService: summaryService,
	}
}

func (r *PortfolioSumRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	summary, err := r.summaryService.GetSummary(ctx, userID, st.PortfolioID, st.Period)
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
		Text: fmt.Sprintf("Сумма портфеля за период %s: %s", summary.Period, summary.Total.String()),
		Kb:   domainEntities.NewInlineKeyboardMarkup(rows...),
	}, nil
}

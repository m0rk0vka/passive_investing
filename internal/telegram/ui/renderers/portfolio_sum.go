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

	// Формируем текст с детальной информацией
	text := fmt.Sprintf("📊 Портфель за период %s\n\n", summary.Period)

	// Общая сумма
	text += fmt.Sprintf("💰 Общая сумма: %s\n\n", summary.Total.String())

	// Пополнения
	text += fmt.Sprintf("📥 Пополнения: %s\n", summary.Deposits.String())
	text += fmt.Sprintf("   └ %s%% от общей суммы\n\n", summary.DepositsPct)

	// Заработано
	text += fmt.Sprintf("📈 Заработано: %s\n", summary.Earnings.String())
	text += fmt.Sprintf("   └ %s%% от общей суммы\n", summary.EarningsPct)
	text += fmt.Sprintf("   └ %s%% доходность\n", summary.ReturnPct)

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
		Text: text,
		Kb:   domainEntities.NewInlineKeyboardMarkup(rows...),
	}, nil
}

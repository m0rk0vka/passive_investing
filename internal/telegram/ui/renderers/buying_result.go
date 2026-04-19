package renderers

import (
	"context"
	"fmt"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/services"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/shopspring/decimal"
)

type buyingCalculatorService interface {
	CalculatePurchases(ctx context.Context, portfolioID string, period string, topUp decimal.Decimal) (*services.PurchasePlan, error)
}

type BuyingResultRenderer struct {
	svc buyingCalculatorService
}

func NewBuyingResultRenderer(svc buyingCalculatorService) *BuyingResultRenderer {
	return &BuyingResultRenderer{svc: svc}
}

func (r *BuyingResultRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	topUp, err := decimal.NewFromString(st.TopUpAmount)
	if err != nil || topUp.IsZero() {
		return entities.Rendered{}, fmt.Errorf("invalid top-up amount: %q", st.TopUpAmount)
	}

	plan, err := r.svc.CalculatePurchases(ctx, st.PortfolioID, st.Period, topUp)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to calculate purchases: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🛒 Рекомендации по покупке\n\nСумма пополнения: %s ₽\n\n", formatAmount(plan.TopUpAmount)))

	if len(plan.Items) == 0 {
		sb.WriteString("Нечего покупать — портфель уже в целевой структуре.\n")
	} else {
		for _, item := range plan.Items {
			name := item.Name
			if name == "" {
				name = item.ISIN
			}
			sb.WriteString(fmt.Sprintf("📌 %s — %d шт. (~%s ₽)\n",
				name, item.Shares, formatAmount(item.TotalCost)))
		}
	}

	if !plan.Remaining.IsZero() {
		sb.WriteString(fmt.Sprintf("\nОстаток: %s ₽", formatAmount(plan.Remaining)))
	}

	kb := domainEntities.NewInlineKeyboardMarkup(
		domainEntities.NewInlineKeyboardRow(
			domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
			domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
		),
	)

	return entities.Rendered{Text: sb.String(), Kb: kb}, nil
}

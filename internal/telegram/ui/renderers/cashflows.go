package renderers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/m0rk0vka/passive_investing/internal/services"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

type cashflowHistoryService interface {
	ListDeposits(ctx context.Context, portfolioID string) ([]services.DepositEntry, error)
}

type CashflowsRenderer struct {
	svc cashflowHistoryService
}

func NewCashflowsRenderer(svc cashflowHistoryService) *CashflowsRenderer {
	return &CashflowsRenderer{svc: svc}
}

func (r *CashflowsRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	deposits, err := r.svc.ListDeposits(ctx, st.PortfolioID)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to list deposits: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("📥 История пополнений\n\n")

	if len(deposits) == 0 {
		sb.WriteString("Пополнений пока нет.")
	} else {
		currentMonth := ""
		for _, d := range deposits {
			month := d.Date.Format("2006-01")
			if month != currentMonth {
				currentMonth = month
				sb.WriteString(fmt.Sprintf("\n%s\n", formatMonth(d.Date)))
			}
			sb.WriteString(fmt.Sprintf("  %s — %s %s\n",
				d.Date.Format("02.01"), formatAmount(d.Amount), d.Currency))
		}
	}

	kb := domainEntities.NewInlineKeyboardMarkup(
		domainEntities.NewInlineKeyboardRow(
			domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
			domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
		),
	)

	return entities.Rendered{Text: sb.String(), Kb: kb}, nil
}

var monthNames = [...]string{
	"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
	"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
}

func formatMonth(t time.Time) string {
	return fmt.Sprintf("%s %d", monthNames[t.Month()-1], t.Year())
}

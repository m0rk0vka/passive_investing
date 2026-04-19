package renderers

import (
	"context"
	"fmt"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"github.com/shopspring/decimal"
)

type buyingRulesService interface {
	ListRules(ctx context.Context, portfolioID string) ([]repository.BuyingRule, error)
	IdealAmount(ctx context.Context, portfolioID string) (decimal.Decimal, error)
}

type BuyingRulesRenderer struct {
	svc buyingRulesService
}

func NewBuyingRulesRenderer(svc buyingRulesService) *BuyingRulesRenderer {
	return &BuyingRulesRenderer{svc: svc}
}

func (r *BuyingRulesRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	rules, err := r.svc.ListRules(ctx, st.PortfolioID)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to list rules: %w", err)
	}

	ideal, err := r.svc.IdealAmount(ctx, st.PortfolioID)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to get ideal amount: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("📋 Правила пополнения\n\n")

	if len(rules) == 0 {
		sb.WriteString("Правила не заданы.\nДобавьте записи в таблицу buying_rule.")
	} else {
		sb.WriteString("Целевая структура:\n")
		for _, rule := range rules {
			name := rule.SecurityName
			if name == "" {
				name = rule.ISIN
			}
			sb.WriteString(fmt.Sprintf("• %s — %s%%\n", name, rule.TargetPct.StringFixed(0)))
		}
	}

	sb.WriteString("\nВыберите сумму пополнения:")

	var rows [][]domainEntities.InlineKeyboardButton

	if len(rules) > 0 {
		idealLabel := "Идеальная"
		idealAmount := "0"
		if !ideal.IsZero() {
			idealLabel = fmt.Sprintf("Идеальная (~%s ₽)", formatAmount(ideal))
			idealAmount = ideal.StringFixed(0)
		}

		rows = append(rows,
			domainEntities.NewInlineKeyboardRow(
				domainEntities.NewInlineKeyboardButton("12 000 ₽", entities.CBSelectAmount("12000")),
				domainEntities.NewInlineKeyboardButton("22 000 ₽", entities.CBSelectAmount("22000")),
			),
			domainEntities.NewInlineKeyboardRow(
				domainEntities.NewInlineKeyboardButton(idealLabel, entities.CBSelectAmount(idealAmount)),
			),
		)
	}

	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
		domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
	))

	return entities.Rendered{
		Text: sb.String(),
		Kb:   domainEntities.NewInlineKeyboardMarkup(rows...),
	}, nil
}

func formatAmount(d decimal.Decimal) string {
	// "12000" → "12 000"
	s := d.StringFixed(0)
	if len(s) <= 3 {
		return s
	}
	result := []byte(s)
	out := make([]byte, 0, len(result)+len(result)/3)
	for i, b := range result {
		if i > 0 && (len(result)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, b)
	}
	return string(out)
}

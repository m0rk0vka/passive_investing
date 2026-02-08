package renderers

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/repos"
	domainEntities "github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
)

type PortfolioPositionsRenderer struct {
	Repo repos.PortfolioRepo
}

func (r *PortfolioPositionsRenderer) Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error) {
	periods, err := r.Repo.ListPeriods(ctx, userID, st.PortfolioID)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to get periods: %w", err)
	}

	if len(periods) == 0 {
		return entities.Rendered{}, fmt.Errorf("portfolio is empty")
	}

	positions, err := r.Repo.ListPositions(ctx, userID, st.PortfolioID, st.Period)
	if err != nil {
		return entities.Rendered{}, fmt.Errorf("failed to get summary: %w", err)
	}

	formatedPositions := formatPostions(positions)

	var rows [][]domainEntities.InlineKeyboardButton

	sort.Strings(periods)
	idx := sort.SearchStrings(periods, st.Period)
	if idx != 0 {
		rows = append(rows, domainEntities.NewInlineKeyboardRow(
			domainEntities.NewInlineKeyboardButton("◀️", entities.CBPeriodPrev)))
	}
	if len(periods)-1 != idx {
		rows = append(rows, domainEntities.NewInlineKeyboardRow(
			domainEntities.NewInlineKeyboardButton("▶️", entities.CBPeriodNext)))
	}

	rows = append(rows, domainEntities.NewInlineKeyboardRow(
		domainEntities.NewInlineKeyboardButton("⬅️ Назад", entities.CBBack),
		domainEntities.NewInlineKeyboardButton("✖️ Закрыть", entities.CBClose),
	))

	return entities.Rendered{
		Text: fmt.Sprintf("Позиции на период %s:\n%s", st.Period,
			formatedPositions),
		Kb: domainEntities.NewInlineKeyboardMarkup(rows...),
	}, nil
}

func formatPostions(positions []entities.Position) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Name\tQty\tValue")
	for _, p := range positions {
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Qty, p.Value)
	}
	w.Flush()
	return buf.String()
}

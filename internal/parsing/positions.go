package parsing

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/m0rk0vka/passive_investing/internal/common"
)

var reISIN = regexp.MustCompile(`[A-Z]{2}[A-Z0-9]{9}\d`)

func ParsePositions(rows [][]string) {
	start := -1
	for i, row := range rows {
		if strings.Contains(strings.Join(row, " "), positionAnchor) {
			start = i
			break
		}
	}
	if start == -1 {
		return
	}

	for i := start; i < len(rows); i++ {
		row := rows[i]
		line := strings.Join(row, " ")
		isin := reISIN.FindString(line)
		if isin == "" {
			if common.NormalizeContains(line, cashFlowAnchor) {
				break
			}
			continue
		}

		// name: всё до ISIN (грубо, но для MVP ок)
		name := strings.TrimSpace(strings.Split(line, isin)[0])

		// qty: попробуем найти первое decimal после ISIN в row-ячейках
		qty := decimal.Zero
		gotQty := false
		seenISIN := false
		for _, c := range row {
			if strings.Contains(c, isin) {
				seenISIN = true
				continue
			}
			if !seenISIN {
				continue
			}
			if d, err := parseDec(c); err == nil && !d.IsZero() {
				qty = d
				gotQty = true
				break
			}
		}
		if !gotQty {
			continue
		}

		// market value: берём последнее число в строке (обычно это оценка в руб)
		mv := decimal.Zero
		gotMV := false
		for j := len(row) - 1; j >= 0; j-- {
			if d, err := parseDec(row[j]); err == nil {
				mv = d
				gotMV = true
				break
			}
		}
		if !gotMV {
			continue
		}

		fmt.Printf("name: %s, isin: %s, qty: %s, mv: %s\n", name, isin, qty, mv)
	}
}

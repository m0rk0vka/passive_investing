package parsing

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

func parseDec(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Zero, fmt.Errorf("empty decimal")
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", ".")
	return decimal.NewFromString(s)
}

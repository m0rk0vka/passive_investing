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
	// Сначала заменяем запятые на точки для десятичных дробей
	s = strings.ReplaceAll(s, ",", ".")
	// Удаляем лишние точки (для тысячных разделителей)
	if strings.Count(s, ".") > 1 {
		parts := strings.Split(s, ".")
		s = strings.Join(parts[:len(parts)-1], "") + "." + parts[len(parts)-1]
	}
	return decimal.NewFromString(s)
}

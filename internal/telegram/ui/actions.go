package ui

import "strings"

// Callback data (actions)
const (
	CBNavHome       = "nav:home"
	CBNavPortfolios = "nav:portfolios"

	CBBack  = "back"
	CBClose = "close"

	CBTabPositions = "tab:positions"
	CBPeriodPrev   = "period:prev"
	CBPeriodNext   = "period:next"
)

// Parameterized callbacks
const cbOpenPortfolioPrefix = "open:p:"

func CBOpenPortfolio(id string) string { return cbOpenPortfolioPrefix + id }

func IsOpenPortfolio(data string) (id string, ok bool) {
	if strings.HasPrefix(data, cbOpenPortfolioPrefix) {
		return strings.TrimPrefix(data, cbOpenPortfolioPrefix), true
	}
	return "", false
}

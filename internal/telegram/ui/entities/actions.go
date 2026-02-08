package entities

import "strings"

// Callback data (actions)
const (
	CBNavHome       = "nav:home"
	CBNavPortfolios = "nav:portfolios"
	CBNavPositions  = "nav:positions"
	CBNavPeriods    = "nav:periods"

	CBBack  = "back"
	CBClose = "close"

	CBPeriodPrev = "period:prev"
	CBPeriodNext = "period:next"
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

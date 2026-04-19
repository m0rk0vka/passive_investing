package entities

import "strings"

// Callback data (actions)
const (
	CBNavHome        = "nav:home"
	CBNavPortfolios  = "nav:portfolios"
	CBNavPositions   = "nav:positions"
	CBNavPeriods     = "nav:periods"
	CBNavBuyingRules = "nav:buying_rules"
	CBNavCashflows   = "nav:cashflows"

	CBBack  = "back"
	CBClose = "close"

	CBPeriodPrev = "period:prev"
	CBPeriodNext = "period:next"
)

// Parameterized callbacks
const (
	cbOpenPortfolioPrefix = "open:p:"
	cbSelectAmountPrefix  = "amount:"
)

func CBOpenPortfolio(id string) string { return cbOpenPortfolioPrefix + id }

func IsOpenPortfolio(data string) (id string, ok bool) {
	if strings.HasPrefix(data, cbOpenPortfolioPrefix) {
		return strings.TrimPrefix(data, cbOpenPortfolioPrefix), true
	}
	return "", false
}

func CBSelectAmount(amount string) string { return cbSelectAmountPrefix + amount }

func IsSelectAmount(data string) (amount string, ok bool) {
	if strings.HasPrefix(data, cbSelectAmountPrefix) {
		return strings.TrimPrefix(data, cbSelectAmountPrefix), true
	}
	return "", false
}

package entities

import "time"

// Useful for summaries
type Money struct {
	Amount   string // keep as string for now (later decimal.Decimal)
	Currency string // "RUB"
}

func (m Money) String() string {
	return m.Amount + " " + m.Currency
}

type Portfolio struct {
	ID   string
	Name string
	Kind string // "real"|"virtual"
}

type Position struct {
	ISIN    string
	Name    string
	Qty     string
	Value   Money
	Percent string // процент от общей суммы портфеля, например "15.5"
}

type PortfolioSummary struct {
	PortfolioID string
	Period      string
	Total       Money
	Deposits    Money
	Earnings    Money  // Заработанные деньги (Total - Deposits)
	DepositsPct string // Процент пополнений от общей суммы
	EarningsPct string // Процент заработка от общей суммы
	ReturnPct   string // Процент доходности (Earnings / Deposits * 100)
	AnnualRate  string
	UpdatedAt   time.Time
}

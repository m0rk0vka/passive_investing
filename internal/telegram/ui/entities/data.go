package entities

import "time"

// Useful for summaries
type Money struct {
	Amount   string // keep as string for now (later decimal.Decimal)
	Currency string // "RUB"
}

type Portfolio struct {
	ID   string
	Name string
	Kind string // "real"|"virtual"
}

type Position struct {
	ISIN  string
	Name  string
	Qty   string
	Value Money
}

type PortfolioSummary struct {
	PortfolioID string
	Period      string
	Total       Money
	Deposits    Money
	ReturnPct   string // "0.80"
	UpdatedAt   time.Time
}

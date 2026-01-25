package entities

import (
	"time"

	"github.com/shopspring/decimal"
)

type Portfolio struct {
	ID   string
	Name string

	TotalValue  Money
	Deposits    Money
	Withdrawals Money

	Periods []Period
}

type Period struct {
	LastDayOfMonth time.Time
}

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

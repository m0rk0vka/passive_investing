package entities

import "github.com/shopspring/decimal"

type AllocationItemDTO struct {
	ShortName string
	Value     Money
	Weight    decimal.Decimal
}

type PortfolioAllocationDTO struct {
	PortfolioName string
	Period        string
	Items         []AllocationItemDTO
}

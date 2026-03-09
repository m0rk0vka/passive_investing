package domain

import "github.com/shopspring/decimal"

// PercentageCalculator handles percentage calculations
type PercentageCalculator struct{}

// NewPercentageCalculator creates a new percentage calculator
func NewPercentageCalculator() *PercentageCalculator {
	return &PercentageCalculator{}
}

// CalculatePercentage calculates percentage of value from total
// Returns formatted string with 1 decimal place (e.g., "15.5")
func (c *PercentageCalculator) CalculatePercentage(value, total decimal.Decimal) string {
	if total.IsZero() {
		return "0.0"
	}
	pct := value.Div(total).Mul(decimal.NewFromInt(100))
	return pct.StringFixed(1)
}

// CalculateTotalValue calculates the sum of all values
func (c *PercentageCalculator) CalculateTotalValue(values []decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, value := range values {
		total = total.Add(value)
	}
	return total
}

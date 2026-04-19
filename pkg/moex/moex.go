package moex

import (
	"context"

	"github.com/shopspring/decimal"
)

// Client is the interface for fetching current security prices from MOEX.
// Implement this with a real HTTP client when ready.
type Client interface {
	GetCurrentPrice(ctx context.Context, isin string) (decimal.Decimal, error)
}

// StubClient returns hardcoded prices for local development.
type StubClient struct {
	// Prices maps ISIN → price per share. Add your securities here.
	Prices map[string]decimal.Decimal
}

func NewStubClient(prices map[string]decimal.Decimal) *StubClient {
	return &StubClient{Prices: prices}
}

func (c *StubClient) GetCurrentPrice(_ context.Context, isin string) (decimal.Decimal, error) {
	if p, ok := c.Prices[isin]; ok {
		return p, nil
	}
	// fallback: 100 RUB so the math doesn't break on unknown ISINs
	return decimal.NewFromInt(100), nil
}

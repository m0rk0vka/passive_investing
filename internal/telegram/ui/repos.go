package ui

import (
	"context"
	"financer/pkg/telegram/visualizer/entities"
)

// These are boundaries. Later: postgres implementations.

type PortfolioRepo interface {
	ListPortfolios(ctx context.Context, userID int64) ([]entities.Portfolio, error)

	// months that have data for this portfolio, sorted asc: ["2025-10","2025-11","2025-12"]
	ListPeriods(ctx context.Context, userID int64, portfolioID string) ([]string, error)

	GetSummary(ctx context.Context, userID int64, portfolioID string, period string) (entities.PortfolioSummary, error)
	ListPositions(ctx context.Context, userID int64, portfolioID string, period string) ([]entities.Position, error)
}

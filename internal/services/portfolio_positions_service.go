package services

import (
	"context"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/shopspring/decimal"
)

// Интерфейсы для зависимостей (контракты)
type (
	// positionAggregator определяет контракт для агрегации позиций
	positionAggregator interface {
		AggregatePositionSnapshots(ctx context.Context, accountIDs []int64, period string) ([]repository.PositionSnapshot, error)
	}

	// percentageCalculator определяет контракт для расчета процентов
	percentageCalculator interface {
		CalculatePercentage(value, total decimal.Decimal) string
		CalculateTotalValue(values []decimal.Decimal) decimal.Decimal
	}
)

// PortfolioPositionsService handles portfolio positions operations
type PortfolioPositionsService struct {
	idResolver portfolioIDResolver
	aggregator positionAggregator
	calculator percentageCalculator
}

// NewPortfolioPositionsService creates a new portfolio positions service
func NewPortfolioPositionsService(
	idResolver portfolioIDResolver,
	aggregator positionAggregator,
	calculator percentageCalculator,
) *PortfolioPositionsService {
	return &PortfolioPositionsService{
		idResolver: idResolver,
		aggregator: aggregator,
		calculator: calculator,
	}
}

// ListPositions returns list of positions for a period with calculated percentages
func (s *PortfolioPositionsService) ListPositions(ctx context.Context, userID int64, portfolioID string, period string) ([]entities.Position, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve portfolio ID: %w", err)
	}

	snapshots, err := s.aggregator.AggregatePositionSnapshots(ctx, accountIDs, period)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate positions: %w", err)
	}

	// Extract market values for total calculation
	values := make([]decimal.Decimal, len(snapshots))
	for i, snap := range snapshots {
		values[i] = snap.MarketValue
	}

	// Calculate total value of all positions
	totalValue := s.calculator.CalculateTotalValue(values)

	// Convert to UI entities with calculated percentages
	positions := make([]entities.Position, 0, len(snapshots))
	for _, snap := range snapshots {
		percent := s.calculator.CalculatePercentage(snap.MarketValue, totalValue)

		positions = append(positions, entities.Position{
			ISIN: snap.ISIN,
			Name: snap.SecurityName,
			Qty:  snap.Quantity.String(),
			Value: entities.Money{
				Amount:   snap.MarketValue.String(),
				Currency: snap.Currency,
			},
			Percent: percent,
		})
	}

	return positions, nil
}

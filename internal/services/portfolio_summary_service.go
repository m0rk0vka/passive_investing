package services

import (
	"context"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
)

// Интерфейсы для зависимостей (контракты)
type (
	// valuationAggregator определяет контракт для агрегации оценочных снапшотов
	valuationAggregator interface {
		AggregateValuationSnapshots(ctx context.Context, accountIDs []int64, period string) (*repository.ValuationSnapshot, error)
	}
)

// PortfolioSummaryService handles portfolio summary operations
type PortfolioSummaryService struct {
	idResolver portfolioIDResolver
	aggregator valuationAggregator
}

// NewPortfolioSummaryService creates a new portfolio summary service
func NewPortfolioSummaryService(
	idResolver portfolioIDResolver,
	aggregator valuationAggregator,
) *PortfolioSummaryService {
	return &PortfolioSummaryService{
		idResolver: idResolver,
		aggregator: aggregator,
	}
}

// GetSummary returns portfolio summary for a period (aggregated for virtual portfolios)
func (s *PortfolioSummaryService) GetSummary(ctx context.Context, userID int64, portfolioID string, period string) (entities.PortfolioSummary, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return entities.PortfolioSummary{}, fmt.Errorf("failed to resolve portfolio ID: %w", err)
	}

	snap, err := s.aggregator.AggregateValuationSnapshots(ctx, accountIDs, period)
	if err != nil {
		return entities.PortfolioSummary{}, fmt.Errorf("failed to aggregate snapshots: %w", err)
	}

	// TODO: calculate deposits and return percentage from historical data
	return entities.PortfolioSummary{
		PortfolioID: portfolioID,
		Period:      period,
		Total: entities.Money{
			Amount:   snap.TotalValue.String(),
			Currency: snap.Currency,
		},
		Deposits: entities.Money{
			Amount:   "0",
			Currency: snap.Currency,
		},
		ReturnPct: "0.00",
		UpdatedAt: snap.UpdatedAt,
	}, nil
}

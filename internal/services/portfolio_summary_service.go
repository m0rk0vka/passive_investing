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
	// valuationAggregator определяет контракт для агрегации оценочных снапшотов
	valuationAggregator interface {
		AggregateValuationSnapshots(ctx context.Context, accountIDs []int64, period string) (*repository.ValuationSnapshot, error)
	}

	// depositsInfoProvider определяет контракт для получения информации о пополнениях
	depositsInfoProvider interface {
		GetDepositsInfo(ctx context.Context, accountIDs []int64, upToPeriod string) (*repository.DepositsInfo, error)
	}
)

// PortfolioSummaryService handles portfolio summary operations
type PortfolioSummaryService struct {
	idResolver       portfolioIDResolver
	aggregator       valuationAggregator
	depositsProvider depositsInfoProvider
}

// NewPortfolioSummaryService creates a new portfolio summary service
func NewPortfolioSummaryService(
	idResolver portfolioIDResolver,
	aggregator valuationAggregator,
	depositsProvider depositsInfoProvider,
) *PortfolioSummaryService {
	return &PortfolioSummaryService{
		idResolver:       idResolver,
		aggregator:       aggregator,
		depositsProvider: depositsProvider,
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

	// Получаем информацию о пополнениях
	depositsInfo, err := s.depositsProvider.GetDepositsInfo(ctx, accountIDs, period)
	if err != nil {
		return entities.PortfolioSummary{}, fmt.Errorf("failed to get deposits info: %w", err)
	}

	// Рассчитываем заработанные деньги
	earnings := snap.TotalValue.Sub(depositsInfo.TotalAmount)

	// Рассчитываем проценты
	var depositsPct, earningsPct, returnPct string

	if snap.TotalValue.IsZero() {
		depositsPct = "0.00"
		earningsPct = "0.00"
		returnPct = "0.00"
	} else {
		// Процент пополнений от общей суммы
		depositsPct = depositsInfo.TotalAmount.Div(snap.TotalValue).Mul(decimal.NewFromInt(100)).StringFixed(2)

		// Процент заработка от общей суммы
		earningsPct = earnings.Div(snap.TotalValue).Mul(decimal.NewFromInt(100)).StringFixed(2)

		// Процент доходности (заработок / пополнения * 100)
		if depositsInfo.TotalAmount.IsZero() {
			returnPct = "0.00"
		} else {
			returnPct = earnings.Div(depositsInfo.TotalAmount).Mul(decimal.NewFromInt(100)).StringFixed(2)
		}
	}

	return entities.PortfolioSummary{
		PortfolioID: portfolioID,
		Period:      period,
		Total: entities.Money{
			Amount:   snap.TotalValue.StringFixed(2),
			Currency: snap.Currency,
		},
		Deposits: entities.Money{
			Amount:   depositsInfo.TotalAmount.StringFixed(2),
			Currency: depositsInfo.Currency,
		},
		Earnings: entities.Money{
			Amount:   earnings.StringFixed(2),
			Currency: snap.Currency,
		},
		DepositsPct: depositsPct,
		EarningsPct: earningsPct,
		ReturnPct:   returnPct,
		UpdatedAt:   snap.UpdatedAt,
	}, nil
}

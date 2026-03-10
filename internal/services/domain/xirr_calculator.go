package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/ChizhovVadim/xirr"
	"github.com/m0rk0vka/passive_investing/internal/repository"
)

type ProfitCalculator interface {
	CalculateProfit(ctx context.Context, accountIDs []int64, period string) (string, error)
}

func NewProfitCalculator(
	cashFlowRepo cashFlowRepository,
	aggregateValuationSnapshotsProvider aggregateValuationSnapshotsProvider,
) ProfitCalculator {
	return &profitCalculator{
		cashFlowRepo:                        cashFlowRepo,
		aggregateValuationSnapshotsProvider: aggregateValuationSnapshotsProvider,
	}
}

type profitCalculator struct {
	cashFlowRepo                        cashFlowRepository
	aggregateValuationSnapshotsProvider aggregateValuationSnapshotsProvider
}

type cashFlowRepository interface {
	GetDepositsInfos(ctx context.Context, accountIDs []int64, upToPeriod string) ([]repository.DepositsInfo, error)
}

type aggregateValuationSnapshotsProvider interface {
	AggregateValuationSnapshots(ctx context.Context, accountIDs []int64, period string) (*repository.ValuationSnapshot, error)
}

func (pc *profitCalculator) CalculateProfit(ctx context.Context, accountIDs []int64, period string) (string, error) {
	// 1. Достаем все пополнения и выводы до целевой даты
	flows, err := pc.cashFlowRepo.GetDepositsInfos(ctx, accountIDs, period)
	if err != nil {
		return "", fmt.Errorf("failed to get deposit info: %w", err)
	}

	xirrFlows := make([]xirr.Payment, 0, len(flows))

	// 2. Транслируем реальные движения денег
	// DEPOSIT - отрицательные (вы вкладываете деньги)
	// WITHDRAWAL - положительные (вы получаете деньги обратно)
	for _, f := range flows {
		xirrFlows = append(xirrFlows, xirr.Payment{
			Date:   f.OperationDate,
			Amount: -f.TotalAmount.InexactFloat64(), // Deposits are negative
		})
	}

	// 3. Достаем общую стоимость портфеля (Total Value) на targetDate
	//    (Берем из valuation_snapshot)
	snap, err := pc.aggregateValuationSnapshotsProvider.AggregateValuationSnapshots(ctx, accountIDs, period)
	if err != nil {
		return "", fmt.Errorf("failed to get valuation snapshot: %w", err)
	}

	dateFromPeriod, err := time.Parse("2006-01", period)
	if err != nil {
		return "", fmt.Errorf("failed to parse period: %w", err)
	}

	dateFromPeriod = dateFromPeriod.AddDate(0, 1, -1)

	// 4. ДОБАВЛЯЕМ ВИРТУАЛЬНУЮ ПРОДАЖУ со знаком ПЛЮС
	xirrFlows = append(xirrFlows, xirr.Payment{
		Date:   dateFromPeriod,
		Amount: snap.TotalValue.InexactFloat64(),
	})

	// 5. Считаем!
	rateInfo := xirr.XIRR(xirrFlows)

	// AnnualRate возвращается в долях (например, 1.082 = 8.2% годовых)
	// Вычитаем 1 и умножаем на 100 для отображения в процентах
	annualizedReturn := (rateInfo.AnnualRate - 1.0) * 100
	return fmt.Sprintf("%.2f", annualizedReturn), nil
}

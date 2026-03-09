package domain

import (
	"context"
	"fmt"
	"sort"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/shopspring/decimal"
)

// SnapshotAggregator handles aggregation of snapshots from multiple accounts
type SnapshotAggregator struct {
	snapshotRepo *repository.SnapshotRepository
}

// NewSnapshotAggregator creates a new snapshot aggregator
func NewSnapshotAggregator(snapshotRepo *repository.SnapshotRepository) *SnapshotAggregator {
	return &SnapshotAggregator{
		snapshotRepo: snapshotRepo,
	}
}

// AggregateValuationSnapshots aggregates valuation snapshots from multiple accounts
func (a *SnapshotAggregator) AggregateValuationSnapshots(ctx context.Context, accountIDs []int64, period string) (*repository.ValuationSnapshot, error) {
	if len(accountIDs) == 0 {
		return nil, fmt.Errorf("no accounts provided")
	}

	if len(accountIDs) == 1 {
		return a.snapshotRepo.GetValuationSnapshot(ctx, accountIDs[0], period)
	}

	totalValue := decimal.Zero
	cashBalance := decimal.Zero
	securitiesValue := decimal.Zero

	for _, accountID := range accountIDs {
		snap, err := a.snapshotRepo.GetValuationSnapshot(ctx, accountID, period)
		if err != nil {
			// Skip accounts without data for this period
			continue
		}

		totalValue = totalValue.Add(snap.TotalValue)
		cashBalance = cashBalance.Add(snap.CashBalance)
		securitiesValue = securitiesValue.Add(snap.SecuritiesValue)
	}

	return &repository.ValuationSnapshot{
		Period:          period,
		TotalValue:      totalValue,
		CashBalance:     cashBalance,
		SecuritiesValue: securitiesValue,
		Currency:        "RUB",
	}, nil
}

// AggregatePositionSnapshots aggregates position snapshots from multiple accounts
func (a *SnapshotAggregator) AggregatePositionSnapshots(ctx context.Context, accountIDs []int64, period string) ([]repository.PositionSnapshot, error) {
	if len(accountIDs) == 0 {
		return nil, fmt.Errorf("no accounts provided")
	}

	if len(accountIDs) == 1 {
		return a.snapshotRepo.ListPositionSnapshots(ctx, accountIDs[0], period)
	}

	positionMap := make(map[string]*repository.PositionSnapshot)

	for _, accountID := range accountIDs {
		positions, err := a.snapshotRepo.ListPositionSnapshots(ctx, accountID, period)
		if err != nil {
			// Skip accounts without data
			continue
		}

		for _, pos := range positions {
			if existing, ok := positionMap[pos.ISIN]; ok {
				// Aggregate existing position
				existing.Quantity = existing.Quantity.Add(pos.Quantity)
				existing.MarketValue = existing.MarketValue.Add(pos.MarketValue)
				// Recalculate average price
				if !existing.Quantity.IsZero() {
					existing.Price = existing.MarketValue.Div(existing.Quantity)
				}
			} else {
				// Add new position
				posCopy := pos
				positionMap[pos.ISIN] = &posCopy
			}
		}
	}

	// Convert map to slice
	var result []repository.PositionSnapshot
	for _, pos := range positionMap {
		result = append(result, *pos)
	}

	// Sort by market value descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].MarketValue.GreaterThan(result[j].MarketValue)
	})

	return result, nil
}

// IntersectPeriods returns periods that exist in all accounts
func (a *SnapshotAggregator) IntersectPeriods(ctx context.Context, accountIDs []int64, basePeriods []string) []string {
	if len(accountIDs) <= 1 {
		return basePeriods
	}

	periodSet := make(map[string]int)
	for _, period := range basePeriods {
		periodSet[period] = 1
	}

	for i := 1; i < len(accountIDs); i++ {
		accountPeriods, err := a.snapshotRepo.ListPeriods(ctx, accountIDs[i])
		if err != nil {
			continue
		}

		accountPeriodSet := make(map[string]bool)
		for _, period := range accountPeriods {
			accountPeriodSet[period] = true
		}

		for period := range periodSet {
			if accountPeriodSet[period] {
				periodSet[period]++
			}
		}
	}

	var result []string
	requiredCount := len(accountIDs)
	for _, period := range basePeriods {
		if periodSet[period] == requiredCount {
			result = append(result, period)
		}
	}

	return result
}

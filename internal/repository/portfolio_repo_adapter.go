package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/shopspring/decimal"
)

// PortfolioRepoAdapter adapts repository layer to UI PortfolioRepo interface
type PortfolioRepoAdapter struct {
	db                   *sql.DB
	snapshotRepo         *SnapshotRepository
	accountRepo          *AccountRepository
	virtualPortfolioRepo *VirtualPortfolioRepository
}

// NewPortfolioRepoAdapter creates a new adapter
func NewPortfolioRepoAdapter(db *sql.DB) *PortfolioRepoAdapter {
	return &PortfolioRepoAdapter{
		db:                   db,
		snapshotRepo:         NewSnapshotRepository(db),
		accountRepo:          NewAccountRepository(db),
		virtualPortfolioRepo: NewVirtualPortfolioRepository(db),
	}
}

// ListPortfolios returns list of portfolios (real accounts + virtual portfolios) for user
func (a *PortfolioRepoAdapter) ListPortfolios(ctx context.Context, userID int64) ([]entities.Portfolio, error) {
	var portfolios []entities.Portfolio

	// Get all real accounts (each account = one real portfolio)
	accounts, err := a.accountRepo.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		portfolios = append(portfolios, entities.Portfolio{
			ID:   fmt.Sprintf("real_%d", acc.ID),
			Name: acc.Name,
			Kind: "real",
		})
	}

	// Get all virtual portfolios
	virtualPortfolios, err := a.virtualPortfolioRepo.ListVirtualPortfolios(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, vp := range virtualPortfolios {
		portfolios = append(portfolios, entities.Portfolio{
			ID:   fmt.Sprintf("virtual_%d", vp.ID),
			Name: vp.Name,
			Kind: "virtual",
		})
	}

	return portfolios, nil
}

// ListPeriods returns months that have data for this portfolio
func (a *PortfolioRepoAdapter) ListPeriods(ctx context.Context, userID int64, portfolioID string) ([]string, error) {
	accountIDs, err := a.getAccountIDsForPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	// For virtual portfolios, we need to find common periods across all accounts
	if len(accountIDs) == 0 {
		return []string{}, nil
	}

	// Get periods from first account
	periods, err := a.snapshotRepo.ListPeriods(ctx, accountIDs[0])
	if err != nil {
		return nil, err
	}

	// For virtual portfolios with multiple accounts, intersect periods
	if len(accountIDs) > 1 {
		periods = a.intersectPeriods(ctx, accountIDs, periods)
	}

	return periods, nil
}

// GetLastPeriod returns the most recent period
func (a *PortfolioRepoAdapter) GetLastPeriod(ctx context.Context, userID int64, portfolioID string) (string, error) {
	accountIDs, err := a.getAccountIDsForPortfolio(ctx, portfolioID)
	if err != nil {
		return "", err
	}
	if len(accountIDs) == 0 {
		return "", fmt.Errorf("no accounts found for portfolio %s", portfolioID)
	}

	// For now, use the first account (for virtual portfolios, we'd need to find common periods)
	return a.snapshotRepo.GetLastPeriod(ctx, accountIDs[0])
}

// GetNextPeriod returns the next period after the given one
func (a *PortfolioRepoAdapter) GetNextPeriod(ctx context.Context, userID int64, portfolioID string, period string) (string, error) {
	accountIDs, err := a.getAccountIDsForPortfolio(ctx, portfolioID)
	if err != nil {
		return "", err
	}
	if len(accountIDs) == 0 {
		return "", fmt.Errorf("no accounts found for portfolio %s", portfolioID)
	}

	// For now, use the first account (for virtual portfolios, we'd need to find common periods)
	return a.snapshotRepo.GetNextPeriod(ctx, accountIDs[0], period)
}

// GetPrevPeriod returns the previous period before the given one
func (a *PortfolioRepoAdapter) GetPrevPeriod(ctx context.Context, userID int64, portfolioID string, period string) (string, error) {
	accountIDs, err := a.getAccountIDsForPortfolio(ctx, portfolioID)
	if err != nil {
		return "", err
	}
	if len(accountIDs) == 0 {
		return "", fmt.Errorf("no accounts found for portfolio %s", portfolioID)
	}

	// For now, use the first account (for virtual portfolios, we'd need to find common periods)
	return a.snapshotRepo.GetPrevPeriod(ctx, accountIDs[0], period)
}

// GetSummary returns portfolio summary for a period (aggregated for virtual portfolios)
func (a *PortfolioRepoAdapter) GetSummary(ctx context.Context, userID int64, portfolioID string, period string) (entities.PortfolioSummary, error) {
	accountIDs, err := a.getAccountIDsForPortfolio(ctx, portfolioID)
	if err != nil {
		return entities.PortfolioSummary{}, err
	}

	snap, err := a.aggregateValuationSnapshots(ctx, accountIDs, period)
	if err != nil {
		return entities.PortfolioSummary{}, err
	}

	// TODO: calculate deposits and return percentage from historical data
	// For now, return zeros
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

// ListPositions returns list of positions for a period (aggregated for virtual portfolios)
func (a *PortfolioRepoAdapter) ListPositions(ctx context.Context, userID int64, portfolioID string, period string) ([]entities.Position, error) {
	accountIDs, err := a.getAccountIDsForPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	snapshots, err := a.aggregatePositionSnapshots(ctx, accountIDs, period)
	if err != nil {
		return nil, err
	}

	positions := make([]entities.Position, 0, len(snapshots))
	for _, snap := range snapshots {
		positions = append(positions, entities.Position{
			ISIN: snap.ISIN,
			Name: snap.SecurityName,
			Qty:  snap.Quantity.String(),
			Value: entities.Money{
				Amount:   snap.MarketValue.String(),
				Currency: snap.Currency,
			},
		})
	}

	return positions, nil
}

// getAccountIDsForPortfolio returns account IDs for a portfolio (real or virtual)
func (a *PortfolioRepoAdapter) getAccountIDsForPortfolio(ctx context.Context, portfolioID string) ([]int64, error) {
	// Parse portfolio ID format: "real_123" or "virtual_456"
	var kind string
	var id int64
	_, err := fmt.Sscanf(portfolioID, "%s_%d", &kind, &id)
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %s", portfolioID)
	}

	switch kind {
	case "real":
		// Real portfolio = single account
		return []int64{id}, nil
	case "virtual":
		// Virtual portfolio = multiple accounts
		return a.virtualPortfolioRepo.GetVirtualPortfolioAccounts(ctx, id)
	default:
		return nil, fmt.Errorf("unknown portfolio kind: %s", kind)
	}
}

// intersectPeriods returns periods that exist in all accounts
func (a *PortfolioRepoAdapter) intersectPeriods(ctx context.Context, accountIDs []int64, basePeriods []string) []string {
	if len(accountIDs) <= 1 {
		return basePeriods
	}

	// Create a map of periods from base
	periodSet := make(map[string]int)
	for _, period := range basePeriods {
		periodSet[period] = 1
	}

	// Check each additional account
	for i := 1; i < len(accountIDs); i++ {
		accountPeriods, err := a.snapshotRepo.ListPeriods(ctx, accountIDs[i])
		if err != nil {
			continue
		}

		// Mark periods that exist in this account
		accountPeriodSet := make(map[string]bool)
		for _, period := range accountPeriods {
			accountPeriodSet[period] = true
		}

		// Increment count for periods that exist
		for period := range periodSet {
			if accountPeriodSet[period] {
				periodSet[period]++
			}
		}
	}

	// Return only periods that exist in all accounts
	var result []string
	requiredCount := len(accountIDs)
	for _, period := range basePeriods {
		if periodSet[period] == requiredCount {
			result = append(result, period)
		}
	}

	return result
}

// aggregateValuationSnapshots aggregates valuation snapshots from multiple accounts
func (a *PortfolioRepoAdapter) aggregateValuationSnapshots(ctx context.Context, accountIDs []int64, period string) (*ValuationSnapshot, error) {
	if len(accountIDs) == 0 {
		return nil, fmt.Errorf("no accounts provided")
	}

	// If single account, just return its snapshot
	if len(accountIDs) == 1 {
		return a.snapshotRepo.GetValuationSnapshot(ctx, accountIDs[0], period)
	}

	// Aggregate multiple accounts
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

	return &ValuationSnapshot{
		Period:          period,
		TotalValue:      totalValue,
		CashBalance:     cashBalance,
		SecuritiesValue: securitiesValue,
		Currency:        "RUB",
	}, nil
}

// aggregatePositionSnapshots aggregates position snapshots from multiple accounts
func (a *PortfolioRepoAdapter) aggregatePositionSnapshots(ctx context.Context, accountIDs []int64, period string) ([]PositionSnapshot, error) {
	if len(accountIDs) == 0 {
		return nil, fmt.Errorf("no accounts provided")
	}

	// If single account, just return its positions
	if len(accountIDs) == 1 {
		return a.snapshotRepo.ListPositionSnapshots(ctx, accountIDs[0], period)
	}

	// Aggregate positions by ISIN
	positionMap := make(map[string]*PositionSnapshot)

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
	var result []PositionSnapshot
	for _, pos := range positionMap {
		result = append(result, *pos)
	}

	return result, nil
}

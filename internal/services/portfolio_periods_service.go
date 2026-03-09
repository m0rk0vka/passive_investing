package services

import (
	"context"
	"fmt"
)

// Интерфейсы для зависимостей (контракты)
type (
	// snapshotRepository определяет контракт для работы со снапшотами
	snapshotRepository interface {
		ListPeriods(ctx context.Context, accountID int64) ([]string, error)
		GetLastPeriod(ctx context.Context, accountID int64) (string, error)
		GetNextPeriod(ctx context.Context, accountID int64, currentPeriod string) (string, error)
		GetPrevPeriod(ctx context.Context, accountID int64, currentPeriod string) (string, error)
	}

	// portfolioIDResolver определяет контракт для резолвинга ID портфелей
	portfolioIDResolver interface {
		GetAccountIDs(ctx context.Context, portfolioID string) ([]int64, error)
	}

	// periodIntersector определяет контракт для пересечения периодов
	periodIntersector interface {
		IntersectPeriods(ctx context.Context, accountIDs []int64, basePeriods []string) []string
	}
)

// PortfolioPeriodsService handles portfolio periods operations
type PortfolioPeriodsService struct {
	snapshotRepo snapshotRepository
	idResolver   portfolioIDResolver
	intersector  periodIntersector
}

// NewPortfolioPeriodsService creates a new portfolio periods service
func NewPortfolioPeriodsService(
	snapshotRepo snapshotRepository,
	idResolver portfolioIDResolver,
	intersector periodIntersector,
) *PortfolioPeriodsService {
	return &PortfolioPeriodsService{
		snapshotRepo: snapshotRepo,
		idResolver:   idResolver,
		intersector:  intersector,
	}
}

// ListPeriods returns months that have data for this portfolio
func (s *PortfolioPeriodsService) ListPeriods(ctx context.Context, userID int64, portfolioID string) ([]string, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve portfolio ID: %w", err)
	}

	if len(accountIDs) == 0 {
		return []string{}, nil
	}

	// Get periods from first account
	periods, err := s.snapshotRepo.ListPeriods(ctx, accountIDs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to list periods: %w", err)
	}

	// For virtual portfolios with multiple accounts, intersect periods
	if len(accountIDs) > 1 {
		periods = s.intersector.IntersectPeriods(ctx, accountIDs, periods)
	}

	return periods, nil
}

// GetLastPeriod returns the most recent period
func (s *PortfolioPeriodsService) GetLastPeriod(ctx context.Context, userID int64, portfolioID string) (string, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve portfolio ID: %w", err)
	}
	if len(accountIDs) == 0 {
		return "", fmt.Errorf("no accounts found for portfolio %s", portfolioID)
	}

	period, err := s.snapshotRepo.GetLastPeriod(ctx, accountIDs[0])
	if err != nil {
		return "", fmt.Errorf("failed to get last period: %w", err)
	}

	return period, nil
}

// GetNextPeriod returns the next period after the given one
func (s *PortfolioPeriodsService) GetNextPeriod(ctx context.Context, userID int64, portfolioID string, period string) (string, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve portfolio ID: %w", err)
	}
	if len(accountIDs) == 0 {
		return "", fmt.Errorf("no accounts found for portfolio %s", portfolioID)
	}

	nextPeriod, err := s.snapshotRepo.GetNextPeriod(ctx, accountIDs[0], period)
	if err != nil {
		return "", fmt.Errorf("failed to get next period: %w", err)
	}

	return nextPeriod, nil
}

// GetPrevPeriod returns the previous period before the given one
func (s *PortfolioPeriodsService) GetPrevPeriod(ctx context.Context, userID int64, portfolioID string, period string) (string, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve portfolio ID: %w", err)
	}
	if len(accountIDs) == 0 {
		return "", fmt.Errorf("no accounts found for portfolio %s", portfolioID)
	}

	prevPeriod, err := s.snapshotRepo.GetPrevPeriod(ctx, accountIDs[0], period)
	if err != nil {
		return "", fmt.Errorf("failed to get previous period: %w", err)
	}

	return prevPeriod, nil
}

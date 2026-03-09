package services

import (
	"context"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
)

// Интерфейсы для зависимостей (контракты)
type (
	// accountRepository определяет контракт для работы с аккаунтами
	accountRepository interface {
		ListAccounts(ctx context.Context) ([]repository.Account, error)
	}

	// virtualPortfolioRepository определяет контракт для работы с виртуальными портфелями
	virtualPortfolioRepository interface {
		ListVirtualPortfolios(ctx context.Context, userID int64) ([]repository.VirtualPortfolio, error)
	}
)

// PortfolioListService handles listing portfolios
type PortfolioListService struct {
	accountRepo          accountRepository
	virtualPortfolioRepo virtualPortfolioRepository
}

// NewPortfolioListService creates a new portfolio list service
func NewPortfolioListService(
	accountRepo accountRepository,
	virtualPortfolioRepo virtualPortfolioRepository,
) *PortfolioListService {
	return &PortfolioListService{
		accountRepo:          accountRepo,
		virtualPortfolioRepo: virtualPortfolioRepo,
	}
}

// ListPortfolios returns list of portfolios (real accounts + virtual portfolios) for user
func (s *PortfolioListService) ListPortfolios(ctx context.Context, userID int64) ([]entities.Portfolio, error) {
	var portfolios []entities.Portfolio

	// Get all real accounts (each account = one real portfolio)
	accounts, err := s.accountRepo.ListAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	for _, acc := range accounts {
		portfolios = append(portfolios, entities.Portfolio{
			ID:   fmt.Sprintf("real_%d", acc.ID),
			Name: acc.Name,
			Kind: "real",
		})
	}

	// Get all virtual portfolios
	virtualPortfolios, err := s.virtualPortfolioRepo.ListVirtualPortfolios(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list virtual portfolios: %w", err)
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

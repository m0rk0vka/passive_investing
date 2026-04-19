package services

import (
	"context"
	"fmt"
	"time"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/shopspring/decimal"
)

type DepositEntry struct {
	Date     time.Time
	Amount   decimal.Decimal
	Currency string
}

type allDepositsProvider interface {
	ListAllDeposits(ctx context.Context, accountIDs []int64) ([]repository.DepositsInfo, error)
}

type allAccountsLister interface {
	ListAccounts(ctx context.Context) ([]repository.Account, error)
}

type CashflowHistoryService struct {
	idResolver  portfolioIDResolver
	repo        allDepositsProvider
	accountRepo allAccountsLister
}

func NewCashflowHistoryService(
	idResolver portfolioIDResolver,
	repo allDepositsProvider,
	accountRepo allAccountsLister,
) *CashflowHistoryService {
	return &CashflowHistoryService{idResolver: idResolver, repo: repo, accountRepo: accountRepo}
}

// ListDeposits returns all deposits for the portfolio.
// When portfolioID is empty, returns deposits across all accounts.
func (s *CashflowHistoryService) ListDeposits(ctx context.Context, portfolioID string) ([]DepositEntry, error) {
	var accountIDs []int64

	if portfolioID == "" {
		accounts, err := s.accountRepo.ListAccounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list accounts: %w", err)
		}
		for _, a := range accounts {
			accountIDs = append(accountIDs, a.ID)
		}
	} else {
		ids, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve portfolio: %w", err)
		}
		accountIDs = ids
	}

	infos, err := s.repo.ListAllDeposits(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list deposits: %w", err)
	}

	entries := make([]DepositEntry, 0, len(infos))
	for _, info := range infos {
		entries = append(entries, DepositEntry{
			Date:     info.OperationDate,
			Amount:   info.TotalAmount,
			Currency: info.Currency,
		})
	}
	return entries, nil
}

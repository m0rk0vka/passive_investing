package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/repository"
)

// PortfolioIDResolver resolves portfolio IDs to account IDs
type PortfolioIDResolver struct {
	virtualPortfolioRepo *repository.VirtualPortfolioRepository
}

// NewPortfolioIDResolver creates a new portfolio ID resolver
func NewPortfolioIDResolver(virtualPortfolioRepo *repository.VirtualPortfolioRepository) *PortfolioIDResolver {
	return &PortfolioIDResolver{
		virtualPortfolioRepo: virtualPortfolioRepo,
	}
}

// GetAccountIDs returns account IDs for a portfolio (real or virtual)
func (r *PortfolioIDResolver) GetAccountIDs(ctx context.Context, portfolioID string) ([]int64, error) {
	parts := strings.Split(portfolioID, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid portfolio ID format: %s", portfolioID)
	}

	kind := parts[0]
	var id int64
	if _, err := fmt.Sscanf(parts[1], "%d", &id); err != nil {
		return nil, fmt.Errorf("invalid portfolio ID format: %s", portfolioID)
	}

	switch kind {
	case "real":
		// Real portfolio = single account
		return []int64{id}, nil
	case "virtual":
		// Virtual portfolio = multiple accounts
		return r.virtualPortfolioRepo.GetVirtualPortfolioAccounts(ctx, id)
	default:
		return nil, fmt.Errorf("unknown portfolio kind: %s", kind)
	}
}

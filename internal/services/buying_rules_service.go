package services

import (
	"context"
	"fmt"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/shopspring/decimal"
)

type buyingRuleRepo interface {
	ListByPortfolio(ctx context.Context, portfolioID string) ([]repository.BuyingRule, error)
}

type moexClient interface {
	GetCurrentPrice(ctx context.Context, isin string) (decimal.Decimal, error)
}

type depositsHistoryProvider interface {
	GetDepositsInfos(ctx context.Context, accountIDs []int64, upToPeriod string) ([]repository.DepositsInfo, error)
}

// PurchaseItem is one recommended buy in a purchase plan.
type PurchaseItem struct {
	ISIN          string
	Name          string
	Ticker        string
	Shares        int64
	PricePerShare decimal.Decimal
	TotalCost     decimal.Decimal
}

// PurchasePlan is the full calculation result for a top-up.
type PurchasePlan struct {
	TopUpAmount decimal.Decimal
	Items       []PurchaseItem
	Remaining   decimal.Decimal
}

type BuyingRulesService struct {
	ruleRepo   buyingRuleRepo
	idResolver portfolioIDResolver
	aggregator positionAggregator
	moex       moexClient
	deposits   depositsHistoryProvider
}

func NewBuyingRulesService(
	ruleRepo buyingRuleRepo,
	idResolver portfolioIDResolver,
	aggregator positionAggregator,
	moex moexClient,
	deposits depositsHistoryProvider,
) *BuyingRulesService {
	return &BuyingRulesService{
		ruleRepo:   ruleRepo,
		idResolver: idResolver,
		aggregator: aggregator,
		moex:       moex,
		deposits:   deposits,
	}
}

// ListRules returns the buying rules for the portfolio.
func (s *BuyingRulesService) ListRules(ctx context.Context, portfolioID string) ([]repository.BuyingRule, error) {
	return s.ruleRepo.ListByPortfolio(ctx, portfolioID)
}

// IdealAmount calculates the average deposit amount for a portfolio across all history.
// Returns zero if there are no deposits yet.
func (s *BuyingRulesService) IdealAmount(ctx context.Context, portfolioID string) (decimal.Decimal, error) {
	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to resolve portfolio: %w", err)
	}

	infos, err := s.deposits.GetDepositsInfos(ctx, accountIDs, "9999-12")
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get deposits: %w", err)
	}

	if len(infos) == 0 {
		return decimal.Zero, nil
	}

	total := decimal.Zero
	for _, d := range infos {
		total = total.Add(d.TotalAmount)
	}
	return total.Div(decimal.NewFromInt(int64(len(infos)))).Round(0), nil
}

// CalculatePurchases returns a purchase plan for the given top-up amount.
//
// Algorithm:
//  1. Fetch current positions for the last known period.
//  2. Fetch current MOEX prices for each rule ISIN.
//  3. For each rule: target_amount = (current_total + topUp) * target_pct / 100
//     need_to_buy = max(0, target_amount - current_position_value)
//     shares = floor(need_to_buy / price)
//  4. remaining = topUp - sum(shares * price)
func (s *BuyingRulesService) CalculatePurchases(
	ctx context.Context,
	portfolioID string,
	period string,
	topUp decimal.Decimal,
) (*PurchasePlan, error) {
	rules, err := s.ruleRepo.ListByPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}
	if len(rules) == 0 {
		return &PurchasePlan{TopUpAmount: topUp, Remaining: topUp}, nil
	}

	accountIDs, err := s.idResolver.GetAccountIDs(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve portfolio: %w", err)
	}

	snapshots, err := s.aggregator.AggregatePositionSnapshots(ctx, accountIDs, period)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate positions: %w", err)
	}

	// current value per ISIN
	currentValue := make(map[string]decimal.Decimal, len(snapshots))
	currentTotal := decimal.Zero
	for _, snap := range snapshots {
		currentValue[snap.ISIN] = snap.MarketValue
		currentTotal = currentTotal.Add(snap.MarketValue)
	}

	targetTotal := currentTotal.Add(topUp)
	spent := decimal.Zero
	items := make([]PurchaseItem, 0, len(rules))

	for _, rule := range rules {
		price, err := s.moex.GetCurrentPrice(ctx, rule.ISIN)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %w", rule.ISIN, err)
		}
		if price.IsZero() {
			continue
		}

		targetAmt := targetTotal.Mul(rule.TargetPct).Div(decimal.NewFromInt(100))
		currentAmt := currentValue[rule.ISIN] // zero if not in positions
		diff := targetAmt.Sub(currentAmt)
		if diff.LessThanOrEqual(decimal.Zero) {
			continue
		}

		shares := diff.Div(price).Floor()
		if shares.IsZero() {
			continue
		}

		cost := shares.Mul(price)
		spent = spent.Add(cost)

		sharesInt := shares.IntPart()
		items = append(items, PurchaseItem{
			ISIN:          rule.ISIN,
			Name:          rule.SecurityName,
			Ticker:        rule.Ticker,
			Shares:        sharesInt,
			PricePerShare: price,
			TotalCost:     cost,
		})
	}

	return &PurchasePlan{
		TopUpAmount: topUp,
		Items:       items,
		Remaining:   topUp.Sub(spent),
	}, nil
}

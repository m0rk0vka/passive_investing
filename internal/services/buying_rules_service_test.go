package services_test

import (
	"context"
	"testing"

	"github.com/m0rk0vka/passive_investing/internal/repository"
	"github.com/m0rk0vka/passive_investing/internal/services"
	"github.com/shopspring/decimal"
)

// --- stubs ---

type stubBuyingRuleRepo struct {
	rules []repository.BuyingRule
}

func (s *stubBuyingRuleRepo) ListByPortfolio(_ context.Context, _ string) ([]repository.BuyingRule, error) {
	return s.rules, nil
}

type stubIDResolver struct {
	accountIDs []int64
}

func (s *stubIDResolver) GetAccountIDs(_ context.Context, _ string) ([]int64, error) {
	return s.accountIDs, nil
}

type stubPositionAggregator struct {
	snapshots []repository.PositionSnapshot
}

func (s *stubPositionAggregator) AggregatePositionSnapshots(_ context.Context, _ []int64, _ string) ([]repository.PositionSnapshot, error) {
	return s.snapshots, nil
}

type stubMoexClient struct {
	prices map[string]decimal.Decimal
}

func (s *stubMoexClient) GetCurrentPrice(_ context.Context, isin string) (decimal.Decimal, error) {
	if p, ok := s.prices[isin]; ok {
		return p, nil
	}
	return decimal.NewFromInt(100), nil
}

type stubDepositsHistoryProvider struct {
	infos []repository.DepositsInfo
}

func (s *stubDepositsHistoryProvider) GetDepositsInfos(_ context.Context, _ []int64, _ string) ([]repository.DepositsInfo, error) {
	return s.infos, nil
}

// --- helpers ---

func dec(f float64) decimal.Decimal { return decimal.NewFromFloat(f) }

func newSvc(
	rules []repository.BuyingRule,
	snapshots []repository.PositionSnapshot,
	prices map[string]decimal.Decimal,
) *services.BuyingRulesService {
	return services.NewBuyingRulesService(
		&stubBuyingRuleRepo{rules: rules},
		&stubIDResolver{accountIDs: []int64{1}},
		&stubPositionAggregator{snapshots: snapshots},
		&stubMoexClient{prices: prices},
		&stubDepositsHistoryProvider{},
	)
}

// --- tests ---

// Portfolio: LQDT 50% / OBLG 50%, both empty, top-up 22 000 ₽.
// Price LQDT = 1.52, price OBLG = 14.80.
// Target each = 11 000 ₽.
// LQDT shares = floor(11000 / 1.52) = 7236, cost ≈ 10998.72
// OBLG shares = floor(11000 / 14.80) = 743,  cost ≈ 10996.40
func TestCalculatePurchases_EmptyPortfolio(t *testing.T) {
	rules := []repository.BuyingRule{
		{ISIN: "LQDT", SecurityName: "LQDT", TargetPct: dec(50)},
		{ISIN: "OBLG", SecurityName: "OBLG", TargetPct: dec(50)},
	}
	prices := map[string]decimal.Decimal{
		"LQDT": dec(1.52),
		"OBLG": dec(14.80),
	}

	plan, err := newSvc(rules, nil, prices).CalculatePurchases(context.Background(), "real_1", "2025-10", dec(22000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(plan.Items))
	}

	lqdt := plan.Items[0]
	if lqdt.ISIN != "LQDT" {
		t.Errorf("expected first item LQDT, got %s", lqdt.ISIN)
	}
	if lqdt.Shares != 7236 {
		t.Errorf("LQDT shares: want 7236, got %d", lqdt.Shares)
	}

	oblg := plan.Items[1]
	if oblg.Shares != 743 {
		t.Errorf("OBLG shares: want 743, got %d", oblg.Shares)
	}

	// remaining must be non-negative and less than both prices
	if plan.Remaining.IsNegative() {
		t.Errorf("remaining is negative: %s", plan.Remaining)
	}
}

// Portfolio already has LQDT above its target — should not recommend buying it.
// LQDT current value = 15 000, OBLG = 0.
// Total current = 15 000, top-up = 10 000 → target total = 25 000.
// LQDT target = 12 500 < 15 000 → skip.
// OBLG target = 12 500, need = 12 500, price = 14.80, shares = floor(12500/14.80) = 844.
func TestCalculatePurchases_OneAssetAlreadyOverweight(t *testing.T) {
	rules := []repository.BuyingRule{
		{ISIN: "LQDT", SecurityName: "LQDT", TargetPct: dec(50)},
		{ISIN: "OBLG", SecurityName: "OBLG", TargetPct: dec(50)},
	}
	snapshots := []repository.PositionSnapshot{
		{ISIN: "LQDT", MarketValue: dec(15000), Currency: "RUB"},
	}
	prices := map[string]decimal.Decimal{
		"LQDT": dec(1.52),
		"OBLG": dec(14.80),
	}

	plan, err := newSvc(rules, snapshots, prices).CalculatePurchases(context.Background(), "real_1", "2025-10", dec(10000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Items) != 1 {
		t.Fatalf("expected 1 item (OBLG only), got %d", len(plan.Items))
	}
	if plan.Items[0].ISIN != "OBLG" {
		t.Errorf("expected OBLG, got %s", plan.Items[0].ISIN)
	}
	if plan.Items[0].Shares != 844 {
		t.Errorf("OBLG shares: want 844, got %d", plan.Items[0].Shares)
	}
}

// No rules → whole top-up goes to remaining.
func TestCalculatePurchases_NoRules(t *testing.T) {
	plan, err := newSvc(nil, nil, nil).CalculatePurchases(context.Background(), "real_1", "2025-10", dec(12000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 0 {
		t.Errorf("expected no items, got %d", len(plan.Items))
	}
	if !plan.Remaining.Equal(dec(12000)) {
		t.Errorf("remaining: want 12000, got %s", plan.Remaining)
	}
}

// IdealAmount: average of 3 deposits → (10000 + 15000 + 20000) / 3 = 15000.
func TestIdealAmount(t *testing.T) {
	svc := services.NewBuyingRulesService(
		&stubBuyingRuleRepo{},
		&stubIDResolver{accountIDs: []int64{1}},
		&stubPositionAggregator{},
		&stubMoexClient{},
		&stubDepositsHistoryProvider{infos: []repository.DepositsInfo{
			{TotalAmount: dec(10000)},
			{TotalAmount: dec(15000)},
			{TotalAmount: dec(20000)},
		}},
	)

	amount, err := svc.IdealAmount(context.Background(), "real_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !amount.Equal(dec(15000)) {
		t.Errorf("ideal amount: want 15000, got %s", amount)
	}
}

// IdealAmount with no deposit history → zero.
func TestIdealAmount_NoDeposits(t *testing.T) {
	svc := services.NewBuyingRulesService(
		&stubBuyingRuleRepo{},
		&stubIDResolver{accountIDs: []int64{1}},
		&stubPositionAggregator{},
		&stubMoexClient{},
		&stubDepositsHistoryProvider{},
	)

	amount, err := svc.IdealAmount(context.Background(), "real_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !amount.IsZero() {
		t.Errorf("expected zero, got %s", amount)
	}
}

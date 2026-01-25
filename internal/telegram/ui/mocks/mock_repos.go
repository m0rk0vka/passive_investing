package mocks

import (
	"context"
	"financer/pkg/telegram/visualizer/entities"
	"time"
)

type MockPortfolioRepo struct{}

func (m *MockPortfolioRepo) ListPortfolios(ctx context.Context, userID int64) ([]entities.Portfolio, error) {
	return []entities.Portfolio{
		{ID: "iis", Name: "ВТБ ИИС", Kind: "real"},
		{ID: "bs", Name: "ВТБ БС", Kind: "real"},
		{ID: "iis_bs", Name: "ИИС+БС (вирт.)", Kind: "virtual"},
	}, nil
}

func (m *MockPortfolioRepo) ListPeriods(ctx context.Context, userID int64, portfolioID string) ([]string, error) {
	return []string{"2025-10", "2025-11", "2025-12"}, nil
}

func (m *MockPortfolioRepo) GetSummary(ctx context.Context, userID int64, portfolioID string, period string) (entities.PortfolioSummary, error) {
	return entities.PortfolioSummary{
		PortfolioID: portfolioID,
		Period:      period,
		Total:       entities.Money{Amount: "201 594.23", Currency: "RUB"},
		Deposits:    entities.Money{Amount: "200 000.00", Currency: "RUB"},
		ReturnPct:   "0.80",
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockPortfolioRepo) ListPositions(ctx context.Context, userID int64, portfolioID string, period string) ([]entities.Position, error) {
	return []entities.Position{
		{ISIN: "RU000A10CKT3", Name: "ОФЗ 26251", Qty: "50", Value: entities.Money{Amount: "43 709.73", Currency: "RUB"}},
		{ISIN: "RU000A101EJ5", Name: "EQMX ETF", Qty: "370", Value: entities.Money{Amount: "52 466.00", Currency: "RUB"}},
		{ISIN: "RU000A101NZ2", Name: "GOLD ETF", Qty: "2000", Value: entities.Money{Amount: "5 475.00", Currency: "RUB"}},
		{ISIN: "RU000A1014L8", Name: "LQDT ETF", Qty: "29119", Value: entities.Money{Amount: "54 982.50", Currency: "RUB"}},
		{ISIN: "RU000A1002S8", Name: "OBLG ETF", Qty: "235", Value: entities.Money{Amount: "44 960.20", Currency: "RUB"}},
	}, nil
}

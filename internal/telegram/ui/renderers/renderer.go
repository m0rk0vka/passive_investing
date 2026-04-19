package renderers

import (
	"context"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
)

type Renderer interface {
	Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error)
}

// NewRenderers creates a map of renderers with the given services
func NewRenderers(
	listService portfolioListService,
	summaryService portfolioSummaryService,
	periodsService portfolioPeriodsService,
	positionsService portfolioPositionsService,
	buyingRulesSvc buyingRulesService,
	buyingCalculatorSvc buyingCalculatorService,
	cashflowSvc cashflowHistoryService,
) map[entities.Screen]Renderer {
	return map[entities.Screen]Renderer{
		entities.ScreenHome:               &HomeRenderer{},
		entities.ScreenPortfolioList:      NewPortfolioListRenderer(listService),
		entities.ScreenPortfolioSum:       NewPortfolioSumRenderer(summaryService),
		entities.ScreenPortfolioPositions: NewPortfolioPositionsRenderer(periodsService, positionsService),
		entities.ScreenBuyingRules:        NewBuyingRulesRenderer(buyingRulesSvc),
		entities.ScreenBuyingResult:       NewBuyingResultRenderer(buyingCalculatorSvc),
		entities.ScreenCashflows:          NewCashflowsRenderer(cashflowSvc),
	}
}

// Renderers is kept for backward compatibility (uses mock)
var Renderers = map[entities.Screen]Renderer{
	entities.ScreenHome: &HomeRenderer{},
}

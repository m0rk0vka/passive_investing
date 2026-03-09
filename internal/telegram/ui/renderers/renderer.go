package renderers

import (
	"context"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/repos"
)

type Renderer interface {
	Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error)
}

// NewRenderers creates a map of renderers with the given repository
func NewRenderers(repo repos.PortfolioRepo) map[entities.Screen]Renderer {
	return map[entities.Screen]Renderer{
		entities.ScreenHome: &HomeRenderer{},
		entities.ScreenPortfolioList: &PortfolioListRenderer{
			Repo: repo,
		},
		entities.ScreenPortfolioSum: &PortfolioSumRenderer{
			Repo: repo,
		},
		entities.ScreenPortfolioPositions: &PortfolioPositionsRenderer{
			Repo: repo,
		},
	}
}

// Renderers is kept for backward compatibility (uses mock)
var Renderers = map[entities.Screen]Renderer{
	entities.ScreenHome: &HomeRenderer{},
}

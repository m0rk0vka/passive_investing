package renderers

import (
	"context"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/mocks"
)

type Renderer interface {
	Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error)
}

var Renderers = map[entities.Screen]Renderer{
	entities.ScreenHome: &HomeRenderer{},
	entities.ScreenPortfolioList: &PortfolioListRenderer{
		Repo: &mocks.MockPortfolioRepo{},
	},
}

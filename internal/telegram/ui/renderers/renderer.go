package renderers

import (
	"context"

	"github.com/m0rk0vka/passive_investing/pkg/telegram/visualizer/entities"
)

type Renderer interface {
	Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error)
}

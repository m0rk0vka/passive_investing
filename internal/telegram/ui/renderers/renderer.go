package renderers

import (
	"context"
	"financer/pkg/telegram/visualizer/entities"
)

type Renderer interface {
	Render(ctx context.Context, userID int64, st entities.UIState) (entities.Rendered, error)
}

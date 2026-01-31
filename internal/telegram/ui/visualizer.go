package ui

import (
	"context"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/renderers"
)

type visualizer struct {
	renderers map[entities.Screen]renderers.Renderer
}

func NewVisualizer(renderers map[entities.Screen]renderers.Renderer) *visualizer {
	return &visualizer{
		renderers: renderers,
	}
}

func (v *visualizer) Render(ctx context.Context, chatID int64, state entities.UIState) (entities.Rendered, error) {
	renderer, ok := v.renderers[state.Screen]
	if !ok {
		return entities.Rendered{}, nil
	}
	return renderer.Render(ctx, chatID, state)
}

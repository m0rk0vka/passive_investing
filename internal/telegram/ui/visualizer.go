package ui

import (
	"context"
	"financer/pkg/telegram/visualizer/entities"
	"financer/pkg/telegram/visualizer/renderers"
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

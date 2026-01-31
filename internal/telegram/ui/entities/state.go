package entities

import (
	"github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"go.uber.org/zap/zapcore"
)

type Screen string

const (
	ScreenHome          Screen = "HOME"
	ScreenPortfolioList Screen = "PORTFOLIO_LIST"
	ScreenPortfolioSum  Screen = "PORTFOLIO_SUMMARY"
	ScreenPositions     Screen = "POSITIONS"
)

type UIState struct {
	Screen      Screen
	PortfolioID string
	Period      string
}

type Rendered struct {
	Text string
	Kb   entities.InlineKeyboardMarkup
}

func (r Rendered) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("text", r.Text)
	encoder.AddObject("kb", r.Kb)
	return nil
}

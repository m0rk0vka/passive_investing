package entities

import (
	"github.com/m0rk0vka/passive_investing/pkg/telegram/entities"
	"go.uber.org/zap/zapcore"
)

type Screen string

const (
	ScreenHome               Screen = "HOME"
	ScreenPortfolioList      Screen = "PORTFOLIO_LIST"
	ScreenPortfolioSum       Screen = "PORTFOLIO_SUMMARY"
	ScreenPortfolioPositions Screen = "POSITIONS"
	ScreenBuyingRules        Screen = "BUYING_RULES"
	ScreenBuyingResult       Screen = "BUYING_RESULT"
	ScreenCashflows          Screen = "CASHFLOWS"
)

type UIState struct {
	Screen      Screen
	PortfolioID string
	Period      string
	TopUpAmount string // selected replenishment amount, e.g. "12000"
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

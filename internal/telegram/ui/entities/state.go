package entities

import "financer/pkg/telegram/entities"

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

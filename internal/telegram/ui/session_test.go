package ui_test

import (
	"testing"

	"github.com/m0rk0vka/passive_investing/internal/telegram/ui"
	"github.com/m0rk0vka/passive_investing/internal/telegram/ui/entities"
)

func homeState() entities.UIState    { return entities.UIState{Screen: entities.ScreenHome} }
func portfolioState() entities.UIState {
	return entities.UIState{Screen: entities.ScreenPortfolioList}
}
func summaryState(id, period string) entities.UIState {
	return entities.UIState{Screen: entities.ScreenPortfolioSum, PortfolioID: id, Period: period}
}

// Back from portfolio list → home screen (the bug that was fixed: initial state must be ScreenHome).
func TestPopOrHome_BackFromPortfoliosGoesToHome(t *testing.T) {
	s := ui.NewSession(1)
	s.SetState(homeState())       // as Visualize now sets it
	s.PushCurrentState()          // going to portfolios
	s.SetState(portfolioState())

	s.PopOrHome()

	if s.State.Screen != entities.ScreenHome {
		t.Errorf("expected ScreenHome after back, got %q", s.State.Screen)
	}
}

// Empty stack → PopOrHome always returns ScreenHome.
func TestPopOrHome_EmptyStackAlwaysHome(t *testing.T) {
	s := ui.NewSession(1)
	s.SetState(summaryState("real_1", "2025-10"))

	s.PopOrHome()

	if s.State.Screen != entities.ScreenHome {
		t.Errorf("expected ScreenHome, got %q", s.State.Screen)
	}
}

// Multi-level navigation: home → portfolios → summary → back → back → home.
func TestPopOrHome_MultiLevel(t *testing.T) {
	s := ui.NewSession(1)
	s.SetState(homeState())

	s.PushCurrentState()
	s.SetState(portfolioState())

	s.PushCurrentState()
	s.SetState(summaryState("real_1", "2025-10"))

	// first back → portfolio list
	s.PopOrHome()
	if s.State.Screen != entities.ScreenPortfolioList {
		t.Errorf("expected PortfolioList, got %q", s.State.Screen)
	}

	// second back → home
	s.PopOrHome()
	if s.State.Screen != entities.ScreenHome {
		t.Errorf("expected Home, got %q", s.State.Screen)
	}
}

// Portfolio ID and Period are preserved when popping.
func TestPopOrHome_PreservesPortfolioState(t *testing.T) {
	s := ui.NewSession(1)
	s.SetState(summaryState("virtual_3", "2025-09"))

	s.PushCurrentState()
	s.SetState(entities.UIState{
		Screen:      entities.ScreenPortfolioPositions,
		PortfolioID: "virtual_3",
		Period:      "2025-09",
	})

	s.PopOrHome()

	if s.State.PortfolioID != "virtual_3" {
		t.Errorf("PortfolioID: want virtual_3, got %s", s.State.PortfolioID)
	}
	if s.State.Period != "2025-09" {
		t.Errorf("Period: want 2025-09, got %s", s.State.Period)
	}
}

// TopUpAmount is preserved through push/pop cycle.
func TestPopOrHome_PreservesTopUpAmount(t *testing.T) {
	s := ui.NewSession(1)
	s.SetState(entities.UIState{
		Screen:      entities.ScreenBuyingRules,
		PortfolioID: "real_1",
		Period:      "2025-10",
	})

	s.PushCurrentState()
	s.SetState(entities.UIState{
		Screen:      entities.ScreenBuyingResult,
		PortfolioID: "real_1",
		Period:      "2025-10",
		TopUpAmount: "22000",
	})

	s.PopOrHome()

	if s.State.TopUpAmount != "" {
		// popped back to BuyingRules which has no TopUpAmount set
		t.Errorf("expected empty TopUpAmount after pop, got %s", s.State.TopUpAmount)
	}
	if s.State.Screen != entities.ScreenBuyingRules {
		t.Errorf("expected ScreenBuyingRules, got %q", s.State.Screen)
	}
}

package menu

import "fmt"

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

func BuildKeyboard(st DashboardState) InlineKeyboardMarkup {
	prev := shiftMonth(st.Month, -1)
	next := shiftMonth(st.Month, +1)

	return InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{Text: "◀️", CallbackData: encodeState(DashboardState{Portfolio: st.Portfolio, Month: prev, Tab: st.Tab})},
				{Text: st.Month, CallbackData: encodeState(st)}, // можно сделать открытие списка месяцев позже
				{Text: "▶️", CallbackData: encodeState(DashboardState{Portfolio: st.Portfolio, Month: next, Tab: st.Tab})},
			},
			{
				{Text: "Портфель: ИИС+БС", CallbackData: encodeState(DashboardState{Portfolio: st.Portfolio, Month: st.Month, Tab: st.Tab})},
			},
			{
				{Text: "Сводка", CallbackData: encodeState(DashboardState{Portfolio: st.Portfolio, Month: st.Month, Tab: "summary"})},
				{Text: "Состав", CallbackData: encodeState(DashboardState{Portfolio: st.Portfolio, Month: st.Month, Tab: "alloc"})},
				{Text: "План", CallbackData: encodeState(DashboardState{Portfolio: st.Portfolio, Month: st.Month, Tab: "plan"})},
			},
		},
	}
}

func encodeState(st DashboardState) string {
	return fmt.Sprintf("d|p=%s|m=%s|tab=%s", st.Portfolio, st.Month, st.Tab)
}

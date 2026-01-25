package menu

import "strings"

type DashboardState struct {
	Portfolio string // "iis_bs"
	Month     string // "2025-12"
	Tab       string // "summary"|"alloc"|"plan"
}

type DashboardDTO struct {
	Title string
	Lines []string
}

func BuildDashboardDTO(st DashboardState) DashboardDTO {
	// тестовые данные
	switch st.Tab {
	case "alloc":
		return DashboardDTO{
			Title: "ИИС+БС / " + st.Month + " / Состав",
			Lines: []string{
				"Акции (EQMX): 60%  (120 000 ₽)",
				"Облигации:     40%  (80 000 ₽)",
				"  - OBLG:      13.3% (26 700 ₽)",
				"  - LQDT:      13.3% (26 700 ₽)",
				"  - ОФЗ:       13.3% (26 600 ₽)",
			},
		}
	case "plan":
		return DashboardDTO{
			Title: "ИИС+БС / " + st.Month + " / План покупок",
			Lines: []string{
				"Пополнение: ИИС 22 000 ₽, БС 56 000 ₽",
				"ИИС: купить EQMX 1 шт",
				"БС:  купить OBLG 1 шт, LQDT 5000 шт",
				"Остаток кэша:  312 ₽",
			},
		}
	default:
		return DashboardDTO{
			Title: "ИИС+БС / " + st.Month + " / Сводка",
			Lines: []string{
				"Баланс:     201 594.23 ₽",
				"Пополнения: 200 000.00 ₽",
				"Доходн.:    0.80%",
			},
		}
	}
}

func FormatDashboardMarkdown(dto DashboardDTO) string {
	// MarkdownV2 сложнее экранировать, поэтому для MVP проще обычный Markdown или plain text.
	// Тут сделаем plain text с моноширинным блоком.
	return dto.Title + "\n\n" + "```\n" + strings.Join(dto.Lines, "\n") + "\n```"
}

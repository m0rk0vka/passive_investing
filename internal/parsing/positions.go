package parsing

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/m0rk0vka/passive_investing/internal/common"
)

var reISIN = regexp.MustCompile(`[A-Z]{2}[A-Z0-9]{9}\d`)
var rePeriod = regexp.MustCompile(`с (\d{2}\.\d{2}\.\d{4}) по (\d{2}\.\d{2}\.\d{4})`)
var reAccount = regexp.MustCompile(`№ субсчета:\s+(\w+)`)
var reClientName = regexp.MustCompile(`клиент:\s+(.+)`)
var reClientINN = regexp.MustCompile(`инн:\s+(\d+)`)
var reReportDate = regexp.MustCompile(`дата формирования отчета\s+(\d{2}\.\d{2}\.\d{4})`)

// PortfolioData содержит все извлеченные данные из отчета
type PortfolioData struct {
	PeriodStart    time.Time
	PeriodEnd      time.Time
	AccountNumber  string
	ClientName     string
	ClientINN      string
	CashBalance    decimal.Decimal
	TotalAssets    decimal.Decimal
	Positions      []Position
	CashFlow       []CashFlowOperation
	SecuritiesFlow []SecuritiesFlowOperation
	ReportDate     time.Time
}

// Position представляет информацию о ценной бумаге
type Position struct {
	ISIN        string
	Name        string
	Qty         decimal.Decimal
	Price       decimal.Decimal
	MarketValue decimal.Decimal
	Currency    string
}

// CashFlowOperation представляет операцию движения денежных средств
type CashFlowOperation struct {
	Date          time.Time
	Amount        decimal.Decimal
	Currency      string
	OperationType string
	Comment       string
}

// SecuritiesFlowOperation представляет операцию движения ценных бумаг
type SecuritiesFlowOperation struct {
	SecurityName  string
	ISIN          string
	Platform      string
	Date          time.Time
	Qty           decimal.Decimal
	OperationType string
	Comment       string
}

func ParsePositions(rows [][]string) *PortfolioData {
	data := &PortfolioData{
		Positions:      make([]Position, 0),
		CashFlow:       make([]CashFlowOperation, 0),
		SecuritiesFlow: make([]SecuritiesFlowOperation, 0),
	}

	// Нормализуем данные для извлечения общей информации
	normalizedRows := common.NormalizeRows(rows)

	// Извлекаем общую информацию
	extractGeneralInfo(normalizedRows, data)

	// Извлекаем позиции (используем исходные данные)
	extractPositions(rows, data)

	// Извлекаем движение денежных средств (используем исходные данные)
	extractCashFlow(rows, data)

	// Извлекаем движение ценных бумаг (используем исходные данные)
	extractSecuritiesFlow(rows, data)

	return data
}

func extractGeneralInfo(rows [][]string, data *PortfolioData) {
	for _, row := range rows {
		line := strings.Join(row, " ")

		// Период
		if matches := rePeriod.FindStringSubmatch(line); len(matches) == 3 {
			if start, err := time.Parse("02.01.2006", matches[1]); err == nil {
				data.PeriodStart = start
			}
			if end, err := time.Parse("02.01.2006", matches[2]); err == nil {
				data.PeriodEnd = end
			}
		}

		// Номер субсчета
		if matches := reAccount.FindStringSubmatch(line); len(matches) == 2 {
			data.AccountNumber = matches[1]
		}

		// Имя клиента
		if matches := reClientName.FindStringSubmatch(line); len(matches) == 2 {
			data.ClientName = strings.TrimSpace(matches[1])
		}

		// ИНН клиента
		if matches := reClientINN.FindStringSubmatch(line); len(matches) == 2 {
			data.ClientINN = matches[1]
		}

		// Дата формирования отчета
		if matches := reReportDate.FindStringSubmatch(line); len(matches) == 2 {
			if date, err := time.Parse("02.01.2006", matches[1]); err == nil {
				data.ReportDate = date
			}
		}

		// Извлекаем остатки денежных средств из сводной информации
		if strings.Contains(line, "исходящий остаток денежных средств") {
			// Ищем число в строке после "исходящий остаток денежных средств"
			parts := strings.Split(line, "исходящий остаток денежных средств")
			if len(parts) > 1 {
				remaining := strings.TrimSpace(parts[1])
				words := strings.Fields(remaining)
				for _, word := range words {
					if amount, err := parseDec(word); err == nil {
						data.CashBalance = amount
						break
					}
				}
			}
		}

		// Извлекаем общую оценку активов
		if strings.Contains(line, "оценка активов по курсу (в том числе денежные средства)") {
			// Ищем число в строке после "оценка активов по курсу"
			parts := strings.Split(line, "оценка активов по курсу (в том числе денежные средства)")
			if len(parts) > 1 {
				remaining := strings.TrimSpace(parts[1])
				words := strings.Fields(remaining)
				for _, word := range words {
					if amount, err := parseDec(word); err == nil {
						data.TotalAssets = amount
						break
					}
				}
			}
		}

		// Дополнительный способ извлечения общей оценки активов
		if strings.Contains(line, "оценка активов по курсу") && data.TotalAssets.IsZero() {
			words := strings.Fields(line)
			for i, word := range words {
				if word == "курс" && i+1 < len(words) {
					if amount, err := parseDec(words[i+1]); err == nil {
						data.TotalAssets = amount
						break
					}
				}
			}
		}
	}
}

func extractPositions(rows [][]string, data *PortfolioData) {
	start := -1
	for i, row := range rows {
		if strings.Contains(strings.Join(row, " "), positionAnchor) {
			start = i
			break
		}
	}
	if start == -1 {
		return
	}

	// Находим заголовок таблицы для определения индексов колонок
	headerRow := -1
	priceCol := -1
	marketValueCol := -1
	qtyCol := -1

	for i := start; i < len(rows) && i < start+5; i++ {
		row := rows[i]
		for j, cell := range row {
			cellLower := strings.ToLower(cell)
			if strings.Contains(cellLower, "цена") && !strings.Contains(cellLower, "номинала") {
				priceCol = j
			}
			if strings.Contains(cellLower, "оценка исходящего остатка") &&
				strings.Contains(cellLower, "руб") {
				marketValueCol = j
			}
			if strings.Contains(cellLower, "исходящий остаток") &&
				strings.Contains(cellLower, "шт") {
				qtyCol = j
			}
		}
		if priceCol != -1 && marketValueCol != -1 && qtyCol != -1 {
			headerRow = i
			break
		}
	}

	for i := start; i < len(rows); i++ {
		row := rows[i]
		line := strings.Join(row, " ")
		isin := reISIN.FindString(line)
		if isin == "" {
			if common.NormalizeContains(line, cashFlowAnchor) {
				break
			}
			continue
		}

		// Пропускаем заголовочную строку
		if i == headerRow {
			continue
		}

		// name: всё до ISIN (грубо, но для MVP ок)
		name := strings.TrimSpace(strings.Split(line, isin)[0])

		// qty: используем найденную колонку количества
		qty := decimal.Zero
		if qtyCol != -1 && qtyCol < len(row) {
			if d, err := parseDec(row[qtyCol]); err == nil {
				qty = d
			}
		}

		// Если не нашли колонку количества, ищем первое число после ISIN
		if qty.IsZero() {
			seenISIN := false
			for _, c := range row {
				if strings.Contains(c, isin) {
					seenISIN = true
					continue
				}
				if !seenISIN {
					continue
				}
				if d, err := parseDec(c); err == nil && !d.IsZero() {
					qty = d
					break
				}
			}
		}

		if qty.IsZero() {
			continue
		}

		// price: используем найденную колонку цены
		price := decimal.Zero
		if priceCol != -1 && priceCol < len(row) {
			if d, err := parseDec(row[priceCol]); err == nil {
				price = d
			}
		}

		// market value: используем найденную колонку рыночной стоимости
		mv := decimal.Zero
		if marketValueCol != -1 && marketValueCol < len(row) {
			if d, err := parseDec(row[marketValueCol]); err == nil {
				mv = d
			}
		}

		// Если не нашли колонки, используем старый метод
		if mv.IsZero() {
			for j := len(row) - 1; j >= 0; j-- {
				if d, err := parseDec(row[j]); err == nil {
					mv = d
					break
				}
			}
		}

		// Определяем валюту (обычно RUR)
		currency := "RUR"
		for _, c := range row {
			if strings.ToLower(c) == "rur" || strings.ToLower(c) == "rub" {
				currency = strings.ToUpper(c)
				break
			}
		}

		position := Position{
			ISIN:        isin,
			Name:        name,
			Qty:         qty,
			Price:       price,
			MarketValue: mv,
			Currency:    currency,
		}

		data.Positions = append(data.Positions, position)
	}
}

func extractCashFlow(rows [][]string, data *PortfolioData) {
	start := -1
	for i, row := range rows {
		line := strings.Join(row, " ")
		if strings.Contains(line, "Движение денежных средств") || strings.Contains(line, "движение денежных средств") {
			start = i
			fmt.Printf("НАШЛИ РАЗДЕЛ ДВИЖЕНИЯ ДЕНЕЖНЫХ СРЕДСТВ НА СТРОКЕ %d\n", i)
			break
		}
	}
	if start == -1 {
		fmt.Println("РАЗДЕЛ ДВИЖЕНИЯ ДЕНЕЖНЫХ СРЕДСТВ НЕ НАЙДЕН")
		return
	}

	// Пропускаем заголовок и ищем данные
	for i := start + 2; i < len(rows); i++ {
		row := rows[i]
		line := strings.Join(row, " ")

		if len(row) < 3 {
			continue
		}

		// Пропускаем строку "Основной рынок"
		if strings.Contains(line, "основной рынок") {
			continue
		}

		// Проверяем, что это строка с данными (содержит дату)
		// Дата находится во втором элементе (индекс 1)
		if len(row) > 1 {
			if date, err := time.Parse("02.01.2006", row[1]); err == nil {
				amount, _ := parseDec(row[2])

				// Валюта и тип операции могут быть в разных позициях
				currency := ""
				operationType := ""
				comment := ""

				// Ищем валюту (обычно RUR)
				for j := 3; j < len(row); j++ {
					if strings.ToUpper(row[j]) == "RUR" || strings.ToUpper(row[j]) == "RUB" {
						currency = strings.ToUpper(row[j])
						break
					}
				}

				// Ищем тип операции
				for j := 3; j < len(row); j++ {
					if row[j] != "" && !strings.Contains(row[j], "Перечисление") {
						operationType = row[j]
						break
					}
				}

				// Комментарий - последний непустой элемент
				for j := len(row) - 1; j >= 0; j-- {
					if row[j] != "" {
						comment = row[j]
						break
					}
				}

				op := CashFlowOperation{
					Date:          date,
					Amount:        amount,
					Currency:      currency,
					OperationType: operationType,
					Comment:       comment,
				}

				data.CashFlow = append(data.CashFlow, op)
			}
		}
	}
}

func extractSecuritiesFlow(rows [][]string, data *PortfolioData) {
	start := -1
	for i, row := range rows {
		line := strings.Join(row, " ")
		if strings.Contains(line, "Движение ценных бумаг") || strings.Contains(line, "движение ценных бумаг") {
			start = i
			break
		}
	}
	if start == -1 {
		return
	}

	// Анализируем заголовок таблицы для определения порядка колонок
	headerRow := -1
	var columnOrder []string // порядок колонок

	for i := start; i < len(rows) && i < start+5; i++ {
		row := rows[i]
		line := strings.Join(row, " ")

		// Ищем строку с заголовками
		if strings.Contains(line, "Наименование ценной бумаги") && strings.Contains(line, "Дата операции") {
			headerRow = i

			// Определяем порядок колонок по непустым значениям в заголовке
			for _, cell := range row {
				cellLower := strings.ToLower(cell)
				if strings.Contains(cellLower, "наименование") {
					columnOrder = append(columnOrder, "name")
				} else if strings.Contains(cellLower, "площадка") {
					columnOrder = append(columnOrder, "platform")
				} else if strings.Contains(cellLower, "дата операции") {
					columnOrder = append(columnOrder, "date")
				} else if strings.Contains(cellLower, "количество") {
					columnOrder = append(columnOrder, "quantity")
				} else if strings.Contains(cellLower, "тип операции") {
					columnOrder = append(columnOrder, "operation")
				} else if strings.Contains(cellLower, "комментарий") {
					columnOrder = append(columnOrder, "comment")
				}
			}

			break
		}
	}

	if headerRow == -1 {
		return
	}

	// Ищем данные в строках после заголовка
	for i := headerRow + 1; i < len(rows); i++ {
		row := rows[i]
		line := strings.Join(row, " ")

		// Пропускаем пустые строки
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Проверяем, что строка содержит ISIN
		isin := reISIN.FindString(line)
		if isin == "" {
			continue
		}

		// Собираем непустые значения из строки в том же порядке, что и колонки
		var values []string
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				values = append(values, strings.TrimSpace(cell))
			}
		}

		// Извлекаем данные согласно порядку колонок
		name := ""
		platform := ""
		var date time.Time
		var qty decimal.Decimal
		operationType := ""
		comment := ""

		for i, colType := range columnOrder {
			if i < len(values) {
				value := values[i]
				switch colType {
				case "name":
					// Извлекаем имя до ISIN
					name = strings.TrimSpace(strings.Split(value, isin)[0])
				case "platform":
					platform = value
				case "date":
					if d, err := time.Parse("02.01.2006", value); err == nil {
						date = d
					}
				case "quantity":
					if q, err := parseDec(value); err == nil {
						qty = q
					}
				case "operation":
					operationType = value
				case "comment":
					comment = value
				}
			}
		}

		// Создаем операцию только если дата корректная
		if !date.IsZero() {
			op := SecuritiesFlowOperation{
				SecurityName:  name,
				ISIN:          isin,
				Platform:      platform,
				Date:          date,
				Qty:           qty,
				OperationType: operationType,
				Comment:       comment,
			}

			data.SecuritiesFlow = append(data.SecuritiesFlow, op)
		}
	}
}

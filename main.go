package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/m0rk0vka/passive_investing/internal/parsing"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("no input filename")
		os.Exit(1)
	}
	f, err := excelize.OpenFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Убираем отладочный вывод

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	data := parsing.ParsePositions(rows)

	// Выводим извлеченные данные
	fmt.Printf("Период: %s - %s\n", data.PeriodStart.Format("02.01.2006"), data.PeriodEnd.Format("02.01.2006"))
	fmt.Printf("Субсчет: %s\n", data.AccountNumber)
	fmt.Printf("Клиент: %s (ИНН: %s)\n", data.ClientName, data.ClientINN)
	fmt.Printf("Дата отчета: %s\n", data.ReportDate.Format("02.01.2006"))
	fmt.Printf("Остаток денежных средств: %s\n", data.CashBalance.String())
	fmt.Printf("Общая оценка активов: %s\n", data.TotalAssets.String())

	fmt.Println("\nПозиции:")
	for _, pos := range data.Positions {
		fmt.Printf("  %s (%s): %s шт. × %s = %s %s\n",
			pos.Name, pos.ISIN, pos.Qty.String(), pos.Price.String(),
			pos.MarketValue.String(), pos.Currency)
	}

	fmt.Println("\nДвижение денежных средств:")
	for _, op := range data.CashFlow {
		fmt.Printf("  %s: %s %s (%s) - %s\n",
			op.Date.Format("02.01.2006"), op.Amount.String(), op.Currency,
			op.OperationType, op.Comment)
	}

	fmt.Println("\nДвижение ценных бумаг:")
	for _, op := range data.SecuritiesFlow {
		fmt.Printf("  %s: %s %s шт. (%s) - %s\n",
			op.Date.Format("02.01.2006"), op.SecurityName, op.Qty.String(),
			op.OperationType, op.Comment)
	}
}

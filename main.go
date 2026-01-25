package main

import (
	"financer/internal/parsing"
	"fmt"
	"os"

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
	// rows = common.NormalizeRows(rows)
	// for _, row := range rows {
	// 	for _, cell := range row {
	// 		fmt.Print(cell, "\t")
	// 	}
	// 	fmt.Println()
	// }

	parsing.ParsePositions(rows)
}

package parsing

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

var updateGolden = flag.Bool("update-golden", false, "update golden test files")

// anonymizePortfolioData заменяет чувствительные данные на хешированные версии
func anonymizePortfolioData(data *PortfolioData) *PortfolioData {
	result := *data

	// Анонимизируем номер счета
	if result.AccountNumber != "" {
		result.AccountNumber = hashString(result.AccountNumber)
	}

	// Анонимизируем имя клиента
	if result.ClientName != "" {
		result.ClientName = hashString(result.ClientName)
	}

	// Анонимизируем ИНН
	if result.ClientINN != "" {
		result.ClientINN = hashString(result.ClientINN)
	}

	// Анонимизируем комментарии в CashFlow
	for i := range result.CashFlow {
		if result.CashFlow[i].Comment != "" {
			result.CashFlow[i].Comment = hashString(result.CashFlow[i].Comment)
		}
	}

	// Анонимизируем комментарии в SecuritiesFlow
	for i := range result.SecuritiesFlow {
		if result.SecuritiesFlow[i].Comment != "" {
			result.SecuritiesFlow[i].Comment = hashString(result.SecuritiesFlow[i].Comment)
		}
	}

	return &result
}

// hashString создает короткий хеш строки для анонимизации
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("ANON_%x", h[:8])
}

func TestParsePositions_GoldenFiles(t *testing.T) {
	// Получаем список всех xlsx файлов в data/raw
	rawDir := "../../data/raw"
	files, err := filepath.Glob(filepath.Join(rawDir, "*.xlsx"))
	require.NoError(t, err, "failed to list xlsx files")
	require.NotEmpty(t, files, "no xlsx files found in data/raw")

	for _, filePath := range files {
		fileName := filepath.Base(filePath)
		t.Run(fileName, func(t *testing.T) {
			// Открываем xlsx файл
			f, err := excelize.OpenFile(filePath)
			require.NoError(t, err, "failed to open xlsx file")
			defer f.Close()

			// Получаем строки из первого листа
			sheetName := f.GetSheetName(0)
			rows, err := f.GetRows(sheetName)
			require.NoError(t, err, "failed to get rows")

			// Парсим данные
			data := ParsePositions(rows)
			require.NotNil(t, data, "parsed data should not be nil")

			// Анонимизируем чувствительные данные
			anonData := anonymizePortfolioData(data)

			// Путь к golden файлу
			goldenPath := filepath.Join("../../testdata/golden", fileName+".json")

			if *updateGolden {
				// Режим обновления golden файлов
				jsonData, err := json.MarshalIndent(anonData, "", "  ")
				require.NoError(t, err, "failed to marshal data")

				err = os.MkdirAll(filepath.Dir(goldenPath), 0755)
				require.NoError(t, err, "failed to create golden directory")

				err = os.WriteFile(goldenPath, jsonData, 0644)
				require.NoError(t, err, "failed to write golden file")

				t.Logf("Updated golden file: %s", goldenPath)
			} else {
				// Режим проверки
				goldenData, err := os.ReadFile(goldenPath)
				require.NoError(t, err, "failed to read golden file (run with -update-golden to create)")

				var expected PortfolioData
				err = json.Unmarshal(goldenData, &expected)
				require.NoError(t, err, "failed to unmarshal golden data")

				// Сравниваем результаты (анонимизированные)
				actualJSON, _ := json.MarshalIndent(anonData, "", "  ")
				expectedJSON, _ := json.MarshalIndent(&expected, "", "  ")

				assert.JSONEq(t, string(expectedJSON), string(actualJSON),
					"parsed data does not match golden file")
			}
		})
	}
}

func TestParsePositions_BasicValidation(t *testing.T) {
	// Получаем первый доступный файл для базовой проверки
	rawDir := "../../data/raw"
	files, err := filepath.Glob(filepath.Join(rawDir, "*.xlsx"))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	filePath := files[0]
	f, err := excelize.OpenFile(filePath)
	require.NoError(t, err)
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	require.NoError(t, err)

	data := ParsePositions(rows)
	require.NotNil(t, data)

	// Базовые проверки
	assert.NotEmpty(t, data.AccountNumber, "account number should not be empty")
	assert.False(t, data.PeriodEnd.IsZero(), "period end should not be zero")
	assert.False(t, data.TotalAssets.IsZero(), "total assets should not be zero")

	t.Logf("Parsed data summary:")
	t.Logf("  Account: %s", data.AccountNumber)
	t.Logf("  Period: %s - %s", data.PeriodStart.Format("2006-01-02"), data.PeriodEnd.Format("2006-01-02"))
	t.Logf("  Total Assets: %s", data.TotalAssets.String())
	t.Logf("  Cash Balance: %s", data.CashBalance.String())
	t.Logf("  Positions: %d", len(data.Positions))
	t.Logf("  Cash Flow Operations: %d", len(data.CashFlow))
	t.Logf("  Securities Flow Operations: %d", len(data.SecuritiesFlow))
}

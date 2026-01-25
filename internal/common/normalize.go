package common

import "strings"

func NormalizeContains(s, subString string) bool {
	return strings.Contains(normalize(s), normalize(subString))
}

func NormalizeRows(rows [][]string) [][]string {
	normalizedRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		normalizedRows = append(normalizedRows, normalizeRow(row))
	}
	return normalizedRows
}

func normalizeRow(row []string) []string {
	normalizedRow := make([]string, 0, len(row))
	for _, s := range row {
		normalized := normalize(s)
		if normalized == "" {
			continue
		}
		normalizedRow = append(normalizedRow, normalized)
	}
	return normalizedRow
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return strings.Join(strings.Fields(s), " ")
}

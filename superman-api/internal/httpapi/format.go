package httpapi

import "strconv"

func formatNumber(value float64, decimals int) string {
	formatted := strconv.FormatFloat(value, 'f', decimals, 64)
	decimalOffset := len(formatted)
	for index, character := range formatted {
		if character == '.' {
			decimalOffset = index
			break
		}
	}

	start := 0
	if formatted[0] == '-' {
		start = 1
	}

	for index := decimalOffset - 3; index > start; index -= 3 {
		formatted = formatted[:index] + "," + formatted[index:]
	}

	return formatted
}

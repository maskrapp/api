package utils

import "strconv"

func ConvertWithFallback(value string, fallback int) int {
	val, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return val
}

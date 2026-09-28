// Package strutils ...
package strutils

import (
	"strconv"
	"strings"
)

// ToInt ...
func ToInt(text string) int {
	number, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0
	}
	return number
}

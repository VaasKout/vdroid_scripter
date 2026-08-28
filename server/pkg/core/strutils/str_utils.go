// Package strutils ...
package strutils

// GetUniqueChars ...
func GetUniqueChars(text string) []rune {
	if text == "" {
		return []rune{}
	}

	uniqueChars := map[rune]struct{}{}
	for _, ch := range text {
		uniqueChars[ch] = struct{}{}
	}
	filteredRunes := make([]rune, 0, len(uniqueChars))

	for k := range uniqueChars {
		filteredRunes = append(filteredRunes, k)
	}
	return filteredRunes
}

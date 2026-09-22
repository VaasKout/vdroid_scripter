// Package numutils ...
package numutils

import (
	"math/rand"
	"sort"
	"time"
)

// RandInt ...
func RandInt(minNum int, maxNum int) int {
	return minNum + rand.Intn(maxNum-minNum)
}

// RandDelay ...
func RandDelay(minNum int, maxNum int) time.Duration {
	return time.Duration(minNum + rand.Intn(maxNum-minNum))
}

// GetPercentValue ...
func GetPercentValue(number int, percent int) int {
	return (number / 100) * percent
}

// MedianInt ...
func MedianInt(values []int) int {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int{}, values...)
	sort.Ints(sorted)
	return sorted[len(sorted)/2]
}

// MedianFloat ...
func MedianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)
	return sorted[len(sorted)/2]
}

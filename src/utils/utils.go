package utils

import "math"

func Linspace(start, end, n float64) []float64 {
	if n <= 0 {
		return []float64{}
	}

	numItems := math.Ceil(n)
	result := make([]float64, int(numItems))
	step := (end - start) / numItems

	for i := range result {
		result[i] = start + step*float64(i)
	}
	return result
}

func LinspaceInt(start, end float64, numItems int) []float64 {
	if numItems <= 0 {
		return []float64{}
	}

	result := make([]float64, numItems)
	step := (end - start) / float64(numItems-1)

	for i := range result {
		result[i] = start + step*float64(i)
	}
	return result
}

func AddSliceElems(s1, s2 []float64) {
	// Adds elements of s2 in-place to elements of s1
	size := min(len(s1), len(s2))
	for i := 0; i < size; i++ {
		s1[i] += s2[i]
	}
}

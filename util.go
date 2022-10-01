package main

import "math"

type Number interface {
	int | int64 | float32 | float64
}

func Min[T Number](numbers []T) T {
	var min T = T(math.Inf(1))
	for _, x := range numbers {
		if x > 0 && x < min {
			min = x
		}
	}
	return min
}

func Max[T Number](numbers []T) T {
	var max T
	for _, x := range numbers {
		if x > max {
			max = x
		}
	}
	return max
}

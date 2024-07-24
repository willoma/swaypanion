package modules

import (
	"cmp"
	"math"
)

func limit[T cmp.Ordered](value, minimum, maximum T) T {
	return max(min(value, maximum), minimum)
}

func onePercent[T ~int | ~float64 | ~uint32](maximum T) float64 {
	return float64(maximum) / 100
}

func roundStep(raw float64, stepRaw float64) float64 {
	return math.Round(float64(raw)/stepRaw) * stepRaw
}

func round[T ~int | ~float64 | ~uint32](in float64) T {
	return T(math.Round(in))
}

func toPercent[T ~int | ~float64 | ~uint32](value, maximum T) int {
	return int(math.Round(float64(value) * 100 / float64(maximum)))
}

package analysis

import "math"

const twoPi = 2 * math.Pi

func polarToCartesian(theta, radius float64) (float64, float64) {
	return radius * math.Cos(theta), radius * math.Sin(theta)
}

func angularDifference(theta1, theta2 float64) float64 {
	diff := math.Abs(theta1 - theta2)
	diff = math.Mod(diff, twoPi)
	return math.Min(diff, twoPi-diff)
}

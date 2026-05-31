package analysis

import (
	"math"

	"github.com/brogergvhs/value-dsl/internal/values"
)

type weights struct {
	Angular   float64
	Euclidean float64
	Radial    float64
}

var defaultWeights = weights{
	Angular:   0.5,
	Euclidean: 0.3,
	Radial:    0.2,
}

// ConflictScore computes conflict using default weights.
func ConflictScore(a, b values.Definition) float64 {
	return conflictScoreWithWeights(a, b, defaultWeights)
}

// conflictScoreWithWeights computes a weighted conflict score for two values:
//
//	wAngular * (delta Theta / pi) +
//	wEuclidean * (euclideanDistance / 2) +
//	wRadial * |r1 - r2|
func conflictScoreWithWeights(a, b values.Definition, w weights) float64 {
	angularConflict := angularDifference(a.Angle, b.Angle) / math.Pi
	radialDiff := math.Abs(a.Radius - b.Radius)

	ax, ay := polarToCartesian(a.Angle, a.Radius)
	bx, by := polarToCartesian(b.Angle, b.Radius)
	euclidean := math.Hypot(ax-bx, ay-by) / 2

	score := (w.Angular * angularConflict) +
		(w.Euclidean * euclidean) +
		(w.Radial * radialDiff)

	return clampUnit(score)
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

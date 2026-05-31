package analysis

import (
	"math"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/values"
)

func TestConflictScoreZeroForSameValue(t *testing.T) {
	v := values.Definition{Name: "privacy_pref", Angle: 1.2, Radius: 0.7}
	score := ConflictScore(v, v)
	if math.Abs(score) > 1e-12 {
		t.Fatalf("expected score 0, got %f", score)
	}
}

func TestConflictScoreMatchesSpecificationFormula(t *testing.T) {
	a := values.Definition{Name: "a", Angle: 0.0, Radius: 1.0}
	b := values.Definition{Name: "b", Angle: math.Pi, Radius: 1.0}

	score := ConflictScore(a, b)
	if math.Abs(score-0.8) > 1e-9 {
		t.Fatalf("expected score 0.8, got %f", score)
	}
}

func TestConflictScoreWithCustomWeights(t *testing.T) {
	a := values.Definition{Name: "a", Angle: 0.0, Radius: 1.0}
	b := values.Definition{Name: "b", Angle: math.Pi, Radius: 0.0}

	score := conflictScoreWithWeights(a, b, weights{
		Angular:   1.0,
		Euclidean: 0.0,
		Radial:    0.0,
	})
	if math.Abs(score-1.0) > 1e-9 {
		t.Fatalf("expected score 1.0, got %f", score)
	}
}

func TestConflictScoreDefaultWeightsWithinUnitInterval(t *testing.T) {
	radii := []float64{0.0, 0.1, 0.3, 0.5, 0.8, 1.0}
	angles := []float64{
		0.0,
		math.Pi / 6,
		math.Pi / 3,
		math.Pi / 2,
		math.Pi,
		1.5 * math.Pi,
		2*math.Pi - 0.001,
	}

	for _, aRadius := range radii {
		for _, bRadius := range radii {
			for _, aAngle := range angles {
				for _, bAngle := range angles {
					a := values.Definition{Name: "a", Angle: aAngle, Radius: aRadius}
					b := values.Definition{Name: "b", Angle: bAngle, Radius: bRadius}
					score := ConflictScore(a, b)
					if score < 0 || score > 1 {
						t.Fatalf("expected score in [0,1], got %f (a=%+v b=%+v)", score, a, b)
					}
				}
			}
		}
	}
}

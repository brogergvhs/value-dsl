package analysis

import (
	"math"
	"testing"
)

func TestPolarToCartesian(t *testing.T) {
	x, y := polarToCartesian(math.Pi/2, 1.0)
	if math.Abs(x-0.0) > 1e-9 {
		t.Fatalf("expected x≈0, got %f", x)
	}
	if math.Abs(y-1.0) > 1e-9 {
		t.Fatalf("expected y≈1, got %f", y)
	}
}

func TestAngularDifference(t *testing.T) {
	diff := angularDifference(0.0, 2*math.Pi-0.1)
	if math.Abs(diff-0.1) > 1e-9 {
		t.Fatalf("expected diff≈0.1, got %f", diff)
	}
}

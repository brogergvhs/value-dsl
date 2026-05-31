package values

import (
	"testing"
)

func TestBuildRegistryValid(t *testing.T) {
	if err := buildRegistry(); err != nil {
		t.Fatalf("buildRegistry failed on valid data: %v", err)
	}
}

func TestBuildRegistryRejectsInvalidData(t *testing.T) {
	orig := registry
	t.Cleanup(func() { registry = orig; _ = buildRegistry() })

	registry = append([]Definition{}, orig...)
	registry[0].Name = ""
	if err := buildRegistry(); err == nil {
		t.Fatal("expected error for empty value name")
	}
}

func TestLookupByName(t *testing.T) {
	def, ok := LookupByName("privacy")
	if !ok {
		t.Fatal("expected privacy to be present")
	}
	if def.Name != "privacy" {
		t.Fatalf("unexpected name: %q", def.Name)
	}
	if def.Category != "self_direction" {
		t.Fatalf("unexpected category: %q", def.Category)
	}
	if def.Angle == 0 {
		t.Fatalf("expected non-zero angle for privacy")
	}
	if def.Radius <= 0 {
		t.Fatalf("expected positive radius")
	}
}

func TestLookupByNameNormalization(t *testing.T) {
	def, ok := LookupByName("  PRIVACY ")
	if !ok {
		t.Fatal("expected normalized lookup to succeed")
	}
	if def.Name != "privacy" {
		t.Fatalf("unexpected name: %q", def.Name)
	}
}

func TestCategoryForAngleReturnsNearestCategory(t *testing.T) {
	for _, def := range All() {
		category, ok := categoryForAngle(def.Angle)
		if !ok {
			t.Fatalf("expected category for %q at angle %.2f", def.Name, def.Angle)
		}
		if category.Name == "" {
			t.Fatalf("expected non-empty category for angle %.2f", def.Angle)
		}
	}
}

func TestCategoryForAngleBoundaryBin(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
		want  string
	}{
		{name: "exact power boundary", angle: 0.00, want: "power"},
		{name: "inside achievement sector", angle: 0.20, want: "achievement"},
		{name: "inside achievement sector mid", angle: 0.32, want: "achievement"},
		{name: "inside achievement sector high", angle: 0.50, want: "achievement"},
		{name: "inside stimulation sector", angle: 1.57, want: "stimulation"},
		{name: "inside stimulation sector high", angle: 1.80, want: "stimulation"},
		{name: "past security wraps to power", angle: 5.90, want: "power"},
		{name: "high angle wraps to power", angle: 6.20, want: "power"},
		{name: "negative angle wraps to power", angle: -0.10, want: "power"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, ok := categoryForAngle(tt.angle)
			if !ok {
				t.Fatal("expected category")
			}
			if category.Name != tt.want {
				t.Fatalf("categoryForAngle(%v) = %q, want %q", tt.angle, category.Name, tt.want)
			}
		})
	}
}

func TestClosestCategoryRejectsOutOfBoundsAngles(t *testing.T) {
	tests := []float64{-0.01, MaxAngle, 10.0}
	for _, angle := range tests {
		if _, ok := ClosestCategory(angle); ok {
			t.Fatalf("ClosestCategory(%v) unexpectedly succeeded", angle)
		}
	}
}

func TestGeometryBoundsHelpers(t *testing.T) {
	if !IsAngleInBounds(0) || !IsAngleInBounds(MaxAngle-0.01) {
		t.Fatal("expected in-range angles to be accepted")
	}
	if IsAngleInBounds(-0.01) || IsAngleInBounds(MaxAngle) {
		t.Fatal("expected out-of-range angles to be rejected")
	}
	if !IsRadiusInBounds(0) || !IsRadiusInBounds(1) {
		t.Fatal("expected in-range radii to be accepted")
	}
	if IsRadiusInBounds(-0.01) || IsRadiusInBounds(1.01) {
		t.Fatal("expected out-of-range radii to be rejected")
	}
}

func TestAllReturnsCopy(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("expected non-empty registry")
	}

	originalFirst := all[0]
	all[0].Name = "mutated"

	refetched := All()
	if refetched[0].Name != originalFirst.Name {
		t.Fatalf("expected registry data to be immutable copy, got %q", refetched[0].Name)
	}
}

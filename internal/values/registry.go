// Package values defines the Schwartz value-circle registry and geometric helpers
//
// The registry models values on a polar coordinate system:
// - Angle is measured in radians around the Schwartz value circle in [0, 2pi]
// - Radius is measured from the center in [0, 1]
package values

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const geometryTolerance = 1e-9

const (
	MinAngle  = 0.0
	MaxAngle  = 2 * math.Pi
	MinRadius = 0.0
	MaxRadius = 1.0
)

// Definition describes a value in the Schwartz circle geometry
type Definition struct {
	Name     string
	Category string
	Angle    float64 // radians in [0, 2pi]
	Radius   float64 // distance from the center in [0, 1]
}

// Category defines a top-level Schwartz value category and its angular anchor
type Category struct {
	Name        string
	Angle       float64
	Description string
}

const twoPi = 2 * math.Pi

var categories = []Category{
	{Name: "power", Angle: 0.00, Description: "social status, control, and authority"},
	{Name: "achievement", Angle: 0.63, Description: "success, competence, and influence"},
	{Name: "hedonism", Angle: 1.26, Description: "pleasure, enjoyment, and comfort"},
	{Name: "stimulation", Angle: 1.88, Description: "novelty, excitement, and challenge"},
	{Name: "self_direction", Angle: 2.51, Description: "autonomy, creativity, and privacy"},
	{Name: "universalism", Angle: 3.14, Description: "justice, equality, and broad concern for others"},
	{Name: "benevolence", Angle: 3.77, Description: "care, loyalty, and helpfulness"},
	{Name: "tradition", Angle: 4.40, Description: "devotion, humility, and respect for customs"},
	{Name: "conformity", Angle: 5.03, Description: "obedience, discipline, and politeness"},
	{Name: "security", Angle: 5.65, Description: "safety, stability, and order"},
}

var registry = []Definition{
	{Name: "success", Category: "achievement", Angle: 0.40, Radius: 0.86},
	{Name: "competence", Category: "achievement", Angle: 0.48, Radius: 0.84},
	{Name: "influence", Category: "achievement", Angle: 0.56, Radius: 0.82},

	{Name: "pleasure", Category: "hedonism", Angle: 0.78, Radius: 0.80},
	{Name: "enjoyment", Category: "hedonism", Angle: 0.86, Radius: 0.78},
	{Name: "comfort", Category: "hedonism", Angle: 0.94, Radius: 0.76},

	{Name: "excitement", Category: "stimulation", Angle: 1.33, Radius: 0.86},
	{Name: "novelty", Category: "stimulation", Angle: 1.49, Radius: 0.84},
	{Name: "challenge", Category: "stimulation", Angle: 1.73, Radius: 0.82},

	{Name: "privacy", Category: "self_direction", Angle: 2.04, Radius: 0.92},
	{Name: "autonomy", Category: "self_direction", Angle: 2.20, Radius: 0.90},
	{Name: "creativity", Category: "self_direction", Angle: 2.36, Radius: 0.88},

	{Name: "equality", Category: "universalism", Angle: 2.67, Radius: 0.88},
	{Name: "justice", Category: "universalism", Angle: 2.83, Radius: 0.90},
	{Name: "environmentalism", Category: "universalism", Angle: 2.99, Radius: 0.86},

	{Name: "helpfulness", Category: "benevolence", Angle: 3.30, Radius: 0.82},
	{Name: "care", Category: "benevolence", Angle: 3.46, Radius: 0.84},
	{Name: "loyalty", Category: "benevolence", Angle: 3.62, Radius: 0.80},

	{Name: "humility", Category: "tradition", Angle: 3.93, Radius: 0.74},
	{Name: "devotion", Category: "tradition", Angle: 4.09, Radius: 0.72},
	{Name: "respect_for_customs", Category: "tradition", Angle: 4.25, Radius: 0.70},

	{Name: "obedience", Category: "conformity", Angle: 4.56, Radius: 0.78},
	{Name: "self_discipline", Category: "conformity", Angle: 4.72, Radius: 0.80},
	{Name: "politeness", Category: "conformity", Angle: 4.88, Radius: 0.76},

	{Name: "safety", Category: "security", Angle: 5.19, Radius: 0.94},
	{Name: "stability", Category: "security", Angle: 5.35, Radius: 0.92},
	{Name: "order", Category: "security", Angle: 5.51, Radius: 0.90},

	{Name: "authority", Category: "power", Angle: 5.78, Radius: 0.90},
	{Name: "social_status", Category: "power", Angle: 5.94, Radius: 0.85},
	{Name: "wealth", Category: "power", Angle: 6.10, Radius: 0.88},
}

var byName map[string]Definition
var byCategory map[string][]Definition
var categoryByName map[string]Category
var registryErr error

func init() {
	registryErr = buildRegistry()
}

func RegistryError() error {
	return registryErr
}

func buildRegistry() error {
	byName = make(map[string]Definition, len(registry))
	byCategory = make(map[string][]Definition)
	categoryByName = make(map[string]Category, len(categories))

	for i, category := range categories {
		category.Name = strings.ToLower(strings.TrimSpace(category.Name))
		if category.Name == "" {
			return fmt.Errorf("invalid empty category name")
		}
		category.Angle = math.Mod(category.Angle, twoPi)
		if category.Angle < 0 {
			category.Angle += twoPi
		}
		if category.Angle < 0 || category.Angle >= twoPi {
			return fmt.Errorf("invalid angle for category %q", category.Name)
		}
		if _, exists := categoryByName[category.Name]; exists {
			return fmt.Errorf("duplicate category %q", category.Name)
		}
		categories[i] = category
		categoryByName[category.Name] = category
	}

	for i, def := range registry {
		def.Name = strings.ToLower(strings.TrimSpace(def.Name))
		def.Category = strings.ToLower(strings.TrimSpace(def.Category))
		if def.Name == "" {
			return fmt.Errorf("invalid empty value name")
		}
		if def.Category == "" {
			return fmt.Errorf("invalid empty category for value %q", def.Name)
		}
		if _, ok := categoryByName[def.Category]; !ok {
			return fmt.Errorf("invalid category for value %q", def.Name)
		}
		def.Angle = math.Mod(def.Angle, twoPi)
		if def.Angle < 0 {
			def.Angle += twoPi
		}
		if def.Angle < 0 || def.Angle >= twoPi {
			return fmt.Errorf("invalid angle for value %q", def.Name)
		}
		if def.Radius < 0 || def.Radius > 1 {
			return fmt.Errorf("invalid radius for value %q", def.Name)
		}
		if _, exists := byName[def.Name]; exists {
			return fmt.Errorf("duplicate value %q", def.Name)
		}
		registry[i] = def
		byName[def.Name] = def
		byCategory[def.Category] = append(byCategory[def.Category], def)
	}

	for key := range byCategory {
		vals := byCategory[key]
		sort.Slice(vals, func(i, j int) bool { return vals[i].Name < vals[j].Name })
		byCategory[key] = vals
	}
	return nil
}

func Categories() []Category {
	out := make([]Category, len(categories))
	copy(out, categories)
	return out
}

// categoryForAngle finds the category whose sector contains the angle.
// Each category's angle is the upper bound of its sector.
// Sector spans from the previous category's angle to this one.
func categoryForAngle(angle float64) (Category, bool) {
	if len(categories) == 0 {
		return Category{}, false
	}

	normalized := math.Mod(angle, twoPi)
	if normalized < 0 {
		normalized += twoPi
	}

	for _, category := range categories {
		if normalized <= category.Angle+geometryTolerance {
			return category, true
		}
	}

	first := categories[0]
	last := categories[len(categories)-1]
	if normalized > last.Angle && normalized < twoPi {
		return first, true
	}
	if math.Abs(normalized-twoPi) <= geometryTolerance || normalized <= first.Angle+geometryTolerance {
		return first, true
	}
	return Category{}, false
}

func ClosestCategory(angle float64) (Category, bool) {
	if !IsAngleInBounds(angle) {
		return Category{}, false
	}
	return categoryForAngle(angle)
}

func IsAngleInBounds(angle float64) bool {
	return angle >= MinAngle && angle < MaxAngle
}

func IsRadiusInBounds(radius float64) bool {
	return radius >= MinRadius && radius <= MaxRadius
}

func All() []Definition {
	out := make([]Definition, len(registry))
	copy(out, registry)
	return out
}

func LookupByName(name string) (Definition, bool) {
	def, ok := byName[strings.ToLower(strings.TrimSpace(name))]
	return def, ok
}

func LookupByGeometry(angle, radius float64) []Definition {
	matches := make([]Definition, 0)
	for _, def := range registry {
		if math.Abs(def.Angle-angle) <= geometryTolerance && math.Abs(def.Radius-radius) <= geometryTolerance {
			matches = append(matches, def)
		}
	}
	return matches
}

package analysis

import (
	"sort"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/values"
)

type conflictInputs struct {
	stakeholders map[string]struct{}
	requirements map[string]struct{}
	valueDefs    map[string]values.Definition
}

type valuePair struct {
	a string
	b string
}

func AnalyzeConflicts(m *semantic.SemanticModel) []model.ConflictResult {
	if m == nil || len(m.AssignmentEntries) == 0 {
		return nil
	}

	inputs := buildConflictInputs(m)
	grouped := groupAssignmentsForConflictAnalysis(m.AssignmentEntries, inputs)
	scoreCache := make(map[valuePair]float64)
	results := make([]model.ConflictResult, 0)
	for _, reqID := range sortedKeys(grouped) {
		results = append(results, scoreRequirementConflicts(reqID, grouped[reqID], inputs.valueDefs, scoreCache)...)
	}
	return results
}

func buildConflictInputs(m *semantic.SemanticModel) conflictInputs {
	stks := make(map[string]struct{}, len(m.Stakeholders))
	for _, s := range m.Stakeholders {
		if n := strings.TrimSpace(s.Name); n != "" && !s.Invalid {
			stks[n] = struct{}{}
		}
	}
	reqs := make(map[string]struct{}, len(m.Requirements))
	for _, r := range m.Requirements {
		if n := strings.TrimSpace(r.ID); n != "" && !r.Invalid {
			reqs[n] = struct{}{}
		}
	}
	return conflictInputs{stakeholders: stks, requirements: reqs, valueDefs: collectValueDefinitions(m.Values)}
}

// Value geometry is resolved once here, then reused across grouping and scoring.
func collectValueDefinitions(valuesList []model.Value) map[string]values.Definition {
	known := make(map[string]values.Definition, len(valuesList)+len(values.All()))
	for _, definition := range values.All() {
		if name := strings.TrimSpace(definition.Name); name != "" {
			known[name] = definition
		}
	}
	for _, value := range valuesList {
		name := strings.TrimSpace(value.Name)
		if name == "" || value.Invalid {
			continue
		}
		if _, exists := known[name]; exists && value.Builtin {
			continue
		}
		known[name] = values.Definition{Name: value.Name, Category: value.Category, Angle: value.Angle, Radius: value.Radius}
	}
	return known
}

func groupAssignmentsForConflictAnalysis(assignments []model.AssignmentEntry, inputs conflictInputs) map[string]map[string]map[string]struct{} {
	grouped := make(map[string]map[string]map[string]struct{}, len(assignments))
	for _, entry := range assignments {
		reqID := strings.TrimSpace(entry.RequirementID)
		if _, ok := inputs.requirements[reqID]; !ok {
			continue
		}
		stakeholders := validRefNames(entry.Stakeholders, inputs.stakeholders)
		values := validRefNames(entry.Values, inputs.valueDefs)
		if len(stakeholders) == 0 || len(values) == 0 {
			continue
		}
		if grouped[reqID] == nil {
			grouped[reqID] = make(map[string]map[string]struct{})
		}
		for _, stakeholder := range stakeholders {
			if grouped[reqID][stakeholder] == nil {
				grouped[reqID][stakeholder] = make(map[string]struct{})
			}
			for _, value := range values {
				grouped[reqID][stakeholder][value] = struct{}{}
			}
		}
	}
	return grouped
}

func scoreRequirementConflicts(reqID string, assignments map[string]map[string]struct{}, defs map[string]values.Definition, scoreCache map[valuePair]float64) []model.ConflictResult {
	stakeholders := sortedKeys(assignments)
	valueNames := make(map[string][]string, len(assignments))
	for stakeholder, set := range assignments {
		valueNames[stakeholder] = sortedKeys(set)
	}
	var cap int
	for i, s := range stakeholders {
		for j := i + 1; j < len(stakeholders); j++ {
			cap += len(valueNames[s]) * len(valueNames[stakeholders[j]])
		}
	}
	results := make([]model.ConflictResult, 0, cap)

	for i := range stakeholders {
		left := stakeholders[i]
		for j := i + 1; j < len(stakeholders); j++ {
			right := stakeholders[j]
			for _, leftValue := range valueNames[left] {
				for _, rightValue := range valueNames[right] {
					results = append(results, model.ConflictResult{
						RequirementID: reqID,
						StakeholderA:  left,
						ValueA:        leftValue,
						StakeholderB:  right,
						ValueB:        rightValue,
						ConflictScore: conflictScoreCached(leftValue, rightValue, defs, scoreCache),
					})
				}
			}
		}
	}
	return results
}

func conflictScoreCached(a, b string, defs map[string]values.Definition, cache map[valuePair]float64) float64 {
	if a > b {
		a, b = b, a
	}
	if score, ok := cache[valuePair{a, b}]; ok {
		return score
	}
	score := ConflictScore(defs[a], defs[b])
	cache[valuePair{a, b}] = score
	return score
}

func validRefNames[T any](refs []model.Ref, known map[string]T) []string {
	seen := make(map[string]struct{}, len(refs))
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		name := strings.TrimSpace(ref.Name)
		if _, ok := known[name]; name != "" && ok {
			if _, dup := seen[name]; !dup {
				seen[name] = struct{}{}
				out = append(out, name)
			}
		}
	}
	sort.Strings(out)
	return out
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

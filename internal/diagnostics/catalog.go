// Package diagnostics centralizes diagnostic metadata shared across parser,
// validation, and LSP conversion.
package diagnostics

import "strings"

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Source string

const (
	SourceGrammar  Source = "grammar"
	SourceSemantic Source = "semantic"
)

type Spec struct {
	Code     string
	Severity Severity
	Message  string
	Source   Source
}

var wordReplacer = strings.NewReplacer(".", " ", "_", " ")

var semanticCodes = map[string]struct{}{
	"validation.model.nil":                    {},
	"stakeholder.duplicate":                   {},
	"stakeholder.unused":                      {},
	"value.duplicate":                         {},
	"value.angle.out_of_bounds":               {},
	"value.radius.out_of_bounds":              {},
	"value.builtin.conflict":                  {},
	"value.builtin.match":                     {},
	"reference.invalid_declaration":           {},
	"requirement.duplicate":                   {},
	"requirement.context.unknown_stakeholder": {},
	"requirement.stakeholder.unknown":         {},
	"requirement.action.target.unknown":       {},
	"assignment.requirement.unknown":          {},
	"assignment.stakeholder.unknown":          {},
	"assignment.value.unknown":                {},
	"assignment.duplicate":                    {},
}

var warningCodes = map[string]struct{}{
	"value.builtin.match":           {},
	"stakeholder.unused":            {},
	"reference.invalid_declaration": {},
}

var customMessages = map[string]string{
	"parse.line.unclassified":       "line could not be classified in current block context",
	"reference.invalid_declaration": "reference resolves only to invalid declarations",
}

func SpecFor(code string) Spec {
	spec := Spec{
		Code:     code,
		Severity: SeverityError,
		Message:  wordReplacer.Replace(code),
		Source:   fallbackSource(code),
	}
	if _, ok := semanticCodes[code]; ok {
		spec.Source = SourceSemantic
	}
	if _, ok := warningCodes[code]; ok {
		spec.Severity = SeverityWarning
	}
	if message, ok := customMessages[code]; ok {
		spec.Message = message
	}
	return spec
}

func fallbackSource(code string) Source {
	switch {
	case strings.HasPrefix(code, "parse."),
		strings.HasPrefix(code, "validation."),
		strings.HasPrefix(code, "stakeholder."),
		strings.HasPrefix(code, "value."),
		strings.HasPrefix(code, "requirement."),
		strings.HasPrefix(code, "assignment."):
		return SourceGrammar
	default:
		return SourceSemantic
	}
}

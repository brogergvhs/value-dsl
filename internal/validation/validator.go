// Package validation checks semantic model integrity and produces diagnostics.
package validation

import (
	"fmt"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/diagnostics"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/values"
)

// Public types

type Severity = diagnostics.Severity
type DiagnosticSource = diagnostics.Source

const (
	SeverityError   = diagnostics.SeverityError
	SeverityWarning = diagnostics.SeverityWarning
)

type Diagnostic struct {
	Line        int
	Column      int
	Severity    Severity
	Code        DiagnosticCode
	Message     string
	Source      DiagnosticSource
	SubjectKind grammar.DeclarationKind
	SubjectName string
}

type DiagnosticCode string

const (
	CodeValidationModelNil                   DiagnosticCode = "validation.model.nil"
	CodeStakeholderDuplicate                 DiagnosticCode = "stakeholder.duplicate"
	CodeStakeholderUnused                    DiagnosticCode = "stakeholder.unused"
	CodeValueDuplicate                       DiagnosticCode = "value.duplicate"
	CodeValueAngleOutOfBounds                DiagnosticCode = "value.angle.out_of_bounds"
	CodeValueRadiusOutOfBounds               DiagnosticCode = "value.radius.out_of_bounds"
	CodeValueBuiltinConflict                 DiagnosticCode = "value.builtin.conflict"
	CodeValueBuiltinMatch                    DiagnosticCode = "value.builtin.match"
	CodeRequirementDuplicate                 DiagnosticCode = "requirement.duplicate"
	CodeReferenceInvalidDeclaration          DiagnosticCode = "reference.invalid_declaration"
	CodeRequirementContextUnknownStakeholder DiagnosticCode = "requirement.context.unknown_stakeholder"
	CodeRequirementStakeholderUnknown        DiagnosticCode = "requirement.stakeholder.unknown"
	CodeRequirementActionTargetUnknown       DiagnosticCode = "requirement.action.target.unknown"
	CodeAssignmentRequirementUnknown         DiagnosticCode = "assignment.requirement.unknown"
	CodeAssignmentStakeholderUnknown         DiagnosticCode = "assignment.stakeholder.unknown"
	CodeAssignmentValueUnknown               DiagnosticCode = "assignment.value.unknown"
	CodeAssignmentDuplicate                  DiagnosticCode = "assignment.duplicate"
)

// Entry point

func Validate(m *semantic.SemanticModel) []Diagnostic {
	if m == nil {
		return []Diagnostic{NewDiagnostic(CodeValidationModelNil, 0, 1)}
	}
	ctx := newCtx(m)
	diags := validateAll(ctx)
	applyDeclarationValidity(ctx, diags)
	return append(diags, append(unusedStakeholderWarnings(ctx), invalidReferenceWarnings(ctx)...)...)
}

// Validators

func validateAll(ctx *vctx) []Diagnostic {
	diags := checkStakeholders(ctx)
	diags = append(diags, checkValues(ctx)...)
	diags = append(diags, checkRequirements(ctx)...)
	diags = append(diags, checkAssignments(ctx)...)
	return diags
}

func checkStakeholders(ctx *vctx) []Diagnostic {
	var diags []Diagnostic
	seen := make(map[string]struct{})

	for _, s := range ctx.m.Stakeholders {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			diags = append(diags, declDiagAt(s.Line, s.Range.Start.Column, CodeStakeholderDuplicate, grammar.DeclarationKindStakeholder, name))
			continue
		}
		seen[name] = struct{}{}
	}

	return diags
}

func checkValues(ctx *vctx) []Diagnostic {
	var diags []Diagnostic
	seen := make(map[string]struct{})

	for _, v := range ctx.m.Values {
		name := strings.TrimSpace(v.Name)
		if name == "" {
			continue
		}
		if _, dup := seen[name]; dup {
			diags = append(diags, declDiagAt(v.Line, v.Range.Start.Column, CodeValueDuplicate, grammar.DeclarationKindValue, name))
		} else {
			seen[name] = struct{}{}
		}
		if v.HasAngle && !values.IsAngleInBounds(v.Angle) {
			diags = append(diags, declDiagAt(v.Line, v.Range.Start.Column, CodeValueAngleOutOfBounds, grammar.DeclarationKindValue, name))
		}
		if v.HasRadius && !values.IsRadiusInBounds(v.Radius) {
			diags = append(diags, declDiagAt(v.Line, v.Range.Start.Column, CodeValueRadiusOutOfBounds, grammar.DeclarationKindValue, name))
		}
		if _, builtin := ctx.builtins[name]; builtin {
			diags = append(diags, declDiagAt(v.Line, v.Range.Start.Column, CodeValueBuiltinConflict, grammar.DeclarationKindValue, name))
			continue
		}
		if v.HasAngle && v.HasRadius && values.IsAngleInBounds(v.Angle) && values.IsRadiusInBounds(v.Radius) {
			if len(values.LookupByGeometry(v.Angle, v.Radius)) > 0 {
				diags = append(diags, declDiagAt(v.Line, v.Range.Start.Column, CodeValueBuiltinMatch, grammar.DeclarationKindValue, name))
			}
		}
	}

	return diags
}

func checkRequirements(ctx *vctx) []Diagnostic {
	var diags []Diagnostic
	seen := make(map[string]struct{})

	for _, req := range ctx.m.Requirements {
		id := strings.TrimSpace(req.ID)
		if id != "" {
			if _, dup := seen[id]; dup {
				diags = append(diags, declDiagAt(req.Line, req.Range.Start.Column, CodeRequirementDuplicate, grammar.DeclarationKindRequirement, id))
			} else {
				seen[id] = struct{}{}
			}
		}

		for _, clause := range req.Context.Clauses {
			if actor := strings.TrimSpace(clause.Actor); actor != "" && !ctx.hasStakeholder(actor) {
				diags = append(diags, declDiagAt(
					firstNonZero(clause.Line, req.Line),
					firstNonZero(clause.Range.Start.Column, req.Range.Start.Column),
					CodeRequirementContextUnknownStakeholder, grammar.DeclarationKindRequirement, id))
			}
		}
		for _, s := range req.Stakeholders {
			if name := strings.TrimSpace(s.Name); name != "" && !ctx.hasStakeholder(name) {
				diags = append(diags, declDiagAt(
					firstNonZero(req.StakeholdersLine, req.Line),
					firstNonZero(req.StakeholdersRange.Start.Column, req.Range.Start.Column),
					CodeRequirementStakeholderUnknown, grammar.DeclarationKindRequirement, id))
			}
		}
		if target := strings.TrimSpace(req.Action.Target); target != "" && !ctx.hasStakeholder(target) {
			diags = append(diags, declDiagAt(
				firstNonZero(req.Action.Line, req.Line),
				firstNonZero(req.Action.Range.Start.Column, req.Range.Start.Column),
				CodeRequirementActionTargetUnknown, grammar.DeclarationKindRequirement, id))
		}
	}

	return diags
}

func checkAssignments(ctx *vctx) []Diagnostic {
	var diags []Diagnostic
	seenHeaders := make(map[int]struct{})
	seenPairs := make(map[string]struct{})

	for _, entry := range ctx.m.AssignmentEntries {
		line := firstNonZero(entry.Line, entry.HeaderLine)
		reqID := strings.TrimSpace(entry.RequirementID)
		if reqID != "" && !ctx.hasRequirement(reqID) {
			header := firstNonZero(entry.HeaderLine, line)
			if _, ok := seenHeaders[header]; !ok {
				diags = append(diags, NewDiagnostic(CodeAssignmentRequirementUnknown, header, 1))
				seenHeaders[header] = struct{}{}
			}
		}

		stks := uniqueRefs(entry.Stakeholders)
		vals := uniqueRefs(entry.Values)
		for _, s := range stks {
			if !ctx.hasStakeholder(s) {
				diags = append(diags, NewDiagnostic(CodeAssignmentStakeholderUnknown, line, 1))
			}
		}
		for _, v := range vals {
			if !ctx.hasValue(v) {
				diags = append(diags, NewDiagnostic(CodeAssignmentValueUnknown, line, 1))
			}
		}
		if reqID == "" || len(stks) == 0 || len(vals) == 0 {
			continue
		}

		for _, s := range stks {
			for _, v := range vals {
				key := reqID + "|" + s + "|" + v
				if _, dup := seenPairs[key]; dup {
					diags = append(diags, NewDiagnostic(CodeAssignmentDuplicate, line, 1))
					continue
				}
				seenPairs[key] = struct{}{}
			}
		}
	}

	return diags
}

// Declaration validity propagation

type declarationRef struct {
	Kind grammar.DeclarationKind
	Name string
}

type nameAtLine struct {
	Name string
	Line int
}

type invalidTarget struct {
	Kind grammar.DeclarationKind
	Name string
	Line int
}

func applyDeclarationValidity(ctx *vctx, diags []Diagnostic) {
	targets := collectInvalidTargets(ctx, diags)

	for i := range ctx.m.Stakeholders {
		markStakeholderInvalid(&ctx.m.Stakeholders[i], targets)
	}
	for name, s := range ctx.m.StakeholderByName {
		markStakeholderInvalid(&s, targets)
		ctx.m.StakeholderByName[name] = s
	}
	for i := range ctx.m.Values {
		markValueInvalid(&ctx.m.Values[i], targets)
	}
	for name, v := range ctx.m.ValueByName {
		markValueInvalid(&v, targets)
		ctx.m.ValueByName[name] = v
	}
	for i := range ctx.m.Requirements {
		markRequirementInvalid(&ctx.m.Requirements[i], targets)
	}
}

func markStakeholderInvalid(s *model.Stakeholder, targets invalidTargets) {
	s.Invalid = s.Invalid || targets.matches(grammar.DeclarationKindStakeholder, s.Line, s.Name)
}

func markValueInvalid(v *model.Value, targets invalidTargets) {
	v.Invalid = v.Invalid || targets.matches(grammar.DeclarationKindValue, v.Line, v.Name)
}

func markRequirementInvalid(r *model.Requirement, targets invalidTargets) {
	r.Invalid = r.Invalid || targets.matches(grammar.DeclarationKindRequirement, r.Line, r.ID)
}

type invalidTargets struct{ m map[invalidTarget]struct{} }

func (t invalidTargets) matches(kind grammar.DeclarationKind, line int, name string) bool {
	_, ok := t.m[invalidTarget{kind, strings.TrimSpace(name), line}]
	return ok
}

func collectInvalidTargets(ctx *vctx, diags []Diagnostic) invalidTargets {
	t := invalidTargets{m: make(map[invalidTarget]struct{})}
	for _, d := range diags {
		if d.Severity != SeverityError || d.SubjectKind == "" || (d.SubjectKind != grammar.DeclarationKindStakeholder && d.SubjectKind != grammar.DeclarationKindValue && d.SubjectKind != grammar.DeclarationKindRequirement) {
			continue
		}
		line := resolveDeclarationLine(ctx, d.SubjectKind, d.Line, d.SubjectName)
		t.m[invalidTarget{d.SubjectKind, strings.TrimSpace(d.SubjectName), line}] = struct{}{}
	}

	return t
}

func resolveDeclarationLine(ctx *vctx, kind grammar.DeclarationKind, line int, name string) int {
	if name = strings.TrimSpace(name); name == "" {
		return line
	}
	if _, ok := ctx.declExact[nameAtLine{name, line}]; ok {
		return line
	}
	if resolved, ok := ctx.declLine[declarationRef{kind, name}]; ok {
		return resolved
	}

	return line
}

// Context

type vctx struct {
	m         *semantic.SemanticModel
	declLine  map[declarationRef]int // first-seen line (1-indexed); 0 = absent
	builtins  map[string]struct{}
	declExact map[nameAtLine]struct{}
}

func newCtx(m *semantic.SemanticModel) *vctx {
	ctx := &vctx{
		m:         m,
		declLine:  make(map[declarationRef]int),
		builtins:  make(map[string]struct{}),
		declExact: make(map[nameAtLine]struct{}),
	}

	for _, b := range values.All() {
		if name := strings.TrimSpace(b.Name); name != "" {
			ctx.builtins[name] = struct{}{}
		}
	}
	for _, s := range m.Stakeholders {
		ctx.index(grammar.DeclarationKindStakeholder, s.Line, s.Name)
	}
	for _, v := range m.Values {
		ctx.index(grammar.DeclarationKindValue, v.Line, v.Name)
	}
	for _, r := range m.Requirements {
		ctx.index(grammar.DeclarationKindRequirement, r.Line, r.ID)
	}

	return ctx
}

func (ctx *vctx) index(kind grammar.DeclarationKind, line int, name string) {
	if name = strings.TrimSpace(name); name == "" {
		return
	}

	ref := declarationRef{kind, name}
	if _, exists := ctx.declLine[ref]; !exists {
		ctx.declLine[ref] = line
	}

	ctx.declExact[nameAtLine{name, line}] = struct{}{}
}

func (ctx *vctx) hasStakeholder(name string) bool {
	return ctx.declLine[declarationRef{grammar.DeclarationKindStakeholder, strings.TrimSpace(name)}] > 0
}

func (ctx *vctx) hasRequirement(name string) bool {
	return ctx.declLine[declarationRef{grammar.DeclarationKindRequirement, strings.TrimSpace(name)}] > 0
}

func (ctx *vctx) hasValue(name string) bool {
	name = strings.TrimSpace(name)
	_, isBuiltin := ctx.builtins[name]
	return isBuiltin || ctx.declLine[declarationRef{grammar.DeclarationKindValue, name}] > 0
}

func (ctx *vctx) hasBuiltinValue(name string) bool {
	_, ok := ctx.builtins[strings.TrimSpace(name)]
	return ok
}

// Diagnostic constructors

func NormalizeDiagnostic(d Diagnostic) Diagnostic {
	spec := diagnostics.SpecFor(string(d.Code))
	if d.Line <= 0 {
		d.Line = 1
	}
	if d.Column <= 0 {
		d.Column = 1
	}
	if d.Severity == "" {
		d.Severity = Severity(spec.Severity)
	}
	if strings.TrimSpace(d.Message) == "" {
		d.Message = spec.Message
	}
	if d.Source == "" {
		d.Source = DiagnosticSource(spec.Source)
	}

	return d
}

func NewDiagnostic(code DiagnosticCode, line, column int) Diagnostic {
	return NormalizeDiagnostic(Diagnostic{Line: line, Column: column, Code: code})
}

func FormatDiagnostic(d Diagnostic) string {
	d = NormalizeDiagnostic(d)
	return fmt.Sprintf("%d:%d - %s - %s - [%s]", d.Line, d.Column, d.Severity, d.Message, d.Code)
}

func NewDeclarationDiagnostic(code DiagnosticCode, line, column int, kind grammar.DeclarationKind, name string) Diagnostic {
	d := NewDiagnostic(code, line, column)
	d.SubjectKind = kind
	d.SubjectName = strings.TrimSpace(name)
	return d
}

func declDiagAt(line, column int, code DiagnosticCode, kind grammar.DeclarationKind, name string) Diagnostic {
	return NewDeclarationDiagnostic(code, line, column, kind, name)
}

// Helpers

func firstNonZero(vals ...int) int {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}

func uniqueRefs(refs []model.Ref) []string {
	seen := make(map[string]struct{}, len(refs))
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		if name := strings.TrimSpace(r.Name); name != "" {
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				out = append(out, name)
			}
		}
	}

	return out
}

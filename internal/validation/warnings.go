package validation

import (
	"strings"

	"github.com/brogergvhs/value-dsl/internal/grammar"
)

func unusedStakeholderWarnings(ctx *vctx) []Diagnostic {
	used := make(map[string]struct{})

	for _, req := range ctx.m.Requirements {
		for _, clause := range req.Context.Clauses {
			if name := strings.TrimSpace(clause.Actor); name != "" {
				used[name] = struct{}{}
			}
		}
		if name := strings.TrimSpace(req.Action.Target); name != "" {
			used[name] = struct{}{}
		}
		for _, s := range req.Stakeholders {
			if name := strings.TrimSpace(s.Name); name != "" {
				used[name] = struct{}{}
			}
		}
	}

	for _, entry := range ctx.m.AssignmentEntries {
		for _, s := range entry.Stakeholders {
			if name := strings.TrimSpace(s.Name); name != "" {
				used[name] = struct{}{}
			}
		}
	}

	var diags []Diagnostic
	for _, s := range ctx.m.Stakeholders {
		name := strings.TrimSpace(s.Name)
		if name == "" || s.Invalid {
			continue
		}
		if _, ok := used[name]; !ok {
			diags = append(diags, declDiagAt(s.Line, s.Range.Start.Column, CodeStakeholderUnused, grammar.DeclarationKindStakeholder, name))
		}
	}

	return diags
}

func invalidReferenceWarnings(ctx *vctx) []Diagnostic {
	invalid := invalidDeclarationNames(ctx)
	var diags []Diagnostic

	for _, req := range ctx.m.Requirements {
		for _, clause := range req.Context.Clauses {
			if invalid.has(grammar.DeclarationKindStakeholder, clause.Actor) {
				diags = append(diags, NewDeclarationDiagnostic(CodeReferenceInvalidDeclaration,
					firstNonZero(clause.Line, req.Line),
					firstNonZero(clause.Range.Start.Column, req.Range.Start.Column),
					grammar.DeclarationKindStakeholder, clause.Actor))
			}
		}
		if invalid.has(grammar.DeclarationKindStakeholder, req.Action.Target) {
			diags = append(diags, NewDeclarationDiagnostic(CodeReferenceInvalidDeclaration,
				firstNonZero(req.Action.Line, req.Line),
				firstNonZero(req.Action.Range.Start.Column, req.Range.Start.Column),
				grammar.DeclarationKindStakeholder, req.Action.Target))
		}
		for _, s := range req.Stakeholders {
			if invalid.has(grammar.DeclarationKindStakeholder, s.Name) {
				diags = append(diags, NewDeclarationDiagnostic(CodeReferenceInvalidDeclaration,
					firstNonZero(req.StakeholdersLine, req.Line),
					firstNonZero(req.StakeholdersRange.Start.Column, req.Range.Start.Column),
					grammar.DeclarationKindStakeholder, s.Name))
			}
		}
	}

	for _, entry := range ctx.m.AssignmentEntries {
		line := firstNonZero(entry.Line, entry.HeaderLine)
		if invalid.has(grammar.DeclarationKindRequirement, entry.RequirementID) {
			diags = append(diags, NewDeclarationDiagnostic(CodeReferenceInvalidDeclaration,
				firstNonZero(entry.HeaderLine, line), 1,
				grammar.DeclarationKindRequirement, entry.RequirementID))
		}
		for _, s := range entry.Stakeholders {
			if invalid.has(grammar.DeclarationKindStakeholder, s.Name) {
				diags = append(diags, NewDeclarationDiagnostic(CodeReferenceInvalidDeclaration,
					line, 1, grammar.DeclarationKindStakeholder, s.Name))
			}
		}
		for _, v := range entry.Values {
			if invalid.has(grammar.DeclarationKindValue, v.Name) {
				diags = append(diags, NewDeclarationDiagnostic(CodeReferenceInvalidDeclaration,
					line, 1, grammar.DeclarationKindValue, v.Name))
			}
		}
	}

	return diags
}

type invalidNames map[declarationRef]struct{}

func invalidDeclarationNames(ctx *vctx) invalidNames {
	inv := make(invalidNames)

	for _, s := range ctx.m.Stakeholders {
		if s.Invalid {
			inv[declarationRef{grammar.DeclarationKindStakeholder, strings.TrimSpace(s.Name)}] = struct{}{}
		}
	}
	for _, v := range ctx.m.Values {
		if v.Invalid && !v.Builtin && !ctx.hasBuiltinValue(v.Name) {
			inv[declarationRef{grammar.DeclarationKindValue, strings.TrimSpace(v.Name)}] = struct{}{}
		}
	}
	for _, r := range ctx.m.Requirements {
		if r.Invalid {
			inv[declarationRef{grammar.DeclarationKindRequirement, strings.TrimSpace(r.ID)}] = struct{}{}
		}
	}

	return inv
}

func (n invalidNames) has(kind grammar.DeclarationKind, name string) bool {
	_, ok := n[declarationRef{kind, strings.TrimSpace(name)}]
	return ok
}

package parser

import (
	"testing"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
)

func TestParsePreservesPartialRequirementLines(t *testing.T) {
	result := Parse(`stakeholder Worker

requirement R1
when Worker enters
system shall notify
stakeholders
priority
linked_to
`)

	reqs := requirements(result.AST)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %+v", reqs)
	}

	req := reqs[0]
	if req.Header.First("id") != "R1" {
		t.Fatalf("expected requirement id R1, got %+v", req)
	}
	if len(requirementEntries(req, grammar.RequirementClauseWhen)) != 1 {
		t.Fatalf("expected 1 preserved clause, got %+v", req.Body)
	}
	whenClause := requirementEntries(req, grammar.RequirementClauseWhen)[0]
	if whenClause.First(ast.FieldRawText) != "Worker enters" {
		t.Fatalf("expected partial when clause, got %+v", whenClause)
	}
	action := firstRequirementEntry(t, req, grammar.RequirementClauseSystemShall)
	if action.First("verb") != "notify" || action.First("target") != "" {
		t.Fatalf("expected partial action line, got %+v", action)
	}
	sh := firstRequirementEntry(t, req, grammar.RequirementClauseStakeholders)
	if sh.Line == 0 || len(sh.Texts("actors")) != 0 {
		t.Fatalf("expected empty stakeholders line, got %+v", sh)
	}
	priority := firstRequirementEntry(t, req, grammar.RequirementClausePriority)
	if priority.Line == 0 || priority.First("level") != "" {
		t.Fatalf("expected empty priority line, got %+v", priority)
	}
	traceability := requirementEntries(req, grammar.RequirementClauseLinkedTo)
	if len(traceability) != 1 || traceabilityValue(traceability[0]) != "" {
		t.Fatalf("expected empty linked_to line, got %+v", traceability)
	}
}

func traceabilityValue(node ast.GenericNode) string {
	for _, name := range []string{"target", "value", "level"} {
		if value := node.First(name); value != "" {
			return value
		}
	}
	return ""
}

func TestParseReportsKnownLineShapeIssuesFromGrammar(t *testing.T) {
	result := Parse(`requirement R1
where
system shall notify
stakeholders
assignment R1
Worker ->
`)

	for _, code := range []string{
		"requirement.where.rest.missing",
		"requirement.system_shall.target.missing",
		"requirement.stakeholders.actors.missing",
		"assignment.assignment_entry.values.missing",
	} {
		if !hasParseDiagnostic(result.Diagnostics, code) {
			t.Fatalf("expected grammar diagnostic %q, got %+v", code, result.Diagnostics)
		}
	}
}

func TestParsePreservesUnknownLinesAsMalformed(t *testing.T) {
	result := Parse(`mystery line
requirement R1
unknown clause
`)

	if len(result.AST.Malformed) != 1 {
		t.Fatalf("expected malformed top-level line, got %+v", result.AST.Malformed)
	}
	reqs := requirements(result.AST)
	if len(reqs) != 1 || len(reqs[0].Malformed) != 1 {
		t.Fatalf("expected malformed requirement body line, got %+v", reqs)
	}
	for _, code := range []string{
		"parse.line.unclassified",
		"requirement.system_shall.missing",
		"requirement.stakeholders.missing",
	} {
		if !hasParseDiagnostic(result.Diagnostics, code) {
			t.Fatalf("expected parse diagnostic %q, got %+v", code, result.Diagnostics)
		}
	}
}

func TestParseReportsKnownDeclarationShapeIssuesFromGrammar(t *testing.T) {
	result := Parse(`stakeholder
value
requirement
assignment
`)

	for _, code := range []string{
		"stakeholder.names.missing",
		"value.name.missing",
		"requirement.id.missing",
		"assignment.requirement.missing",
	} {
		if !hasParseDiagnostic(result.Diagnostics, code) {
			t.Fatalf("expected declaration diagnostic %q, got %+v", code, result.Diagnostics)
		}
	}
}

func hasParseDiagnostic(diags []Diagnostic, code string) bool {
	for _, diag := range diags {
		if diag.Code == code {
			return true
		}
	}
	return false
}

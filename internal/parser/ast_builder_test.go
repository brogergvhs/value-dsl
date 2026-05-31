package parser

import (
	"testing"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
)

func TestParseASTPreservesMalformedLines(t *testing.T) {
	doc := Parse(`mystery line

requirement R1
when Worker enters
unknown clause
system shall notify
stakeholders

assignment R1
Worker ->
unknown assignment line
`).AST

	if len(doc.Malformed) != 1 {
		t.Fatalf("expected 1 malformed top-level line, got %+v", doc.Malformed)
	}
	if doc.Malformed[0].Line != 1 || doc.Malformed[0].Reason != "unknown line kind" {
		t.Fatalf("unexpected malformed top-level line: %+v", doc.Malformed[0])
	}

	reqs := requirements(doc)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %+v", reqs)
	}
	req := reqs[0]
	if req.Header.First("id") != "R1" {
		t.Fatalf("unexpected requirement id: %+v", req)
	}
	if len(requirementEntries(req, grammar.RequirementClauseWhen)) != 1 {
		t.Fatalf("expected 1 preserved context clause, got %+v", req.Body)
	}
	action := firstRequirementEntry(t, req, grammar.RequirementClauseSystemShall)
	if action.First("verb") != "notify" || action.First("target") != "" {
		t.Fatalf("expected partial action to be preserved, got %+v", action)
	}
	sh := firstRequirementEntry(t, req, grammar.RequirementClauseStakeholders)
	if sh.Line == 0 || len(sh.Texts("actors")) != 0 {
		t.Fatalf("expected empty stakeholders clause to be preserved, got %+v", sh)
	}
	if len(req.Malformed) != 1 || req.Malformed[0].Line != 5 {
		t.Fatalf("expected malformed requirement body line, got %+v", req.Malformed)
	}

	assigns := assignments(doc)
	if len(assigns) != 1 {
		t.Fatalf("expected 1 assignment block, got %+v", assigns)
	}
	assign := assigns[0]
	if len(assign.Body) != 2 || assign.Body[0].Joined("stakeholders", ", ") != "Worker" || assign.Body[0].Joined("values", ", ") != "" || assign.Body[1].Joined("stakeholders", ", ") != "unknown assignment line" || assign.Body[1].HasText(ast.FieldKeyword, "->") {
		t.Fatalf("expected partial assignment entry to be preserved, got %+v", assign.Body)
	}
}

func TestParseASTPreservesPartialValueGeometry(t *testing.T) {
	doc := Parse(`value privacy_pref = 1.57,
value safety_pref = , 0.93
value authority_pref = 0.02, 0.88
`).AST

	decls := vals(doc)
	if len(decls) != 3 {
		t.Fatalf("expected 3 values, got %+v", decls)
	}

	if decls[0].Header.First("angle") == "" || decls[0].Header.First("radius") != "" || !decls[0].Header.HasText(ast.FieldKeyword, ",") {
		t.Fatalf("expected first value to preserve only angle, got %+v", decls[0].Header)
	}
	if decls[0].Header.First("angle") != "1.57" {
		t.Fatalf("unexpected first angle: %+v", decls[0].Header.First("angle"))
	}

	if decls[1].Header.First("angle") != "" || decls[1].Header.First("radius") == "" || !decls[1].Header.HasText(ast.FieldKeyword, ",") {
		t.Fatalf("expected second value to preserve only radius, got %+v", decls[1].Header)
	}
	if decls[1].Header.First("radius") != "0.93" {
		t.Fatalf("unexpected second radius: %+v", decls[1].Header.First("radius"))
	}

	if decls[2].Header.First("angle") == "" || decls[2].Header.First("radius") == "" || !decls[2].Header.HasText(ast.FieldKeyword, ",") {
		t.Fatalf("expected full geometry on third value, got %+v", decls[2].Header)
	}
}

func TestParseReferencesSkipDeclarationHeadersButKeepAssignmentHeader(t *testing.T) {
	result := Parse(`stakeholder Worker
value privacy_pref = 1.57, 0.9

requirement R1
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`)

	for _, ref := range result.References {
		if ref.Line == 1 && ref.Kind == grammar.DeclarationKindStakeholder && ref.Name == "Worker" {
			t.Fatalf("stakeholder declaration header must not be emitted as reference: %+v", ref)
		}
		if ref.Line == 2 && ref.Kind == grammar.DeclarationKindValue && ref.Name == "privacy_pref" {
			t.Fatalf("value declaration header must not be emitted as reference: %+v", ref)
		}
		if ref.Line == 4 && ref.Kind == grammar.DeclarationKindRequirement && ref.Name == "R1" {
			t.Fatalf("requirement declaration header must not be emitted as reference: %+v", ref)
		}
	}

	var assignmentHeaderRef bool
	for _, ref := range result.References {
		if ref.Line == 8 && ref.Kind == grammar.DeclarationKindRequirement && ref.Name == "R1" {
			assignmentHeaderRef = true
			break
		}
	}
	if !assignmentHeaderRef {
		t.Fatalf("expected assignment header requirement reference, got %+v", result.References)
	}
}

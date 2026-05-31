package parser

import (
	"testing"

	"github.com/brogergvhs/value-dsl/internal/ast"
)

func TestParseRecordsRequirementBodyKinds(t *testing.T) {
	result := Parse(`stakeholder Worker

requirement R1
when Worker enters DangerousArea
while Worker is in DangerousArea
system shall notify
stakeholders
priority
linked_to
unknown_clause whatever
`)

	reqs := requirements(result.AST)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %+v", reqs)
	}

	req := reqs[0]
	if len(req.Body) != 6 {
		t.Fatalf("expected 6 recognized body lines, got %+v", req.Body)
	}

	want := []string{
		"when",
		"while",
		"system_shall",
		"stakeholders",
		"priority",
		"linked_to",
	}

	for index, kind := range want {
		if req.Body[index].Kind != kind {
			t.Fatalf("body line %d kind = %q, want %q; line=%+v", index, req.Body[index].Kind, kind, req.Body[index])
		}
	}
	if len(req.Malformed) != 1 || req.Malformed[0].Line != 10 {
		t.Fatalf("expected unknown clause preserved as malformed, got %+v", req.Malformed)
	}
	for _, code := range []string{
		"parse.line.unclassified",
		"requirement.system_shall.target.missing",
		"requirement.stakeholders.actors.missing",
		"requirement.priority.level.missing",
		"requirement.linked_to.target.missing",
		"requirement.while.order.invalid",
	} {
		if !hasParseDiagnostic(result.Diagnostics, code) {
			t.Fatalf("expected diagnostic %q, got %+v", code, result.Diagnostics)
		}
	}
}

func TestParseAssignmentLinesStayEntries(t *testing.T) {
	result := Parse(`assignment R1
Worker -> privacy_pref
unknown entry
`)

	assigns := assignments(result.AST)
	if len(assigns) != 1 {
		t.Fatalf("expected one assignment block, got %+v", assigns)
	}
	block := assigns[0]
	if len(block.Body) != 2 {
		t.Fatalf("expected 2 assignment entries, got %+v", block.Body)
	}
	if !block.Body[0].HasText(ast.FieldKeyword, "->") {
		t.Fatalf("expected first entry arrow preserved, got %+v", block.Body[0])
	}
	if block.Body[1].Joined("stakeholders", ", ") != "unknown entry" || block.Body[1].HasText(ast.FieldKeyword, "->") {
		t.Fatalf("expected second body line preserved as loose assignment entry, got %+v", block.Body[1])
	}
}

func TestParsePreservesBareRecognizableLines(t *testing.T) {
	result := Parse(`requirement R1
system shall
stakeholders
priority
retention
access
linked_to
`)

	reqs := requirements(result.AST)
	if len(reqs) != 1 {
		t.Fatalf("expected one requirement, got %+v", reqs)
	}
	body := reqs[0].Body
	want := []string{
		"system_shall",
		"stakeholders",
		"priority",
		"retention",
		"access",
		"linked_to",
	}

	if len(body) != len(want) {
		t.Fatalf("unexpected body lines: %+v", body)
	}
	for index, kind := range want {
		if body[index].Kind != kind {
			t.Fatalf("body line %d kind = %q, want %q", index, body[index].Kind, kind)
		}
	}
}

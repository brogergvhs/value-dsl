package lsp

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestResolveReferencesIncludesDeclarationWhenRequested(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n"
	result := mustAnalyzeText(t, text)

	locations := resolveReferences(
		"file:///spec.dsl",
		protocol.Position{Line: 3, Character: 20},
		true,
		result,
	)
	if len(locations) != 3 {
		t.Fatalf("expected declaration plus two references, got %+v", locations)
	}
}

func TestResolveReferencesCanExcludeDeclaration(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n"
	result := mustAnalyzeText(t, text)

	locations := resolveReferences(
		"file:///spec.dsl",
		protocol.Position{Line: 3, Character: 20},
		false,
		result,
	)
	if len(locations) != 2 {
		t.Fatalf("expected references without declaration, got %+v", locations)
	}
	for _, location := range locations {
		if location.Range.Start.Line == 0 {
			t.Fatalf("did not expect declaration location in %+v", locations)
		}
	}
}

func TestResolveReferencesRequirementReferencesFromAssignmentHeader(t *testing.T) {
	text := "requirement R1\nsystem shall notify Worker\nstakeholders Worker\n\nassignment R1\nWorker -> privacy_pref\n"
	result := mustAnalyzeText(t, text)

	locations := resolveReferences(
		"file:///spec.dsl",
		protocol.Position{Line: 4, Character: 11},
		true,
		result,
	)
	if len(locations) != 2 {
		t.Fatalf("expected declaration plus assignment reference, got %+v", locations)
	}
}

func TestResolveReferencesSkipsCommentOccurrences(t *testing.T) {
	text := "stakeholder Worker\nrequirement R1\nsystem shall notify Worker // Worker\nstakeholders Worker\n"
	result := mustAnalyzeText(t, text)

	locations := resolveReferences(
		"file:///spec.dsl",
		protocol.Position{Line: 2, Character: 20},
		true,
		result,
	)
	if len(locations) != 3 {
		t.Fatalf("expected declaration plus two live references, got %+v", locations)
	}
}

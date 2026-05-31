package lsp

import (
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestResolveRenameBuildsWorkspaceEditForCurrentDocument(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n"
	result := mustAnalyzeText(t, text)

	edit, err := resolveRename(
		"file:///spec.dsl",
		protocol.Position{Line: 3, Character: 20},
		"Employee",
		result,
	)
	if err != nil {
		t.Fatalf("resolveRename() error = %v", err)
	}
	if edit == nil || len(edit.Changes["file:///spec.dsl"]) != 3 {
		t.Fatalf("expected three edits, got %+v", edit)
	}
}

func TestResolveRenameRejectsReservedKeyword(t *testing.T) {
	_, err := resolveRename(
		"file:///spec.dsl",
		protocol.Position{Line: 0, Character: 13},
		"when",
		coreanalysis.Result{},
	)
	if err == nil {
		t.Fatal("expected reserved keyword error")
	}
}

func TestResolveRenameRejectsConflictingDeclaration(t *testing.T) {
	text := "stakeholder Worker\nstakeholder Manager\n"
	result := mustAnalyzeText(t, text)

	_, err := resolveRename(
		"file:///spec.dsl",
		protocol.Position{Line: 0, Character: 13},
		"Manager",
		result,
	)
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestResolveRenameRenamesEntireRequirementIdentifier(t *testing.T) {
	text := "requirement R2\nsystem shall notify Worker\nstakeholders Worker\n\nassignment R2\nWorker -> privacy_pref\n"
	result := mustAnalyzeText(t, text)

	edit, err := resolveRename(
		"file:///spec.dsl",
		protocol.Position{Line: 4, Character: 12},
		"R2Test",
		result,
	)
	if err != nil {
		t.Fatalf("resolveRename() error = %v", err)
	}

	edits := edit.Changes["file:///spec.dsl"]
	if len(edits) != 2 {
		t.Fatalf("expected two edits, got %+v", edits)
	}
	if edits[1].Range.Start.Character != 11 || edits[1].Range.End.Character != 13 {
		t.Fatalf("unexpected assignment edit range: %+v", edits[1].Range)
	}
}

func TestResolveRenameCanRenameRequirementIdentifierToShorterName(t *testing.T) {
	text := "requirement R2Test\nsystem shall notify Worker\nstakeholders Worker\n\nassignment R2Test\nWorker -> privacy_pref\n"
	result := mustAnalyzeText(t, text)

	edit, err := resolveRename(
		"file:///spec.dsl",
		protocol.Position{Line: 4, Character: 16},
		"R2",
		result,
	)
	if err != nil {
		t.Fatalf("resolveRename() error = %v", err)
	}

	edits := edit.Changes["file:///spec.dsl"]
	if len(edits) != 2 {
		t.Fatalf("expected two edits, got %+v", edits)
	}
	if edits[1].Range.Start.Character != 11 || edits[1].Range.End.Character != 17 {
		t.Fatalf("unexpected assignment edit range: %+v", edits[1].Range)
	}
}

func TestResolveRenameRenamesRequirementIdentifierInPartiallyBrokenFile(t *testing.T) {
	text := `stakeholder Worker
value privacy_pref = 10.8,

requirement
system shall
stakeholders

requirement R2
when Worker enters DangerZone
system shall notify Worker using Alarm
stakeholders Worker

assignment R2
Worker -> privacy_pref
`
	result := mustAnalyzeText(t, text)

	edit, err := resolveRename(
		"file:///spec.dsl",
		protocol.Position{Line: 12, Character: 12},
		"R2Test",
		result,
	)
	if err != nil {
		t.Fatalf("resolveRename() error = %v", err)
	}

	edits := edit.Changes["file:///spec.dsl"]
	if len(edits) != 2 {
		t.Fatalf("expected two edits from partial file, got %+v", edits)
	}
}

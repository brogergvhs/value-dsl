package lsp

import (
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestCodeActionAddsMissingStakeholders(t *testing.T) {
	actions := actionsForText(t, `stakeholder Worker

requirement R1
when Driver enters DangerousArea
system shall notify Passenger
stakeholders Worker, Driver, Passenger
`)

	action := findCodeAction(t, actions, "Add missing stakeholders")
	edit := onlyEdit(t, action)
	if edit.NewText != "\nstakeholder Driver\nstakeholder Passenger" {
		t.Fatalf("unexpected edit text: %q", edit.NewText)
	}
	if edit.Range.Start.Line != 0 || edit.Range.Start.Character == 0 {
		t.Fatalf("expected insert at end of first line, got %+v", edit.Range)
	}
}

func TestCodeActionReplacesDuplicateValueGeometry(t *testing.T) {
	actions := actionsForText(t, `stakeholder Worker
value privacy_pref = 2.04, 0.92

requirement R1
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`)

	action := findCodeAction(t, actions, "Replace duplicate value privacy_pref with builtin privacy")
	edits := onlyURIChanges(t, action)
	if len(edits) != 2 {
		t.Fatalf("expected delete and rename edits, got %+v", edits)
	}
	if edits[0].NewText != "" {
		t.Fatalf("expected value declaration deletion, got %q", edits[0].NewText)
	}
	if edits[1].NewText != "privacy" {
		t.Fatalf("expected reference rename, got %q", edits[1].NewText)
	}
}

func TestCodeActionAddsAssignmentBlocksAndEntries(t *testing.T) {
	text := `stakeholder Worker
stakeholder Manager

requirement R1
system shall notify Worker
stakeholders Worker, Manager

requirement R2
system shall notify Worker
stakeholders Worker, Manager

assignment R2
Worker -> privacy
`
	actions := actionsForText(t, text)

	all := findCodeAction(t, actions, "Add missing assignment blocks")
	var combined strings.Builder
	for _, edit := range onlyURIChanges(t, all) {
		combined.WriteString(edit.NewText)
	}
	got := combined.String()
	if !strings.Contains(got, "\n\nassignment R1\nManager -> \nWorker -> ") {
		t.Fatalf("expected R1 assignment block, got %q", got)
	}
	if !strings.Contains(got, "\nManager -> ") {
		t.Fatalf("expected missing Manager entry, got %q", got)
	}

	current := codeActions("file:///spec.dsl", mustAnalyzeText(t, text), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
		Range:        protocol.Range{Start: protocol.Position{Line: 4}},
	})
	edit := onlyEdit(t, findCodeAction(t, current, "Add assignment block for current requirement"))
	if !strings.Contains(edit.NewText, "assignment R1") || strings.Contains(edit.NewText, "assignment R2") {
		t.Fatalf("expected only current requirement edit, got %q", edit.NewText)
	}
}

func TestCodeActionTargetsClosedWorkspaceFiles(t *testing.T) {
	root := t.TempDir()
	mainText := `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`
	writeWorkspaceFile(t, root, "main.dsl", mainText)
	writeWorkspaceFile(t, root, "values.dsl", "value privacy_pref = 2.04, 0.92\n")

	server := NewServer()
	uri := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "main.dsl")}).String())
	valuesURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "values.dsl")}).String())
	if err := server.didOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: uri, Version: 1, Text: mainText},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	result, err := server.codeAction(&glsp.Context{}, &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("codeAction() error = %v", err)
	}
	action := findCodeAction(t, result.([]protocol.CodeAction), "Replace duplicate value privacy_pref with builtin privacy")
	if action.Edit == nil || len(action.Edit.Changes[valuesURI]) != 1 {
		t.Fatalf("expected edit in closed values.dsl, got %+v", action.Edit)
	}
}

func actionsForText(t *testing.T, text string) []protocol.CodeAction {
	t.Helper()

	actions := codeActions("file:///spec.dsl", mustAnalyzeText(t, text), &protocol.CodeActionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
		Range:        protocol.Range{Start: protocol.Position{}},
	})
	if len(actions) == 0 {
		t.Fatal("expected code actions")
	}
	return actions
}

func findCodeAction(t *testing.T, actions []protocol.CodeAction, title string) protocol.CodeAction {
	t.Helper()

	for _, action := range actions {
		if action.Title == title {
			return action
		}
	}
	t.Fatalf("missing code action %q in %+v", title, actions)
	return protocol.CodeAction{}
}

func onlyEdit(t *testing.T, action protocol.CodeAction) protocol.TextEdit {
	t.Helper()

	edits := onlyURIChanges(t, action)
	if len(edits) != 1 {
		t.Fatalf("expected one edit, got %+v", edits)
	}
	return edits[0]
}

func onlyURIChanges(t *testing.T, action protocol.CodeAction) []protocol.TextEdit {
	t.Helper()

	if action.Edit == nil || len(action.Edit.Changes) != 1 {
		t.Fatalf("expected one changed URI, got %+v", action.Edit)
	}
	for _, edits := range action.Edit.Changes {
		return edits
	}
	return nil
}

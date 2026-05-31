package completion

import (
	"strings"
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestItemsSuggestTopLevelKeywords(t *testing.T) {
	items := Items("", protocol.Position{}, coreanalysis.Result{})
	if !containsLabel(items, "requirement") {
		t.Fatalf("expected top-level completion items, got %+v", labels(items))
	}
}

func TestItemsSuggestClauseKeywordsInsideRequirementContext(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\n"
	items := Items(text, protocol.Position{Line: 3, Character: 0}, coreanalysis.Result{})
	if !containsLabel(items, "when") || !containsLabel(items, "while") || !containsLabel(items, "system shall") {
		t.Fatalf("expected clause completions, got %+v", labels(items))
	}
}

func TestItemsSuggestWhenVerbsForInvalidOrPartialVerbSlot(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nwhen Worker test"
	items := Items(text, protocol.Position{Line: 3, Character: 16}, coreanalysis.Result{})
	if !containsLabel(items, "enters") || !containsLabel(items, "reports") {
		t.Fatalf("expected when-verb completions, got %+v", labels(items))
	}
}

func TestItemsSuggestActionVerbsAfterSystemShall(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall "
	items := Items(text, protocol.Position{Line: 3, Character: 13}, coreanalysis.Result{})
	if !containsLabel(items, "notify") || !containsLabel(items, "track") {
		t.Fatalf("expected action verb completions, got %+v", labels(items))
	}
}

func TestItemsSuggestActionKeywordsWithinActionShape(t *testing.T) {
	text := "requirement R1\nsystem shall track location "
	items := Items(text, protocol.Position{Line: 1, Character: 35}, coreanalysis.Result{})
	if !containsLabel(items, "of") {
		t.Fatalf("expected 'of' completion, got %+v", labels(items))
	}

	text = "requirement R1\nsystem shall notify Worker "
	items = Items(text, protocol.Position{Line: 1, Character: 27}, coreanalysis.Result{})
	if !containsLabel(items, "using") {
		t.Fatalf("expected 'using' completion, got %+v", labels(items))
	}
}

func TestItemsSuggestStakeholdersInsideActionTargetSlot(t *testing.T) {
	text := "stakeholder Worker\nstakeholder Manager\n\nrequirement R1\nsystem shall anonymize location of Wor"
	items := Items(text, protocol.Position{Line: 4, Character: 37}, buildTestResult())
	if !containsLabel(items, "Worker") {
		t.Fatalf("expected Worker completion in action target slot, got %+v", labels(items))
	}
	worker := itemByLabel(items, "Worker")
	edit, ok := worker.TextEdit.(protocol.TextEdit)
	if !ok {
		t.Fatalf("expected text edit for Worker completion, got %#v", worker.TextEdit)
	}
	if edit.Range.Start.Character != 35 || edit.Range.End.Character != 38 {
		t.Fatalf("unexpected Worker replacement range: %+v", edit.Range)
	}
}

func TestItemsSuggestStakeholdersAfterStakeholdersKeyword(t *testing.T) {
	text := "stakeholder Worker\nstakeholder Manager\n\nrequirement R1\nsystem shall notify Worker\nstakeholders "
	items := Items(text, protocol.Position{Line: 5, Character: 13}, buildTestResult())
	if !containsLabel(items, "Worker") || !containsLabel(items, "Manager") {
		t.Fatalf("expected stakeholder completions, got %+v", labels(items))
	}
}

func TestItemsSuggestStakeholdersWhileTypingFirstStakeholderName(t *testing.T) {
	text := "stakeholder Worker\nstakeholder Manager\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Wor"
	items := Items(text, protocol.Position{Line: 5, Character: 16}, buildTestResult())
	if !containsLabel(items, "Worker") {
		t.Fatalf("expected Worker completion while typing first stakeholder, got %+v", labels(items))
	}
	if containsLabel(items, "priority") {
		t.Fatalf("did not expect metadata completions while typing first stakeholder, got %+v", labels(items))
	}
}

func TestItemsDoNotSuggestStakeholdersKeywordInsideStakeholderList(t *testing.T) {
	text := "stakeholder Worker\nstakeholder Supervisor\n\nrequirement R1\nsystem shall track location of Worker using Camera\nstakeholders Worker, s"
	items := Items(text, protocol.Position{Line: 5, Character: 22}, buildTestResult())
	if containsLabel(items, "stakeholders") {
		t.Fatalf("did not expect stakeholders keyword inside stakeholder list, got %+v", labels(items))
	}
	if !containsLabel(items, "Manager") {
		t.Fatalf("expected stakeholder name suggestions, got %+v", labels(items))
	}
}

func TestItemsDoNotSuggestValueGeometryCompletions(t *testing.T) {
	text := "value privacy_pref = "
	items := Items(text, protocol.Position{Line: 0, Character: 16}, coreanalysis.Result{})
	if containsLabel(items, "self_direction") || containsLabel(items, "security") {
		t.Fatalf("did not expect category completions for custom value geometry, got %+v", labels(items))
	}
}

func TestItemsSuggestStakeholdersKeywordAfterAction(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\n"
	items := Items(text, protocol.Position{Line: 4, Character: 0}, coreanalysis.Result{})
	if !containsLabel(items, "stakeholders") {
		t.Fatalf("expected stakeholders completion, got %+v", labels(items))
	}
}

func TestItemsSuggestPartialStakeholdersKeywordAfterAction(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nst"
	items := Items(text, protocol.Position{Line: 4, Character: 2}, coreanalysis.Result{})
	if !containsLabel(items, "stakeholders") {
		t.Fatalf("expected stakeholders completion for partial prefix, got %+v", labels(items))
	}
	stakeholders := itemByLabel(items, "stakeholders")
	edit, ok := stakeholders.TextEdit.(protocol.TextEdit)
	if !ok {
		t.Fatalf("expected text edit for stakeholders completion, got %#v", stakeholders.TextEdit)
	}
	if edit.Range.Start.Character != 0 || edit.Range.End.Character != 2 {
		t.Fatalf("unexpected stakeholders replacement range: %+v", edit.Range)
	}
}

func TestItemsSuggestStakeholdersAfterInvalidRequirementLines(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R4\nwhen Worker enters Dungeop\nsystem shall anonymize Worker\nst"
	result := coreanalysis.Result{}
	items := Items(text, protocol.Position{Line: 5, Character: 2}, result)
	if !containsLabel(items, "stakeholders") {
		t.Fatalf("expected stakeholders completion after invalid requirement lines, got %+v", labels(items))
	}
}

func TestItemsSuggestMetadataKeywordsAfterStakeholders(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n"
	items := Items(text, protocol.Position{Line: 5, Character: 0}, buildTestResult())
	if !containsLabel(items, "priority") || !containsLabel(items, "linked_to") {
		t.Fatalf("expected metadata completions, got %+v", labels(items))
	}
}

func TestItemsSuggestAssignmentSymbols(t *testing.T) {
	text := "assignment "
	result := buildTestResult()
	items := Items(text, protocol.Position{Line: 0, Character: 11}, result)
	if !containsLabel(items, "R1") {
		t.Fatalf("expected requirement completions, got %+v", labels(items))
	}

	text = "assignment R1\n"
	text += "Worker -> "
	items = Items(text, protocol.Position{Line: 1, Character: 10}, result)
	if !containsLabel(items, "privacy") {
		t.Fatalf("expected value completions after arrow, got %+v", labels(items))
	}
	if !containsLabel(items, "safety") {
		t.Fatalf("expected built-in value completions after arrow, got %+v", labels(items))
	}
	valueItem := itemByLabel(items, "privacy")
	doc, ok := valueItem.Documentation.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("expected value documentation markup, got %+v", valueItem.Documentation)
	}
	if !strings.Contains(doc.Value, "Category: `self_direction`") || !strings.Contains(doc.Value, "Radius:") || !strings.Contains(doc.Value, "deg") {
		t.Fatalf("expected value documentation to include geometry and dial, got %q", doc.Value)
	}
}

func TestItemsSuggestTopLevelKeywordsBelowCompletedRequirementBlock(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n\n"
	items := Items(text, protocol.Position{Line: 6, Character: 0}, coreanalysis.Result{})
	if !containsLabel(items, "requirement") || !containsLabel(items, "assignment") {
		t.Fatalf("expected top-level completions below block, got %+v", labels(items))
	}
}

func TestItemsFallsBackToLatestSuccessfulSymbols(t *testing.T) {
	text := "requirement R1\nsystem shall notify Worker\nstakeholders "
	items := Items(text, protocol.Position{Line: 2, Character: 13}, buildTestResult())
	if !containsLabel(items, "Worker") {
		t.Fatalf("expected fallback symbol completions, got %+v", labels(items))
	}
}

func TestItemsDoNotCompleteInsideComments(t *testing.T) {
	text := "requirement R1\nsystem shall notify Worker // st\nstakeholders Worker\n"
	items := Items(text, protocol.Position{Line: 1, Character: 34}, coreanalysis.Result{})
	if len(items) != 0 {
		t.Fatalf("expected no completion items inside comment, got %+v", labels(items))
	}
}

func buildTestResult() coreanalysis.Result {
	model := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{
			{Line: 1, Name: "Worker"},
			{Line: 2, Name: "Manager"},
		},
		Values: []model.Value{
			{Line: 3, Name: "privacy_pref"},
		},
		Requirements: []model.Requirement{
			{Line: 4, ID: "R1"},
		},
	}
	return coreanalysis.Result{Index: docindex.Build(nil, nil, model, nil, nil, nil, nil)}
}

func containsLabel(items []protocol.CompletionItem, label string) bool {
	for _, item := range items {
		if item.Label == label {
			return true
		}
	}
	return false
}

func labels(items []protocol.CompletionItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Label)
	}
	return out
}

func itemByLabel(items []protocol.CompletionItem, label string) protocol.CompletionItem {
	for _, item := range items {
		if item.Label == label {
			return item
		}
	}
	return protocol.CompletionItem{}
}

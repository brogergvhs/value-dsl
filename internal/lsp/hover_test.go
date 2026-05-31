package lsp

import (
	"strings"
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestResolveHoverReturnsKeywordHoverFromLangspec(t *testing.T) {
	text := "requirement R1\nwhen Worker enters DangerousArea\n"

	result := resolveHover(text, protocol.Position{Line: 1, Character: 1}, coreanalysis.Result{})
	if result == nil {
		t.Fatal("expected hover result")
	}

	content, ok := result.Contents.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("unexpected hover contents: %#v", result.Contents)
	}
	if !strings.Contains(content.Value, "EARS event-driven clause") {
		t.Fatalf("unexpected hover text: %q", content.Value)
	}
}

func TestResolveHoverReturnsContextualMetadataHover(t *testing.T) {
	text := "requirement R1\nsystem shall notify Worker\nstakeholders Worker\npriority high\n"

	result := resolveHover(text, protocol.Position{Line: 3, Character: 10}, coreanalysis.Result{})
	if result == nil {
		t.Fatal("expected hover result")
	}

	content := result.Contents.(protocol.MarkupContent)
	if !strings.Contains(content.Value, "Priority value `high`") {
		t.Fatalf("unexpected hover text: %q", content.Value)
	}
}

func TestResolveHoverReturnsRichStakeholderHover(t *testing.T) {
	model := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{{Line: 1, Name: "Worker"}},
		StakeholderByName: map[string]model.Stakeholder{
			"Worker": {Line: 1, Name: "Worker"},
		},
		Requirements: []model.Requirement{
			{
				Line: 2,
				ID:   "R1",
				Context: model.RequirementContext{
					Clauses: []model.Clause{
						{Line: 3, Type: grammar.RequirementClauseWhen, Actor: "Worker", Verb: "enters", Target: "DangerousArea", RawText: "when Worker enters DangerousArea"},
					},
				},
				Stakeholders: []model.Stakeholder{{Name: "Worker"}},
			},
		},
		AssignmentEntries: []model.AssignmentEntry{
			{RequirementID: "R1", Stakeholders: []model.Ref{{Name: "Worker"}}},
		},
	}

	text := "stakeholder Worker\n"
	result := resolveHover(text, protocol.Position{Line: 0, Character: 15}, coreanalysis.Result{Index: docindex.Build(nil, nil, model, nil, nil, nil, nil)})
	if result == nil {
		t.Fatal("expected hover result")
	}

	content := result.Contents.(protocol.MarkupContent)
	if !strings.Contains(content.Value, "Requirement clauses: R1") || !strings.Contains(content.Value, "Assignments in: R1") {
		t.Fatalf("unexpected stakeholder hover text: %q", content.Value)
	}
}

func TestResolveHoverReturnsRequirementHoverWithMetadataAndTraceability(t *testing.T) {
	model := &semantic.SemanticModel{
		Requirements: []model.Requirement{
			{
				Line: 1,
				ID:   "R1",
				Action: model.Action{
					RawText: "notify Worker",
				},
				Stakeholders: []model.Stakeholder{{Name: "Worker"}},
				Metadata: model.RequirementMetadata{
					Priority:  model.MetadataField{Value: "high"},
					Retention: model.MetadataField{Value: "short_term"},
				},
				Traceability: []model.TraceabilityLink{{Value: "HazardAnalysis"}},
			},
		},
	}

	text := "requirement R1\n"
	result := resolveHover(text, protocol.Position{Line: 0, Character: 13}, coreanalysis.Result{Index: docindex.Build(nil, nil, model, nil, nil, nil, nil)})
	if result == nil {
		t.Fatal("expected hover result")
	}

	content := result.Contents.(protocol.MarkupContent)
	if !strings.Contains(content.Value, "Metadata: priority=high, retention=short_term") || !strings.Contains(content.Value, "Traceability: HazardAnalysis") {
		t.Fatalf("unexpected requirement hover text: %q", content.Value)
	}
}

func TestResolveHoverIgnoresTokensInsideComments(t *testing.T) {
	text := "stakeholder Worker // comment Worker\n"
	result := resolveHover(text, protocol.Position{Line: 0, Character: 34}, coreanalysis.Result{})
	if result != nil {
		t.Fatalf("expected no hover inside comment, got %+v", result)
	}
}

func TestResolveHoverReturnsInvalidCustomValueHoverForIncompleteGeometry(t *testing.T) {
	text := "value authority_pref =\n"
	result := mustAnalyzeText(t, text)

	hover := resolveHover(text, protocol.Position{Line: 0, Character: 8}, result)
	if hover == nil {
		t.Fatal("expected hover result")
	}

	content := hover.Contents.(protocol.MarkupContent)
	if !strings.Contains(content.Value, "**invalid custom value** `authority_pref`") {
		t.Fatalf("unexpected hover text: %q", content.Value)
	}
	if strings.Contains(content.Value, "Angle: `0.00 rad`") || strings.Contains(content.Value, "Radius: `0.00`") {
		t.Fatalf("expected no synthetic geometry in invalid hover, got %q", content.Value)
	}
}

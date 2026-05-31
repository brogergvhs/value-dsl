package lsp

import (
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestBuildDocumentSymbolsBuildsHierarchicalDocumentSymbols(t *testing.T) {
	semanticModel := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{
			{Line: 1, Range: sourcepos.NewRange(sourcepos.NewPosition(1, 13), sourcepos.NewPosition(1, 18)), Name: "Worker"},
		},
		Values: []model.Value{
			{Line: 3, Range: sourcepos.NewRange(sourcepos.NewPosition(3, 7), sourcepos.NewPosition(3, 13)), Name: "privacy_pref", Category: "stimulation"},
		},
		Requirements: []model.Requirement{
			{
				Line:  5,
				Range: sourcepos.NewRange(sourcepos.NewPosition(5, 13), sourcepos.NewPosition(5, 14)),
				ID:    "R1",
				Context: model.RequirementContext{
					Clauses: []model.Clause{
						func() model.Clause {
							clause := model.Clause{Line: 6, Type: grammar.RequirementClauseWhen, Actor: "Worker", Verb: "enters", Target: "DangerousArea", RawText: "when Worker enters DangerousArea"}
							clause.Range = sourcepos.NewRange(sourcepos.NewPosition(6, 1), sourcepos.NewPosition(6, 4))
							return clause
						}(),
					},
				},
				Action: model.Action{
					Line:    7,
					Range:   sourcepos.NewRange(sourcepos.NewPosition(7, 1), sourcepos.NewPosition(7, 12)),
					RawText: "notify Worker",
				},
				StakeholdersLine:  8,
				StakeholdersRange: sourcepos.NewRange(sourcepos.NewPosition(8, 1), sourcepos.NewPosition(8, 12)),
				Stakeholders:      []model.Stakeholder{{Name: "Worker"}},
				Metadata: model.RequirementMetadata{
					Priority: model.MetadataField{Line: 9, Range: sourcepos.NewRange(sourcepos.NewPosition(9, 1), sourcepos.NewPosition(9, 8)), Value: "high"},
				},
				Traceability: []model.TraceabilityLink{
					{Line: 10, Range: sourcepos.NewRange(sourcepos.NewPosition(10, 1), sourcepos.NewPosition(10, 9)), Value: "HazardAnalysis"},
				},
			},
		},
	}

	symbols := buildDocumentSymbols(coreanalysis.Result{
		Index: docindex.Build(nil, nil, semanticModel, nil, nil, nil, nil),
	})
	if len(symbols) != 3 {
		t.Fatalf("expected 3 top-level symbols, got %+v", symbols)
	}

	requirement := symbols[2]
	if requirement.Name != "R1" || requirement.Kind != protocol.SymbolKindObject {
		t.Fatalf("unexpected requirement symbol: %+v", requirement)
	}
	if requirement.Range.Start.Character != 12 {
		t.Fatalf("expected requirement to start at column 13, got %+v", requirement.Range)
	}
	if len(requirement.Children) < 4 {
		t.Fatalf("expected requirement children, got %+v", requirement.Children)
	}
}

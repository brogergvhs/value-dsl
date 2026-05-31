package lsp

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestResolveDefinitionStakeholderReference(t *testing.T) {
	text := "stakeholder Worker\n\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n"
	result := mustAnalyzeText(t, text)

	locations := resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 3, Character: 20},
		result,
	)
	if len(locations) != 1 {
		t.Fatalf("expected one location, got %+v", locations)
	}
	if locations[0].Range.Start.Line != 0 || locations[0].Range.Start.Character != 12 {
		t.Fatalf("unexpected location: %+v", locations[0])
	}
}

func TestResolveDefinitionAssignmentValueReference(t *testing.T) {
	text := "value privacy_pref = 1.58, 0.91\n\nassignment R1\nWorker -> privacy_pref\n"
	result := mustAnalyzeText(t, text)

	locations := resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 3, Character: 18},
		result,
	)
	if len(locations) != 1 {
		t.Fatalf("expected one location, got %+v", locations)
	}
	if locations[0].Range.Start.Line != 0 || locations[0].Range.Start.Character != 6 {
		t.Fatalf("unexpected location: %+v", locations[0])
	}
}

func TestResolveDefinitionAssignmentRequirementReference(t *testing.T) {
	text := "requirement R1\nsystem shall notify Worker\nstakeholders Worker\n\nassignment R1\nWorker -> privacy_pref\n"
	result := mustAnalyzeText(t, text)

	locations := resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 4, Character: 11},
		result,
	)
	if len(locations) != 1 {
		t.Fatalf("expected one location, got %+v", locations)
	}
	if locations[0].Range.Start.Line != 0 || locations[0].Range.Start.Character != 12 {
		t.Fatalf("unexpected location: %+v", locations[0])
	}
}

func TestResolveDefinitionIgnoresCommentTokens(t *testing.T) {
	text := "stakeholder Worker // comment Worker\n"
	result := mustAnalyzeText(t, text)

	locations := resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 0, Character: 34},
		result,
	)
	if len(locations) != 0 {
		t.Fatalf("expected no locations inside comment, got %+v", locations)
	}
}

func TestResolveDefinitionAssignmentRequirementReferencesInMultiRequirementDocument(t *testing.T) {
	text := `stakeholder Worker
stakeholder Supervisor
stakeholder SafetyOfficer

value privacy_pref = 1.58, 0.91
value safety_pref = 4.42, 0.93
value authority_pref = 0.02, 0.88

requirement R1
when Worker enters DangerousArea
system shall notify Worker
stakeholders Worker

requirement R2
if Worker has_no_protective_equipement
system shall notify Worker using AudioAlarm
stakeholders Worker, SafetyOfficer

requirement R3
where SafetyMode enabled
system shall log location of Worker using Database
stakeholders Worker, Supervisor

assignment R2
Worker -> privacy_pref

// test comment here
assignment R3 // test
Worker -> authority_pref
`
	result := mustAnalyzeText(t, text)

	locations := resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 23, Character: 11},
		result,
	)
	if len(locations) != 1 {
		t.Fatalf("expected one R2 definition location, got %+v", locations)
	}
	if locations[0].Range.Start.Line != 13 || locations[0].Range.Start.Character != 12 {
		t.Fatalf("unexpected R2 definition location: %+v", locations[0])
	}

	locations = resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 27, Character: 11},
		result,
	)
	if len(locations) != 1 {
		t.Fatalf("expected one R3 definition location, got %+v", locations)
	}
	if locations[0].Range.Start.Line != 18 || locations[0].Range.Start.Character != 12 {
		t.Fatalf("unexpected R3 definition location: %+v", locations[0])
	}
}

func TestResolveDefinitionRequirementDefinitionInPartiallyBrokenFile(t *testing.T) {
	text := `stakeholder Worker
value privacy_pref = 10.8,
unknown top level

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

	locations := resolveDefinition(
		"file:///spec.dsl",
		protocol.Position{Line: 13, Character: 11},
		result,
	)
	if len(locations) != 1 {
		t.Fatalf("expected one location from partial file, got %+v", locations)
	}
	if locations[0].Range.Start.Line != 8 || locations[0].Range.Start.Character != 12 {
		t.Fatalf("unexpected definition location from partial file: %+v", locations[0])
	}
}

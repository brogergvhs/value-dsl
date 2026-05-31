package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/parser"
	"github.com/brogergvhs/value-dsl/internal/semantic"
)

func mustValidate(t *testing.T, m *semantic.SemanticModel) []Diagnostic {
	t.Helper()
	return Validate(m)
}

func requirement(id string) model.Requirement {
	return model.Requirement{
		Line: 3,
		ID:   id,
		Action: model.Action{
			Line:   4,
			Verb:   "notify",
			Target: "Worker",
		},
		StakeholdersLine: 5,
		Stakeholders:     []model.Stakeholder{{Name: "Worker"}},
	}
}

func TestValidateDeclarationSemantics(t *testing.T) {
	m := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{{Line: 1, Name: "Worker"}, {Line: 2, Name: "Worker"}},
		Values: []model.Value{
			{Line: 4, Name: "privacy_pref", Angle: 10.0, Radius: 1.10, HasAngle: true, HasRadius: true},
			{Line: 5, Name: "privacy"},
		},
		Requirements: []model.Requirement{
			requirement("R1"),
			{Line: 6, ID: "R1", Action: model.Action{Target: "Worker"}},
		},
	}

	diags := mustValidate(t, m)
	for _, code := range []DiagnosticCode{
		CodeStakeholderDuplicate,
		CodeValueAngleOutOfBounds,
		CodeValueRadiusOutOfBounds,
		CodeValueBuiltinConflict,
		CodeRequirementDuplicate,
	} {
		if !hasDiagCode(diags, code) {
			t.Fatalf("expected %s, got %+v", code, diags)
		}
	}
}

func TestValidateRequirementReferences(t *testing.T) {
	m := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{{Line: 1, Name: "Worker"}},
		Requirements: []model.Requirement{{
			Line: 3,
			ID:   "R1",
			Context: model.RequirementContext{Clauses: []model.Clause{{
				Line:  4,
				Type:  grammar.RequirementClauseWhen,
				Actor: "UnknownActor",
			}}},
			Action:           model.Action{Line: 5, Target: "UnknownTarget"},
			StakeholdersLine: 6,
			Stakeholders:     []model.Stakeholder{{Name: "UnknownStakeholder"}},
		}},
	}

	diags := mustValidate(t, m)
	for _, code := range []DiagnosticCode{
		CodeRequirementContextUnknownStakeholder,
		CodeRequirementActionTargetUnknown,
		CodeRequirementStakeholderUnknown,
	} {
		if !hasDiagCode(diags, code) {
			t.Fatalf("expected %s, got %+v", code, diags)
		}
	}
}

func TestValidateAssignmentSemantics(t *testing.T) {
	m := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{{Line: 1, Name: "Worker"}},
		Values:       []model.Value{{Line: 2, Name: "privacy_pref", HasAngle: true, HasRadius: true}},
		Requirements: []model.Requirement{requirement("R1")},
		AssignmentEntries: []model.AssignmentEntry{
			{HeaderLine: 7, RequirementID: "R2"},
			{Line: 8, RequirementID: "R2", Stakeholders: []model.Ref{{Name: "Unknown"}}, Values: []model.Ref{{Name: "missing_value"}}},
			{Line: 9, RequirementID: "R1", Stakeholders: []model.Ref{{Name: "Worker"}}, Values: []model.Ref{{Name: "privacy_pref"}}},
			{Line: 10, RequirementID: "R1", Stakeholders: []model.Ref{{Name: "Worker"}}, Values: []model.Ref{{Name: "privacy_pref"}}},
		},
	}

	diags := mustValidate(t, m)
	for _, code := range []DiagnosticCode{
		CodeAssignmentRequirementUnknown,
		CodeAssignmentStakeholderUnknown,
		CodeAssignmentValueUnknown,
		CodeAssignmentDuplicate,
	} {
		if !hasDiagCode(diags, code) {
			t.Fatalf("expected %s, got %+v", code, diags)
		}
	}
}

func TestValidateMarksInvalidDeclarations(t *testing.T) {
	m := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{{Line: 1, Name: "Worker"}, {Line: 2, Name: "Worker"}},
		StakeholderByName: map[string]model.Stakeholder{
			"Worker": {Line: 1, Name: "Worker"},
		},
		Values: []model.Value{{Line: 3, Name: "privacy_pref", Angle: 10.8, Radius: 0.91, HasAngle: true, HasRadius: true}},
		ValueByName: map[string]model.Value{
			"privacy_pref": {Line: 3, Name: "privacy_pref", Angle: 10.8, Radius: 0.91},
		},
		Requirements: []model.Requirement{requirement("R1"), {Line: 4, ID: "R1"}},
		AssignmentEntries: []model.AssignmentEntry{
			{Line: 8, RequirementID: "R1", Stakeholders: []model.Ref{{Name: "Worker"}}, Values: []model.Ref{{Name: "privacy_pref"}}},
		},
	}

	_ = mustValidate(t, m)
	if !m.Stakeholders[1].Invalid || !m.Values[0].Invalid || !m.Requirements[1].Invalid {
		t.Fatalf("expected invalid declarations to be marked, got %+v", m)
	}
}

func TestValidateRestoresWarnings(t *testing.T) {
	m := &semantic.SemanticModel{
		Stakeholders: []model.Stakeholder{
			{Line: 1, Name: "Worker"},
			{Line: 2, Name: "Unused"},
		},
		Values: []model.Value{
			{Line: 3, Name: "privacy_pref", Angle: 10.8, Radius: 0.91, HasAngle: true, HasRadius: true},
		},
		Requirements: []model.Requirement{{
			Line:             4,
			ID:               "R1",
			StakeholdersLine: 6,
			Stakeholders:     []model.Stakeholder{{Name: "Worker"}},
			Action:           model.Action{Line: 5, Target: "Worker"},
		}},
		AssignmentEntries: []model.AssignmentEntry{
			{Line: 7, HeaderLine: 7, RequirementID: "R1", Stakeholders: []model.Ref{{Name: "Worker"}}, Values: []model.Ref{{Name: "privacy_pref"}}},
		},
	}

	diags := mustValidate(t, m)
	if !hasDiagCode(diags, CodeStakeholderUnused) {
		t.Fatalf("expected %s, got %+v", CodeStakeholderUnused, diags)
	}
	if !hasDiagCode(diags, CodeReferenceInvalidDeclaration) {
		t.Fatalf("expected %s, got %+v", CodeReferenceInvalidDeclaration, diags)
	}
	if !hasWarningCode(diags, CodeStakeholderUnused) || !hasWarningCode(diags, CodeReferenceInvalidDeclaration) {
		t.Fatalf("expected warning severity for restored warnings, got %+v", diags)
	}
}

func TestValidateCanonicalExamplesProduceNoDiagnostics(t *testing.T) {
	files := []string{
		"01_minimal_valid.dsl",
		"07_valid_notification.dsl",
		"08_valid_persistence.dsl",
		"09_valid_restriction.dsl",
		"17_valid_inline_assignments.dsl",
		"19_valid_metadata.dsl",
		"20_valid_traceability.dsl",
		"21_valid_ubiquitous.dsl",
		"22_valid_while.dsl",
		"23_valid_if.dsl",
		"24_valid_where.dsl",
		"25_valid_while_when.dsl",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			doc := parser.Parse(readValidationExampleDSL(t, file)).AST
			model, err := semantic.Build(doc)
			if err != nil {
				t.Fatalf("Build(%s) error = %v", file, err)
			}
			if diags := mustValidate(t, model); len(diags) != 0 {
				t.Fatalf("expected no diagnostics for %s, got %+v", file, diags)
			}
		})
	}
}

func hasDiagCode(diags []Diagnostic, code DiagnosticCode) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func hasWarningCode(diags []Diagnostic, code DiagnosticCode) bool {
	for _, d := range diags {
		if d.Code == code && d.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

func readValidationExampleDSL(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "examples", "dsl", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}

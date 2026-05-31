package analysis

import (
	"strings"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/validation"
)

func TestRunBuildsConflictsForValidDSL(t *testing.T) {
	dsl := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91
value authority_pref = 0.02, 0.88

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager

assignment R1
Worker -> privacy_pref
Manager -> authority_pref
`

	result, err := Run(dsl, Options{Strict: true, BuildConflicts: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Index == nil || result.Index.Model == nil {
		t.Fatal("expected non-nil model")
	}
	if len(result.Index.Model.AssignmentEntries) != 2 {
		t.Fatalf("expected 2 inline assignments, got %+v", result.Index.Model.AssignmentEntries)
	}
	if len(result.AllDiagnostics()) != 0 {
		t.Fatalf("expected no diagnostics, got %+v", result.AllDiagnostics())
	}
	if len(result.Index.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %+v", result.Index.Conflicts)
	}
	conflict := result.Index.Conflicts[0]
	if conflict.ValueA != "authority_pref" || conflict.ValueB != "privacy_pref" {
		t.Fatalf("expected conflict to carry value pair, got %+v", conflict)
	}
}

func TestRunRejectsStrictParseDiagnostics(t *testing.T) {
	_, err := Run("mystery line", Options{Strict: true})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "strict policy rejected result") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunRejectsStrictValidationErrors(t *testing.T) {
	dsl := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Manager
`

	_, err := Run(dsl, Options{Strict: true})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "strict policy rejected result") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeConflictsSkipsInvalidDeclarations(t *testing.T) {
	dsl := `stakeholder Worker
stakeholder Supervisor

value privacy_pref = 1.58, 0.91
value authority_pref = 10.8, 0.88

requirement R1
when Worker enters DangerousArea
while Worker is in DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Supervisor

assignment R1
Worker -> privacy_pref
Supervisor -> authority_pref
`

	result, err := Run(dsl, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Index == nil || len(result.Index.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics for invalid declarations, got none")
	}
	if conflicts := AnalyzeConflicts(result.Index.Model); len(conflicts) != 0 {
		t.Fatalf("expected no conflicts when declarations invalid, got %+v", conflicts)
	}
}

func TestRunPropagatesInvalidDefinitionsIntoAssignments(t *testing.T) {
	dsl := `stakeholder Worker

value privacy_pref = 10.8, 0.91

requirement R1
system shall inspect Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`

	result, err := Run(dsl, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !hasDiagnosticAt(result.AllDiagnostics(), validation.CodeReferenceInvalidDeclaration, 9) ||
		!hasDiagnosticAt(result.AllDiagnostics(), validation.CodeReferenceInvalidDeclaration, 10) {
		t.Fatalf("expected propagated invalid-declaration warning on assignment line, got %+v", result.AllDiagnostics())
	}
}

func TestAnalyzeConflictsDeduplicatesEffectiveAssignmentPairs(t *testing.T) {
	model := &semantic.SemanticModel{
		Requirements: []model.Requirement{{ID: "R1"}},
		Stakeholders: []model.Stakeholder{{Name: "Worker"}, {Name: "Manager"}},
		Values: []model.Value{
			{Name: "privacy_pref", Angle: 1.58, Radius: 0.91},
			{Name: "authority_pref", Angle: 0.02, Radius: 0.88},
		},
		AssignmentEntries: []model.AssignmentEntry{
			{
				RequirementID: "R1",
				Stakeholders:  []model.Ref{{Name: "Worker"}, {Name: "Worker"}, {Name: "Manager"}},
				Values:        []model.Ref{{Name: "privacy_pref"}, {Name: "privacy_pref"}, {Name: "authority_pref"}},
			},
			{
				RequirementID: "R1",
				Stakeholders:  []model.Ref{{Name: "Worker"}},
				Values:        []model.Ref{{Name: "privacy_pref"}},
			},
		},
	}

	conflicts := AnalyzeConflicts(model)
	if len(conflicts) != 4 {
		t.Fatalf("expected 4 unique value-pair conflicts, got %+v", conflicts)
	}
	for _, conflict := range conflicts {
		if conflict.ValueA == "" || conflict.ValueB == "" {
			t.Fatalf("expected conflict to carry value names, got %+v", conflict)
		}
	}
}

func hasDiagnosticAt(diags []validation.Diagnostic, code validation.DiagnosticCode, line int) bool {
	for _, diag := range diags {
		if diag.Code == code && diag.Line == line {
			return true
		}
	}
	return false
}

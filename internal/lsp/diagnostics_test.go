package lsp

import (
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/validation"
)

func TestDiagnosticsFromResultUsesBuildErrorLocationWhenAvailable(t *testing.T) {
	diagnostics := diagnosticsFromResult(coreanalysis.Result{
		BuildErr: assertError(`semantic error (line 5): [value "privacy_pref"] duplicate value "privacy_pref"`),
	})

	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(diagnostics))
	}
	if diagnostics[0].Range.Start.Line != 4 || diagnostics[0].Range.Start.Character != 0 {
		t.Fatalf("unexpected diagnostic range: %+v", diagnostics[0].Range)
	}
}

func TestDiagnosticsFromResultIncludesStructuredParseDiagnosticsAndBuildError(t *testing.T) {
	diagnostics := diagnosticsFromResult(coreanalysis.Result{
		Index: docindex.Build(nil, nil, nil, []validation.Diagnostic{
			{
				Line:     2,
				Severity: validation.SeverityError,
				Message:  "empty where clause",
			},
			{
				Line:     4,
				Severity: validation.SeverityError,
				Message:  "missing assignment arrow",
			},
		}, nil, nil, nil),
		BuildErr: assertError(`semantic error (line 5): [value "privacy_pref"] duplicate value "privacy_pref"`),
	})

	if len(diagnostics) != 3 {
		t.Fatalf("expected three diagnostics, got %d", len(diagnostics))
	}
	if diagnostics[0].Message != "empty where clause" || diagnostics[0].Range.Start.Line != 1 {
		t.Fatalf("unexpected first diagnostic: %+v", diagnostics[0])
	}
	if diagnostics[1].Message != "missing assignment arrow" || diagnostics[1].Range.Start.Line != 3 {
		t.Fatalf("unexpected second diagnostic: %+v", diagnostics[1])
	}
	if diagnostics[2].Message != `semantic error (line 5): [value "privacy_pref"] duplicate value "privacy_pref"` || diagnostics[2].Range.Start.Line != 4 {
		t.Fatalf("unexpected third diagnostic: %+v", diagnostics[2])
	}
}

func TestDiagnosticsFromResultConvertsValidationDiagnostics(t *testing.T) {
	diagnostics := diagnosticsFromResult(coreanalysis.Result{
		Index: docindex.Build(nil, nil, nil, nil, []validation.Diagnostic{
			{
				Line:     3,
				Severity: validation.SeverityWarning,
				Message:  "warning message",
			},
		}, nil, nil),
	})

	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(diagnostics))
	}
	if diagnostics[0].Range.Start.Line != 2 {
		t.Fatalf("unexpected line: %+v", diagnostics[0].Range)
	}
	if diagnostics[0].Severity == nil || *diagnostics[0].Severity != 2 {
		t.Fatalf("unexpected severity: %+v", diagnostics[0].Severity)
	}
}

type staticError string

func (e staticError) Error() string {
	return string(e)
}

func assertError(message string) error {
	return staticError(message)
}

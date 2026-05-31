package analysis

import (
	"errors"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/validation"
)

func TestAnalyzeTextBuildsModelAndDiagnostics(t *testing.T) {
	result, err := Run(`stakeholder Worker

value privacy_pref = 10.8,

requirement R1
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.BuildErr != nil {
		t.Fatalf("unexpected build error: %v", result.BuildErr)
	}
	if result.Index == nil || result.Index.Model == nil {
		t.Fatal("expected semantic model")
	}
	if len(result.Index.ParseDiagnostics) == 0 || result.Index.ParseDiagnostics[0].Code != validation.DiagnosticCode("value.radius.missing") {
		t.Fatalf("expected grammar parse diagnostic for missing radius, got %+v", result.Index.ParseDiagnostics)
	}
	if len(result.Index.Diagnostics) == 0 {
		t.Fatal("expected validation diagnostics")
	}
}

func TestEnforceStrictPolicyRejectsParseDiagnostics(t *testing.T) {
	_, err := Run(`stakeholder Worker
mystery line

requirement R1
system shall notify Worker
stakeholders Worker
`, Options{Strict: true})
	if err == nil {
		t.Fatal("expected strict policy rejection")
	}
	var strictErr *StrictPolicyError
	if !errors.As(err, &strictErr) {
		t.Fatalf("expected StrictPolicyError, got %T", err)
	}
	if len(strictErr.Diagnostics) != 1 || strictErr.Diagnostics[0].Code != validation.DiagnosticCode("parse.line.unclassified") {
		t.Fatalf("unexpected strict diagnostics: %+v", strictErr.Diagnostics)
	}
}

func TestEnforceStrictPolicyRejectsValidationErrors(t *testing.T) {
	_, err := Run(`stakeholder Worker

value privacy_pref = 10.8,

requirement R1
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`, Options{Strict: true})
	if err == nil {
		t.Fatal("expected strict policy rejection")
	}
	var strictErr *StrictPolicyError
	if !errors.As(err, &strictErr) {
		t.Fatalf("expected StrictPolicyError, got %T", err)
	}
	if len(strictErr.Diagnostics) == 0 {
		t.Fatalf("expected strict diagnostics, got %+v", strictErr.Diagnostics)
	}
}

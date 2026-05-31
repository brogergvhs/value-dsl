// Package analysis runs the canonical DSL analysis pipeline.
package analysis

import (
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/validation"
)

type Options struct {
	Strict         bool
	BuildConflicts bool
	StrictConfig   StrictConfig
}

type StrictConfig struct {
	FailOnWarnings bool
}

type StrictPolicyError struct {
	Diagnostics []validation.Diagnostic
}

func (e *StrictPolicyError) Error() string {
	return "strict policy rejected result"
}

type Result struct {
	Index    *docindex.Document
	BuildErr error
}

func (r Result) AllDiagnostics() []validation.Diagnostic {
	if r.Index == nil {
		return nil
	}
	return r.Index.AllDiagnostics()
}

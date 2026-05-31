package analysis

import (
	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/parser"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/validation"
)

func Run(text string, options Options) (Result, error) {
	result, err := analyzeParsed(parser.Parse(text), options)
	if result.BuildErr != nil {
		return result, result.BuildErr
	}
	return result, err
}

func analyzeParsed(parsed parser.ParseResult, options Options) (Result, error) {
	parseDiags := parseDiagnostics(parsed.Diagnostics)
	result := Result{
		Index: docindex.Build(parsed.AST, parsed.Lines, nil, parseDiags, nil, nil, nil),
	}

	semanticModel, err := semantic.Build(parsed.AST)
	if err != nil {
		result.BuildErr = err
		return result, nil
	}

	markDeclarationsInvalidFromParse(semanticModel, parsed.AST, parsed.Diagnostics)
	diagnostics := validation.Validate(semanticModel)
	var conflicts []model.ConflictResult
	if options.BuildConflicts {
		conflicts = AnalyzeConflicts(semanticModel)
	}
	result.Index = docindex.Build(parsed.AST, parsed.Lines, semanticModel, parseDiags, diagnostics, conflicts, indexReferences(parsed.References))
	err = enforceStrictPolicy(result.AllDiagnostics(), options)

	return result, err
}

func markDeclarationsInvalidFromParse(m *semantic.SemanticModel, doc *ast.DocumentNode, diags []parser.Diagnostic) {
	if m == nil || doc == nil || len(diags) == 0 {
		return
	}

	lines := make(map[int]struct{}, len(diags))
	for _, d := range diags {
		lines[d.Line] = struct{}{}
	}
	for _, decl := range doc.Declarations {
		if !declarationHasParseIssue(decl, lines) {
			continue
		}

		line := decl.Line
		switch decl.Kind {
		case grammar.DeclarationKindStakeholder:
			for i := range m.Stakeholders {
				if m.Stakeholders[i].Line == line {
					m.Stakeholders[i].Invalid = true
				}
			}
			for k, v := range m.StakeholderByName {
				if v.Line == line {
					v.Invalid = true
					m.StakeholderByName[k] = v
				}
			}
		case grammar.DeclarationKindValue:
			for i := range m.Values {
				if m.Values[i].Line == line {
					m.Values[i].Invalid = true
				}
			}
			for k, v := range m.ValueByName {
				if v.Line == line {
					v.Invalid = true
					m.ValueByName[k] = v
				}
			}
		case grammar.DeclarationKindRequirement:
			for i := range m.Requirements {
				if m.Requirements[i].Line == line {
					m.Requirements[i].Invalid = true
				}
			}
		}
	}
}

func declarationHasParseIssue(decl ast.DeclarationNode, lines map[int]struct{}) bool {
	if len(decl.Malformed) > 0 {
		return true
	}
	if _, ok := lines[decl.Line]; ok {
		return true
	}

	for _, node := range decl.Body {
		if _, ok := lines[node.Line]; ok {
			return true
		}
	}

	return false
}

func indexReferences(refs []parser.ReferenceOccurrence) []docindex.Reference {
	if len(refs) == 0 {
		return nil
	}

	out := make([]docindex.Reference, len(refs))
	for i, ref := range refs {
		out[i] = docindex.Reference{Kind: ref.Kind, Name: ref.Name, Line: ref.Line, Range: ref.Range}
	}

	return out
}

func parseDiagnostics(parseDiags []parser.Diagnostic) []validation.Diagnostic {
	out := make([]validation.Diagnostic, len(parseDiags))
	for i, d := range parseDiags {
		out[i] = validation.NewDiagnostic(validation.DiagnosticCode(d.Code), d.Line, d.Column)
	}
	return out
}

func enforceStrictPolicy(diagnostics []validation.Diagnostic, options Options) error {
	if !options.Strict {
		return nil
	}

	failing := make([]validation.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == validation.SeverityWarning && !options.StrictConfig.FailOnWarnings {
			continue
		}
		failing = append(failing, diagnostic)
	}
	if len(failing) == 0 {
		return nil
	}

	return &StrictPolicyError{Diagnostics: failing}
}

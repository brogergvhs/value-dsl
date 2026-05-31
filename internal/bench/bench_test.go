package bench

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/formatting"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/lsp"
	"github.com/brogergvhs/value-dsl/internal/lsp/completion"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/parser"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/validation"
	"github.com/brogergvhs/value-dsl/internal/values"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var benchmarkSink any

// DSL generators

func generateDSL(numRequirements int) string {
	stakeholders := []string{"Worker", "Supervisor", "SafetyOfficer", "Admin", "Auditor"}
	actionVerbs := []string{"track", "monitor", "notify", "log", "restrict"}
	objects := []string{"location", "status", "activity", "access", "identity"}

	var b strings.Builder
	for _, s := range stakeholders {
		fmt.Fprintf(&b, "stakeholder %s\n", s)
	}
	b.WriteString("\n")

	numValues := max(5, numRequirements/2)
	for i := range numValues {
		angle := float64(i) * (2 * math.Pi) / float64(numValues)
		radius := 0.5 + 0.5*float64(i%5)/4.0
		fmt.Fprintf(&b, "value v%d = %.4f, %.4f\n", i, angle, radius)
	}
	b.WriteString("\n")

	for i := range numRequirements {
		sh := stakeholders[i%len(stakeholders)]
		verb := actionVerbs[i%len(actionVerbs)]
		object := objects[i%len(objects)]
		target := stakeholders[(i+1)%len(stakeholders)]

		fmt.Fprintf(&b, "requirement R%d\n", i)
		if i%3 == 0 {
			fmt.Fprintf(&b, "when %s enters zone_%d\n", sh, i)
		}
		switch verb {
		case "notify":
			fmt.Fprintf(&b, "system shall %s %s\n", verb, target)
		case "restrict":
			fmt.Fprintf(&b, "system shall %s %s of %s\n", verb, object, target)
		default:
			fmt.Fprintf(&b, "system shall %s %s of %s using sensor_%d\n", verb, object, target, i)
		}
		fmt.Fprintf(&b, "stakeholders %s, %s\n", sh, target)
		if i%4 == 0 {
			fmt.Fprintf(&b, "priority high\n")
		}
		if i%5 == 0 {
			fmt.Fprintf(&b, "retention permanent\n")
		}
		b.WriteString("\n")
	}

	for i := range numRequirements {
		sh := stakeholders[i%len(stakeholders)]
		fmt.Fprintf(&b, "assignment R%d\n", i)
		fmt.Fprintf(&b, "%s -> v%d\n", sh, i%numValues)
		fmt.Fprintf(&b, "%s -> v%d\n\n", stakeholders[(i+1)%len(stakeholders)], (i+1)%numValues)
	}

	return b.String()
}

func generateConcentratedDSL(numAssignments int) string {
	var b strings.Builder
	numStakeholders := max(2, numAssignments)
	numValues := max(2, numAssignments/2)

	for i := range numStakeholders {
		fmt.Fprintf(&b, "stakeholder S%d\n", i)
	}
	b.WriteString("\n")
	for i := range numValues {
		angle := float64(i) * (2 * math.Pi) / float64(numValues)
		fmt.Fprintf(&b, "value v%d = %.4f, 0.80\n", i, angle)
	}
	b.WriteString("\n")
	b.WriteString("requirement HOTSPOT\n")
	b.WriteString("system shall monitor status of S0\n")
	b.WriteString("stakeholders S0, S1\n\n")
	b.WriteString("assignment HOTSPOT\n")
	for i := range numAssignments {
		fmt.Fprintf(&b, "S%d -> v%d\n", i%numStakeholders, i%numValues)
	}
	b.WriteString("\n")
	return b.String()
}

func generateBalancedDSL(numRequirements, assignmentsPerReq int) string {
	var b strings.Builder
	numStakeholders := max(2, assignmentsPerReq)
	numValues := max(2, assignmentsPerReq)

	for i := range numStakeholders {
		fmt.Fprintf(&b, "stakeholder S%d\n", i)
	}
	b.WriteString("\n")
	for i := range numValues {
		angle := float64(i) * (2 * math.Pi) / float64(numValues)
		fmt.Fprintf(&b, "value v%d = %.4f, 0.80\n", i, angle)
	}
	b.WriteString("\n")

	for r := range numRequirements {
		fmt.Fprintf(&b, "requirement R%d\n", r)
		fmt.Fprintf(&b, "system shall monitor status of S0\n")
		fmt.Fprintf(&b, "stakeholders S0, S1\n\n")
		fmt.Fprintf(&b, "assignment R%d\n", r)
		for a := range assignmentsPerReq {
			fmt.Fprintf(&b, "S%d -> v%d\n", a%numStakeholders, a%numValues)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func generateInvalidDSL(numRequirements int) string {
	var b strings.Builder
	for i := range max(2, numRequirements/10) {
		fmt.Fprintf(&b, "stakeholder S%d\n", i)
	}
	b.WriteString("value v = 1.0\n\n")
	for i := range numRequirements {
		fmt.Fprintf(&b, "requirement R%d extra\n", i)
		b.WriteString("unknown clause payload\n")
		switch i % 4 {
		case 0:
			b.WriteString("when S0 enters\n")
		case 1:
			b.WriteString("system shall notify\n")
		case 2:
			b.WriteString("priority urgent extra\n")
		default:
			b.WriteString("linked_to target extra\n")
		}
		b.WriteString("stakeholders Unknown\n")
		fmt.Fprintf(&b, "assignment R%d extra\nS0 -> unknown_value\n\n", i)
	}
	return b.String()
}

func generateEditLocalityDSL(numRequirements, seed int) string {
	text := generateDSL(numRequirements)
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}
	index := len(lines) - 1
	switch seed % 3 {
	case 0:
		index = min(1, len(lines)-1)
	case 1:
		index = len(lines) / 2
	}
	lines[index] += " // edit"
	return strings.Join(lines, "\n")
}

func generateXrefWorkload(decls, refsN int) (*semantic.SemanticModel, []docindex.Reference) {
	semModel := &semantic.SemanticModel{
		Stakeholders: make([]model.Stakeholder, 0, decls),
		Values:       make([]model.Value, 0, decls),
		Requirements: make([]model.Requirement, 0, decls),
	}
	for i := range decls {
		semModel.Stakeholders = append(semModel.Stakeholders, model.Stakeholder{
			Name: fmt.Sprintf("S%d", i), Line: i + 1,
			Range: benchSpan(i+1, 13, fmt.Sprintf("S%d", i)),
		})
		semModel.Values = append(semModel.Values, model.Value{
			Name: fmt.Sprintf("v%d", i), Line: decls + i + 1,
			Range:  benchSpan(decls+i+1, 7, fmt.Sprintf("v%d", i)),
			Angle:  float64(i) * (2 * math.Pi) / float64(max(1, decls)),
			Radius: 0.8, HasAngle: true, HasRadius: true,
		})
		semModel.Requirements = append(semModel.Requirements, model.Requirement{
			ID: fmt.Sprintf("R%d", i), Line: decls*2 + i + 1,
			Range: benchSpan(decls*2+i+1, 13, fmt.Sprintf("R%d", i)),
		})
	}

	refs := make([]docindex.Reference, 0, refsN)
	kinds := []grammar.DeclarationKind{
		grammar.DeclarationKindStakeholder, grammar.DeclarationKindValue, grammar.DeclarationKindRequirement,
	}
	for i := range refsN {
		kind := kinds[i%len(kinds)]
		idx := i % max(1, decls)
		name := fmt.Sprintf("S%d", idx)
		switch kind {
		case grammar.DeclarationKindValue:
			name = fmt.Sprintf("v%d", idx)
		case grammar.DeclarationKindRequirement:
			name = fmt.Sprintf("R%d", idx)
		}
		line := decls*3 + i + 1
		refs = append(refs, docindex.Reference{
			Kind: kind, Name: name, Line: line,
			Range: benchSpan(line, 5, name),
		})
	}
	return semModel, refs
}

func benchSpan(line, startColumn int, text string) sourcepos.Range {
	return sourcepos.NewRange(
		sourcepos.NewPosition(line, startColumn),
		sourcepos.NewPosition(line, startColumn+len(text)),
	)
}

// Core benchmarks

func BenchmarkParse(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 10000} {
		text := generateDSL(n)
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				benchmarkSink = parser.Parse(text)
			}
		})
	}
}

func BenchmarkSemanticBuild(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 10000} {
		parsed := parser.Parse(generateDSL(n))
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				model, err := semantic.Build(parsed.AST)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = model
			}
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 10000} {
		model := mustModel(b, generateDSL(n))
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSink = validation.Validate(model)
			}
		})
	}
}

func BenchmarkXref(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 10000} {
		text := generateDSL(n)
		parsed := parser.Parse(text)
		model := mustModel(b, text)
		refs := make([]docindex.Reference, 0, len(parsed.References))
		for _, r := range parsed.References {
			refs = append(refs, docindex.Reference{Kind: r.Kind, Name: r.Name, Line: r.Line, Range: r.Range})
		}
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSink = buildIndex(model, refs)
			}
		})
	}
}

func BenchmarkFormat(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 10000} {
		text := generateDSL(n)
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				formatted, err := formatting.Format(text)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = formatted
			}
		})
	}
}

func BenchmarkConflict_Balanced(b *testing.B) {
	for _, n := range []int{10, 50, 200} {
		model := mustModel(b, generateBalancedDSL(n, 10))
		b.Run(fmt.Sprintf("reqs=%d_assign=10each", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSink = coreanalysis.AnalyzeConflicts(model)
			}
		})
	}
}

func BenchmarkConflict_Concentrated(b *testing.B) {
	for _, n := range []int{10, 50, 200, 500} {
		model := mustModel(b, generateConcentratedDSL(n))
		b.Run(fmt.Sprintf("assignments=%d_single_req", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSink = coreanalysis.AnalyzeConflicts(model)
			}
		})
	}
}

func BenchmarkConflict_ConcentratedLarge(b *testing.B) {
	for _, n := range []int{1000, 2000} {
		model := mustModel(b, generateConcentratedDSL(n))
		b.Run(fmt.Sprintf("assignments=%d_single_req", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSink = coreanalysis.AnalyzeConflicts(model)
			}
		})
	}
}

func BenchmarkConflictScore(b *testing.B) {
	a := values.Definition{Angle: 0.40, Radius: 0.86}
	b.Run("opposing", func(b *testing.B) {
		c := values.Definition{Angle: 3.14, Radius: 0.88}
		b.ReportAllocs()
		var total float64
		for b.Loop() {
			total += coreanalysis.ConflictScore(a, c)
		}
		benchmarkSink = total
	})
	b.Run("adjacent", func(b *testing.B) {
		near := values.Definition{Angle: 0.48, Radius: 0.84}
		b.ReportAllocs()
		var total float64
		for b.Loop() {
			total += coreanalysis.ConflictScore(a, near)
		}
		benchmarkSink = total
	})
}

func BenchmarkAnalysisPipeline(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		text := generateDSL(n)
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				result, err := coreanalysis.Run(text, coreanalysis.Options{BuildConflicts: true})
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = result
			}
		})
	}
}

// Invalid-input benchmarks

func BenchmarkParseInvalid(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 2000} {
		text := generateInvalidDSL(n)
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				benchmarkSink = parser.Parse(text)
			}
		})
	}
}

func BenchmarkValidateInvalid(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 2000} {
		model := mustModel(b, generateInvalidDSL(n))
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkSink = validation.Validate(model)
			}
		})
	}
}

func BenchmarkFormatInvalid(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 2000} {
		text := generateInvalidDSL(n)
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				formatted, err := formatting.Format(text)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = formatted
			}
		})
	}
}

func BenchmarkAnalysisPipelineInvalid(b *testing.B) {
	for _, n := range []int{10, 100, 1000, 2000} {
		text := generateInvalidDSL(n)
		b.Run(fmt.Sprintf("reqs=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				result, err := coreanalysis.Run(text, coreanalysis.Options{BuildConflicts: true})
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = result
			}
		})
	}
}

// Xref benchmarks

func BenchmarkXrefBuildDocumentMatrix(b *testing.B) {
	for _, decls := range []int{100, 1000, 5000} {
		for _, refsN := range []int{100, 1000, 5000} {
			model, refs := generateXrefWorkload(decls, refsN)
			b.Run(fmt.Sprintf("decls=%d_refs=%d", decls, refsN), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					benchmarkSink = buildIndex(model, refs)
				}
			})
		}
	}
}

func BenchmarkXrefLookupDeclaration(b *testing.B) {
	model, refs := generateXrefWorkload(5000, 1000)
	graph := buildIndex(model, refs)
	names := []string{"S0", "S499", "S2499", "v0", "v499", "R2499"}

	b.ReportAllocs()
	var hits int
	for b.Loop() {
		for _, name := range names {
			if _, ok := graph.LookupDeclaration(xrefKindForName(name), name); ok {
				hits++
			}
		}
	}
	b.ReportMetric(float64(hits)/float64(b.N), "hits/op")
}

func BenchmarkXrefReferencesFor(b *testing.B) {
	model, refs := generateXrefWorkload(5000, 10000)
	graph := buildIndex(model, refs)
	names := []string{"S0", "S17", "S499", "v0", "v17", "R499"}

	b.ReportAllocs()
	for b.Loop() {
		for _, name := range names {
			benchmarkSink = graph.ReferencesFor(xrefKindForName(name), name)
		}
	}
}

func BenchmarkXrefOccurrenceAt(b *testing.B) {
	model, refs := generateXrefWorkload(5000, 10000)
	graph := buildIndex(model, refs)
	positions := []sourcepos.Position{
		sourcepos.NewPosition(1, 14), sourcepos.NewPosition(500, 14),
		sourcepos.NewPosition(5001, 7), sourcepos.NewPosition(10001, 14),
		sourcepos.NewPosition(13000, 11),
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, pos := range positions {
			benchmarkSink, _ = graph.OccurrenceAt(pos)
		}
	}
}

// LSP benchmarks

func BenchmarkLSPCompletionWarm(b *testing.B) {
	state := newLSPBenchState(b, editorFixtureText())
	params := protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
			Position:     protocol.Position{Line: 3, Character: 20},
		},
	}
	if _, err := state.handler.TextDocumentCompletion(state.context, &params); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		value, err := state.handler.TextDocumentCompletion(state.context, &params)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = value
	}
}

func BenchmarkLSPHover(b *testing.B) {
	state := newLSPBenchState(b, editorFixtureText())
	params := protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
			Position:     protocol.Position{Line: 3, Character: 21},
		},
	}

	b.ReportAllocs()
	for b.Loop() {
		hover, err := state.handler.TextDocumentHover(state.context, &params)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = hover
	}
}

func BenchmarkLSPDefinition(b *testing.B) {
	state := newLSPBenchState(b, editorFixtureText())
	params := protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
			Position:     protocol.Position{Line: 3, Character: 21},
		},
	}

	b.ReportAllocs()
	for b.Loop() {
		value, err := state.handler.TextDocumentDefinition(state.context, &params)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = value
	}
}

func BenchmarkLSPReferences(b *testing.B) {
	state := newLSPBenchState(b, editorFixtureText())
	params := protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
			Position:     protocol.Position{Line: 3, Character: 21},
		},
		Context: protocol.ReferenceContext{IncludeDeclaration: true},
	}

	b.ReportAllocs()
	for b.Loop() {
		locations, err := state.handler.TextDocumentReferences(state.context, &params)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = locations
	}
}

func BenchmarkLSPRename(b *testing.B) {
	state := newLSPBenchState(b, editorFixtureText())
	params := protocol.RenameParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
			Position:     protocol.Position{Line: 3, Character: 21},
		},
		NewName: "Employee",
	}

	b.ReportAllocs()
	for b.Loop() {
		edit, err := state.handler.TextDocumentRename(state.context, &params)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = edit
	}
}

func BenchmarkLSPSemanticTokens(b *testing.B) {
	state := newLSPBenchState(b, generateDSL(1000))
	params := protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
	}

	b.ReportAllocs()
	for b.Loop() {
		tokens, err := state.handler.TextDocumentSemanticTokensFull(state.context, &params)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = tokens
	}
}

func BenchmarkLSPDidOpenToDiagnostics(b *testing.B) {
	text := generateDSL(1000)
	server := lsp.NewServer()
	handler := server.Handler()
	context := &glsp.Context{
		Notify: func(string, any) {},
	}
	b.Cleanup(func() {
		_ = handler.Shutdown(context)
	})

	b.ReportAllocs()
	b.SetBytes(int64(len(text)))
	var version int32
	for b.Loop() {
		version++
		uri := protocol.DocumentUri(fmt.Sprintf("file:///open_%d.dsl", version))
		if err := handler.TextDocumentDidOpen(context, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{URI: uri, Version: version, Text: text},
		}); err != nil {
			b.Fatal(err)
		}
		_ = handler.TextDocumentDidClose(context, &protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		})
	}
}

func BenchmarkLSPWorkspaceDidOpenToDiagnostics(b *testing.B) {
	workspace := newLSPWorkspaceFixture(b, 500)

	b.ReportAllocs()
	b.SetBytes(int64(workspace.bytes))
	var version int32
	for b.Loop() {
		server := lsp.NewServer()
		handler := server.Handler()
		context := &glsp.Context{Notify: func(string, any) {}}
		version++
		if err := handler.TextDocumentDidOpen(context, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{URI: workspace.uri, Version: version, Text: workspace.openText},
		}); err != nil {
			b.Fatal(err)
		}
		_ = handler.Shutdown(context)
	}
}

func BenchmarkLSPDidChangeToDiagnostics(b *testing.B) {
	initialText := generateDSL(200)
	changedText := generateEditLocalityDSL(200, 1)
	server := lsp.NewServer()
	handler := server.Handler()
	uri := protocol.DocumentUri("file:///change.dsl")
	publishedVersions := make(chan int32, 1024)
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			if published, ok := params.(protocol.PublishDiagnosticsParams); ok && published.Version != nil && *published.Version > 1 {
				select {
				case publishedVersions <- int32(*published.Version):
				default:
				}
			}
		},
	}
	b.Cleanup(func() {
		_ = handler.Shutdown(context)
	})

	if err := handler.TextDocumentDidOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: uri, Version: 1, Text: initialText},
	}); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(changedText)))
	var version int32 = 1
	for b.Loop() {
		version++
		if err := handler.TextDocumentDidChange(context, &protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: uri},
				Version:                version,
			},
			ContentChanges: []any{
				protocol.TextDocumentContentChangeEventWhole{Text: changedText},
			},
		}); err != nil {
			b.Fatal(err)
		}
		waitForDiagnosticVersion(b, publishedVersions, version)
	}
}

func BenchmarkLSPWorkspaceWarmRequests(b *testing.B) {
	requests := []struct {
		name string
		run  func(lspWorkspaceBenchState) (any, error)
	}{
		{"completion", func(state lspWorkspaceBenchState) (any, error) {
			return state.handler.TextDocumentCompletion(state.context, &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
					Position:     state.completionPosition,
				},
			})
		}},
		{"hover", func(state lspWorkspaceBenchState) (any, error) {
			return state.handler.TextDocumentHover(state.context, &protocol.HoverParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
					Position:     state.workerPosition,
				},
			})
		}},
		{"definition", func(state lspWorkspaceBenchState) (any, error) {
			return state.handler.TextDocumentDefinition(state.context, &protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
					Position:     state.workerPosition,
				},
			})
		}},
		{"references", func(state lspWorkspaceBenchState) (any, error) {
			return state.handler.TextDocumentReferences(state.context, &protocol.ReferenceParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: state.uri},
					Position:     state.workerPosition,
				},
				Context: protocol.ReferenceContext{IncludeDeclaration: true},
			})
		}},
	}

	for _, request := range requests {
		b.Run(request.name, func(b *testing.B) {
			state := newLSPWorkspaceBenchState(b, 500)
			b.ReportAllocs()
			for b.Loop() {
				value, err := request.run(state)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = value
			}
		})
	}
}

// Parallel benchmarks

func BenchmarkDocumentStoreParallel(b *testing.B) {
	store := lsp.NewStore()

	b.ReportAllocs()
	b.SetParallelism(2)
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			uri := fmt.Sprintf("file:///bench_%d.dsl", i%8)
			store.Set(uri, int32(i), "stakeholder W\n")
			store.Get(uri)
			if i%16 == 0 {
				store.Delete(uri)
			}
			i++
		}
	})
}

func BenchmarkAnalysisCacheParallel(b *testing.B) {
	cache := lsp.NewCache()
	result := coreanalysis.Result{Index: docindex.Build(nil, nil, &semantic.SemanticModel{}, nil, nil, nil, nil)}

	b.ReportAllocs()
	b.SetParallelism(2)
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			uri := fmt.Sprintf("file:///bench_%d.dsl", i%8)
			version := int32(i)
			document := lsp.Document{URI: uri, Version: version}
			cache.PutDocument(document, result)
			cache.GetDocument(document)
			cache.GetNavigationFallback(uri)
			if i%16 == 0 {
				cache.Delete(uri)
			}
			i++
		}
	})
}

func BenchmarkCompletionParallel(b *testing.B) {
	text, result, pos := completionWorkload(b, 1000)
	var firstErr atomic.Value

	b.ReportAllocs()
	b.SetParallelism(2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if items := completion.Items(text, pos, result); len(items) == 0 {
				recordParallelError(&firstErr, "expected completion items")
				return
			}
		}
	})
	failOnParallelError(b, &firstErr)
}

func BenchmarkLSPDidChangeParallel(b *testing.B) {
	state := newLSPBenchState(b, generateDSL(100))
	uris := make([]protocol.DocumentUri, 8)
	for i := range uris {
		uris[i] = protocol.DocumentUri(fmt.Sprintf("file:///bench_%d.dsl", i))
		if err := state.handler.TextDocumentDidOpen(state.context, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{URI: uris[i], Version: int32(i + 1), Text: generateDSL(20)},
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.SetParallelism(2)
	var firstErr atomic.Value
	b.RunParallel(func(pb *testing.PB) {
		var i int
		for pb.Next() {
			uri := uris[i%len(uris)]
			err := state.handler.TextDocumentDidChange(state.context, &protocol.DidChangeTextDocumentParams{
				TextDocument: protocol.VersionedTextDocumentIdentifier{
					TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: uri},
					Version:                int32(i + 1000),
				},
				ContentChanges: []any{
					protocol.TextDocumentContentChangeEventWhole{Text: generateEditLocalityDSL(20, i)},
				},
			})
			if err != nil {
				recordParallelError(&firstErr, "didChange(%s) error: %v", uri, err)
				return
			}
			i++
		}
	})
	failOnParallelError(b, &firstErr)
}

func BenchmarkAnalysisRunParallel(b *testing.B) {
	text := generateDSL(100)
	var firstErr atomic.Value

	b.ReportAllocs()
	b.SetParallelism(2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := coreanalysis.Run(text, coreanalysis.Options{BuildConflicts: true})
			if err != nil {
				recordParallelError(&firstErr, "analysis error: %v", err)
				return
			}
			if result.Index == nil || result.Index.Model == nil {
				recordParallelError(&firstErr, "expected analysis model")
				return
			}
		}
	})
	failOnParallelError(b, &firstErr)
}

// Misc benchmarks

func BenchmarkGrammarExportJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		data, err := grammar.ExportJSON()
		if err != nil {
			b.Fatal(err)
		}
		benchmarkSink = data
	}
}

func BenchmarkEditLocalityAnalyze(b *testing.B) {
	for _, location := range []string{"top", "middle", "end"} {
		seed := 2
		switch location {
		case "top":
			seed = 0
		case "middle":
			seed = 1
		}
		text := generateEditLocalityDSL(1000, seed)
		b.Run(location, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(text)))
			for b.Loop() {
				result, err := coreanalysis.Run(text, coreanalysis.Options{})
				if err != nil {
					b.Fatal(err)
				}
				benchmarkSink = result
			}
		})
	}
}

func BenchmarkCompletion(b *testing.B) {
	text, result, pos := completionWorkload(b, 1000)

	b.ReportAllocs()
	for b.Loop() {
		benchmarkSink = completion.Items(text, pos, result)
	}
}

// Allocation budgets

func TestAllocationBudgets(t *testing.T) {
	smallText := "stakeholder Worker\nvalue privacy_pref = 1.58, 0.91\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\n"
	parsed := parser.Parse(smallText)
	model := mustModel(t, smallText)
	refs := parserRefsToIndexRefs(parsed.References)
	graph := buildIndex(model, refs)

	cases := []struct {
		name   string
		budget float64
		fn     func()
	}{
		{"SemanticBuildSmallAST", 120, func() {
			m, err := semantic.Build(parsed.AST)
			if err != nil {
				t.Fatal(err)
			}
			benchmarkSink = m
		}},
		{"XrefLookupDeclaration", 1, func() {
			benchmarkSink, _ = graph.LookupDeclaration(grammar.DeclarationKindStakeholder, "Worker")
		}},
		{"FormatSmallDocument", 500, func() {
			formatted, err := formatting.Format(smallText)
			if err != nil {
				t.Fatal(err)
			}
			benchmarkSink = formatted
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := testing.AllocsPerRun(100, tc.fn)
			if got > tc.budget {
				t.Fatalf("allocations = %.2f, budget %.2f", got, tc.budget)
			}
		})
	}
}

// Formatter correctness

func TestFormatIdempotence(t *testing.T) {
	for _, n := range []int{10, 100, 500} {
		t.Run(fmt.Sprintf("reqs=%d", n), func(t *testing.T) {
			text := generateDSL(n)
			first, err := formatting.Format(text)
			if err != nil {
				t.Fatal(err)
			}
			second, err := formatting.Format(first)
			if err != nil {
				t.Fatal(err)
			}
			if first != second {
				t.Fatal("format is not idempotent")
			}
		})
	}
}

func TestFormatSemanticPreservation(t *testing.T) {
	for _, n := range []int{10, 100, 500} {
		t.Run(fmt.Sprintf("reqs=%d", n), func(t *testing.T) {
			text := generateDSL(n)
			formatted, err := formatting.Format(text)
			if err != nil {
				t.Fatal(err)
			}
			assertSameModelCounts(t, text, formatted)
		})
	}
}

func TestFormatInvalidDocNoSemanticLoss(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"missing_angle", "stakeholder W\nvalue x\nrequirement R1\nsystem shall notify W\nstakeholders W\nassignment R1\nW -> x\n"},
		{"partial_value", "stakeholder W\nvalue x = 1.5\nrequirement R1\nsystem shall notify W\nstakeholders W\n"},
		{"malformed_when", "stakeholder W\nrequirement R1\nwhen W enters\nsystem shall notify W\nstakeholders W\n"},
		{"truncated_action", "stakeholder W\nrequirement R1\nsystem shall notify\nstakeholders W\n"},
		{"truncated_assignment", "stakeholder W\nvalue v = 1.0, 0.5\nrequirement R1\nsystem shall notify W\nstakeholders W\nassignment R1\nW ->\n"},
		{"unknown_ref", "stakeholder W\nrequirement R1\nsystem shall notify W\nstakeholders W\nassignment R1\nW -> unknown_value\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := formatting.Format(tc.text)
			if err != nil {
				t.Fatal(err)
			}
			assertSameModelCounts(t, tc.text, formatted)
		})
	}
}

// Fuzzing

func FuzzParseLines(f *testing.F) {
	f.Add("stakeholder Worker")
	f.Add(`value v = 1.58, 0.91`)
	f.Add(`system shall track "user location" of Worker using "GPS"`)
	f.Add(`Worker, Supervisor -> privacy, authority`)
	f.Add(`,,,->->===`)
	f.Add(`"unterminated`)
	f.Fuzz(func(t *testing.T, line string) {
		_ = parser.Parse(line)
	})
}

func FuzzCommentSplitting(f *testing.F) {
	f.Add("code // comment")
	f.Add("// only comment")
	f.Add("no comment here")
	f.Add("")
	f.Add(`"string with // inside" // real comment`)
	f.Add("a//b")
	f.Fuzz(func(t *testing.T, line string) {
		tl := grammar.TokenizeRawLine(1, line)
		if tl.CommentOffset >= 0 && !strings.Contains(line, "//") {
			t.Fatal("reported comment without //")
		}
	})
}

func FuzzParseFull(f *testing.F) {
	f.Add("stakeholder W\nvalue v = 1.0, 0.5\nrequirement R\nsystem shall notify W\nstakeholders W\n")
	f.Add("")
	f.Add("requirement\n")
	f.Add("value = ,\n")
	f.Add("assignment\n-> -> ->\n")
	f.Add("// only comments\n")
	f.Add(strings.Repeat("stakeholder S\n", 1000))
	f.Fuzz(func(t *testing.T, text string) {
		result := parser.Parse(text)
		if result.AST == nil {
			t.Fatal("parser returned nil AST")
		}
		if _, err := semantic.Build(result.AST); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzFormat(f *testing.F) {
	f.Add("stakeholder W\nvalue v = 1.0, 0.5\nrequirement R\nsystem shall notify W\nstakeholders W\n")
	f.Add("value x\n")
	f.Add("assignment R\nW -> v\n")
	f.Add("")
	f.Fuzz(func(t *testing.T, text string) {
		if _, err := formatting.Format(text); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzMalformedClausesAndAssignments(f *testing.F) {
	f.Add("when W enters", "W ->")
	f.Add("while W", "-> v")
	f.Add("if W has_no_consent extra", "W, Supervisor -> v,")
	f.Add("system shall", ",,,->")
	f.Add("priority urgent extra", "W -> unknown")
	f.Fuzz(func(t *testing.T, clause, assignment string) {
		if len(clause) > 4096 || len(assignment) > 4096 {
			t.Skip()
		}
		text := "stakeholder W\nstakeholder Supervisor\nvalue v = 1.0, 0.5\nrequirement R1\n" +
			clause + "\nsystem shall notify W\nstakeholders W\nassignment R1\n" + assignment + "\n"
		result := parser.Parse(text)
		if result.AST == nil {
			t.Fatal("parser returned nil AST")
		}
		if _, err := semantic.Build(result.AST); err != nil {
			t.Fatal(err)
		}
		if _, err := formatting.Format(text); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzXrefQueries(f *testing.F) {
	f.Add("S1", 1, 14)
	f.Add("v2", 103, 8)
	f.Add("R3", 205, 14)
	f.Add("missing", 999, 99)
	model, refs := generateXrefWorkload(100, 500)
	graph := buildIndex(model, refs)

	f.Fuzz(func(t *testing.T, name string, line int, column int) {
		if len(name) > 256 {
			t.Skip()
		}
		kind := xrefKindForName(name)
		_, _ = graph.LookupDeclaration(kind, name)
		_ = graph.ReferencesFor(kind, name)
		_, _ = graph.OccurrenceAt(sourcepos.NewPosition(line, column))
	})
}

func FuzzEditorRoundTrip(f *testing.F) {
	f.Add("stakeholder W\nvalue v = 1.0, 0.5\nrequirement R1\nsystem shall notify W\nstakeholders W\n")
	f.Add("stakeholder W\nrequirement R1\nsystem shall track \"unterminated of W using GPS\nstakeholders W\n")
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 16384 {
			t.Skip()
		}
		result, err := coreanalysis.Run(text, coreanalysis.Options{BuildConflicts: true})
		if err != nil && (result.Index == nil || result.Index.AST == nil) {
			t.Fatal(err)
		}
		formatted, err := formatting.Format(text)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := coreanalysis.Run(formatted, coreanalysis.Options{}); err != nil {
			t.Fatal(err)
		}
	})
}

// Race tests

func TestDocumentStore_Race(t *testing.T) {
	store := lsp.NewStore()
	var wg sync.WaitGroup
	wg.Add(20)
	for g := range 20 {
		go func(id int) {
			defer wg.Done()
			uri := fmt.Sprintf("file:///test_%d.dsl", id%5)
			for i := range 500 {
				store.Set(uri, int32(i), fmt.Sprintf("stakeholder S%d\n", i))
				store.Get(uri)
				if i%10 == 0 {
					store.Delete(uri)
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestAnalysisCache_Race(t *testing.T) {
	cache := lsp.NewCache()
	var wg sync.WaitGroup
	wg.Add(20)
	for g := range 20 {
		go func(id int) {
			defer wg.Done()
			uri := fmt.Sprintf("file:///test_%d.dsl", id%5)
			result := coreanalysis.Result{Index: docindex.Build(nil, nil, &semantic.SemanticModel{}, nil, nil, nil, nil)}
			for i := range 500 {
				document := lsp.Document{URI: uri, Version: int32(i)}
				cache.PutDocument(document, result)
				cache.GetDocument(document)
				cache.GetNavigationFallback(uri)
				if i%10 == 0 {
					cache.Delete(uri)
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestLSPTimerScheduling_Race(t *testing.T) {
	server := lsp.NewServer()
	handler := server.Handler()
	var notifications atomic.Int64
	context := &glsp.Context{
		Notify: func(_ string, _ any) { notifications.Add(1) },
	}

	uris := []protocol.DocumentUri{
		"file:///race_0.dsl", "file:///race_1.dsl",
		"file:///race_2.dsl", "file:///race_3.dsl",
	}
	for i, uri := range uris {
		if err := handler.TextDocumentDidOpen(context, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{URI: uri, Version: int32(i + 1), Text: fmt.Sprintf("stakeholder S%d\n", i)},
		}); err != nil {
			t.Fatal(err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(16)
	for g := range 16 {
		go func(id int) {
			defer wg.Done()
			for i := range 40 {
				uri := uris[(id+i)%len(uris)]
				version := int32(id*40 + i + 100)
				_ = handler.TextDocumentDidChange(context, &protocol.DidChangeTextDocumentParams{
					TextDocument: protocol.VersionedTextDocumentIdentifier{
						TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: uri},
						Version:                version,
					},
					ContentChanges: []any{
						protocol.TextDocumentContentChangeEventWhole{
							Text: fmt.Sprintf("stakeholder S%d\nrequirement R%d\nsystem shall notify S%d\nstakeholders S%d\n", id, i, id, id),
						},
					},
				})
			}
		}(g)
	}
	wg.Wait()

	time.Sleep(300 * time.Millisecond)
	for _, uri := range uris {
		_ = handler.TextDocumentDidClose(context, &protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
		})
	}
	if err := handler.Shutdown(context); err != nil {
		t.Fatal(err)
	}
	if notifications.Load() == 0 {
		t.Fatal("expected diagnostics notifications")
	}
}

// Property tests

func TestGrammarExportParity(t *testing.T) {
	exported := grammar.Export()

	if exported.IdentifierPattern != grammar.Compiled.Spec.Identifier {
		t.Fatal("identifier pattern mismatch")
	}
	if exported.NumberPattern != grammar.Compiled.Spec.NumberPattern {
		t.Fatal("number pattern mismatch")
	}
	if exported.StringPattern != grammar.Compiled.Spec.StringPattern {
		t.Fatal("string pattern mismatch")
	}
	if exported.LineCommentPrefix != grammar.Compiled.Spec.LineComment {
		t.Fatal("line comment prefix mismatch")
	}
	if len(exported.Declarations) != len(grammar.Compiled.Spec.Declarations) {
		t.Fatalf("declaration count: %d vs %d", len(exported.Declarations), len(grammar.Compiled.Spec.Declarations))
	}
	for i, decl := range grammar.Compiled.Spec.Declarations {
		exp := exported.Declarations[i]
		if exp.Kind != string(decl.Kind) || exp.Keyword != decl.Keyword {
			t.Fatalf("declaration %d mismatch", i)
		}
		if len(exp.Header) != len(decl.Header) {
			t.Fatalf("declaration %q header count mismatch", decl.Keyword)
		}
		if (decl.Body == nil) != (exp.Body == nil) {
			t.Fatalf("declaration %q body presence mismatch", decl.Keyword)
		}
	}
	if len(exported.ReservedKeywords) != len(grammar.Compiled.ReservedKeywords) {
		t.Fatal("reserved keyword count mismatch")
	}
	if len(exported.PunctuationTokens) != len(grammar.Compiled.PunctuationTokens) {
		t.Fatal("punctuation token count mismatch")
	}
}

func TestGrammarExportJSON(t *testing.T) {
	data, err := grammar.ExportJSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("empty JSON export")
	}
	for _, field := range []string{`"identifier_pattern"`, `"string_pattern"`, `"declarations"`} {
		if !strings.Contains(string(data), field) {
			t.Fatalf("JSON missing %s", field)
		}
	}
}

func TestTreeSitterGeneratedArtifactsDrift(t *testing.T) {
	spec := grammar.Export()
	fragments := readRepoFile(t, "tools/tree-sitter-value-dsl/generated/grammar-fragments.js")
	highlights := readRepoFile(t, "tools/tree-sitter-value-dsl/queries/highlights.scm")
	locals := readRepoFile(t, "tools/tree-sitter-value-dsl/queries/locals.scm")
	folds := readRepoFile(t, "tools/tree-sitter-value-dsl/queries/folds.scm")
	commentstring := readRepoFile(t, "tools/tree-sitter-value-dsl/queries/commentstring.scm")

	for label, source := range map[string]string{
		"grammar-fragments.js": fragments, "highlights.scm": highlights,
		"locals.scm": locals, "folds.scm": folds, "commentstring.scm": commentstring,
	} {
		checkIncludes(t, source, "AUTO-GENERATED from generated/grammar.json", "%s not marked as generated", label)
	}

	checkIncludes(t, fragments, fmt.Sprintf(`"identifierPattern": %q`, spec.IdentifierPattern), "identifier pattern drifted")
	checkIncludes(t, fragments, fmt.Sprintf(`"numberPattern": %q`, spec.NumberPattern), "number pattern drifted")
	checkIncludes(t, fragments, fmt.Sprintf(`"stringPattern": %q`, spec.StringPattern), "string pattern drifted")
	checkIncludes(t, fragments, fmt.Sprintf(`"lineCommentPrefix": %q`, spec.LineCommentPrefix), "comment prefix drifted")
	checkIncludes(t, commentstring, fmt.Sprintf(`commentstring "%s %%s"`, spec.LineCommentPrefix), "commentstring drifted")

	for _, decl := range spec.Declarations {
		checkIncludes(t, fragments, fmt.Sprintf(`"kind": %q`, decl.Kind), "missing fragment for %s", decl.Kind)
		checkIncludes(t, fragments, fmt.Sprintf(`"node_name": %q`, decl.NodeName), "missing node for %s", decl.NodeName)
		if decl.Body != nil {
			checkIncludes(t, folds, fmt.Sprintf("(%s) @fold", decl.NodeName), "missing fold for %s", decl.NodeName)
		}

		for _, fieldName := range collectTopLevelDefinitions(decl) {
			checkIncludes(t, locals, fmt.Sprintf("(%s %s: (identifier) @local.definition)", decl.NodeName, fieldName), "missing def for %s.%s", decl.NodeName, fieldName)
		}
		for _, fieldName := range collectReferences(decl.Header) {
			checkIncludes(t, locals, fmt.Sprintf("(%s %s: (identifier) @local.reference)", decl.NodeName, fieldName), "missing ref for %s.%s", decl.NodeName, fieldName)
		}
		if decl.Body == nil {
			continue
		}
		for _, line := range decl.Body.Lines {
			if spec.ActionSystem != nil && line.NodeName == spec.ActionSystem.ClauseNodeName {
				continue
			}
			for _, fieldName := range collectReferences(line.Pattern) {
				checkIncludes(t, locals, fmt.Sprintf("(%s %s: (identifier) @local.reference)", line.NodeName, fieldName), "missing ref for %s.%s", line.NodeName, fieldName)
			}
		}
	}

	if spec.ActionSystem != nil {
		for _, form := range spec.ActionSystem.Forms {
			for _, fieldName := range collectReferences(form.Pattern) {
				checkIncludes(t, locals, fmt.Sprintf("(%s (%s (%s %s: (identifier) @local.reference)))", spec.ActionSystem.ClauseNodeName, spec.ActionSystem.ExprNodeName, form.NodeName, fieldName), "missing action ref for %s.%s", form.NodeName, fieldName)
			}
		}
	}

	for _, keyword := range collectKeywordTokens(spec) {
		checkIncludes(t, highlights, fmt.Sprintf("%q", keyword), "missing keyword highlight for %q", keyword)
	}
	for _, pattern := range collectStructuralHighlightPatterns(spec) {
		checkIncludes(t, highlights, pattern, "missing highlight pattern %q", pattern)
	}
}

func TestCompletionLegalityAfterClauseOrder(t *testing.T) {
	clauseSequences := [][]string{
		{},
		{"while Worker is active\n"},
		{"when Worker enters zone\n"},
		{"while Worker is active\n", "when Worker enters zone\n"},
		{"while Worker is active\n", "when Worker enters zone\n", "if Worker has access\n"},
		{"while Worker is active\n", "when Worker enters zone\n", "if Worker has access\n", "where zone enabled\n"},
		{"while Worker is active\n", "when Worker enters zone\n", "system shall notify Worker\n"},
		{"while Worker is active\n", "when Worker enters zone\n", "system shall notify Worker\n", "stakeholders Worker\n"},
		{"while Worker is active\n", "when Worker enters zone\n", "system shall notify Worker\n", "stakeholders Worker\n", "priority high\n"},
	}

	allClauseKeywords := append([]string{"while", "when", "if", "where"}, grammar.Compiled.MetadataKeywords...)
	allClauseKeywords = append(allClauseKeywords, "system shall")

	for _, seq := range clauseSequences {
		text := "stakeholder Worker\n\nrequirement R1\n" + strings.Join(seq, "")
		lineNum := 3 + len(seq)
		pos := protocol.Position{Line: uint32(lineNum), Character: 0}
		items := completion.Items(text, pos, coreanalysis.Result{})

		for _, item := range items {
			if !slices.Contains(allClauseKeywords, item.Label) {
				continue
			}
			seen := make([]grammar.RequirementClauseKind, 0)
			for _, clause := range seq {
				if kind, ok := grammar.Compiled.RequirementClauseKindForLine(strings.TrimSpace(clause)); ok {
					seen = append(seen, kind)
				}
			}
			next := grammar.Compiled.NextRequirementClauseKeywords(seen)
			if item.Label != "system shall" && !slices.Contains(next, item.Label) {
				t.Errorf("completion %q offered after %v but not legal (next=%v)", item.Label, seq, next)
			}
		}
	}
}

// Regression corpus

func TestRegressionInvalidIdentifiers(t *testing.T) {
	cases := []struct {
		name string
		text string
		code validation.DiagnosticCode
	}{
		{"numeric_stakeholder", "stakeholder 123Worker\nrequirement R1\nsystem shall notify 123Worker\nstakeholders 123Worker\n", validation.DiagnosticCode("stakeholder.identifier.invalid")},
		{"numeric_requirement", "stakeholder W\nrequirement 99\nsystem shall notify W\nstakeholders W\n", validation.DiagnosticCode("requirement.identifier.invalid")},
		{"numeric_value", "value 123val = 1.0, 0.5\n", validation.DiagnosticCode("value.identifier.invalid")},
		{"hyphen_start_stakeholder", "stakeholder -Worker\n", validation.DiagnosticCode("stakeholder.identifier.invalid")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { expectDiagnostic(t, tc.text, tc.code) })
	}
}

func TestRegressionInvalidEnumValues(t *testing.T) {
	base := "stakeholder W\nrequirement R1\nsystem shall notify W\nstakeholders W\n"
	cases := []struct {
		name  string
		extra string
		code  validation.DiagnosticCode
	}{
		{"priority_urgent", "priority urgent\n", "requirement.priority.level.invalid"},
		{"priority_extreme", "priority extreme\n", "requirement.priority.level.invalid"},
		{"priority_empty_string", "priority\n", "requirement.priority.level.missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { expectDiagnostic(t, base+tc.extra, tc.code) })
	}
}

func TestRegressionSurplusTokens(t *testing.T) {
	cases := []struct {
		name string
		text string
		code validation.DiagnosticCode
	}{
		{"value_extra", "value v = 1.0, 0.5 extra\n", "value.extra_tokens"},
		{"requirement_header_extra", "stakeholder W\nrequirement R1 extra\nsystem shall notify W\nstakeholders W\n", "requirement.extra_tokens"},
		{"assignment_header_extra", "stakeholder W\nvalue v = 1.0, 0.5\nrequirement R1\nsystem shall notify W\nstakeholders W\nassignment R1 extra\nW -> v\n", "assignment.extra_tokens"},
		{"retention_extra", "stakeholder W\nrequirement R1\nsystem shall notify W\nstakeholders W\nretention short extra\n", "requirement.retention.extra_tokens"},
		{"linked_to_extra", "stakeholder W\nrequirement R1\nsystem shall notify W\nstakeholders W\nlinked_to target extra\n", "requirement.linked_to.extra_tokens"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { expectDiagnostic(t, tc.text, tc.code) })
	}
}

func TestRegressionBuiltinShadowing(t *testing.T) {
	text := "stakeholder W\nvalue privacy = 2.04, 0.92\nrequirement R1\nsystem shall notify W\nstakeholders W\nassignment R1\nW -> privacy\n"
	model := mustModel(t, text)

	if _, ok := model.ValueByName["privacy"]; !ok {
		t.Fatal("user-defined 'privacy' missing")
	}
	if model.ValueByName["privacy"].Builtin {
		t.Fatal("user-defined 'privacy' should not be builtin")
	}
	expectNoDiagnostic(t, text, validation.CodeAssignmentValueUnknown)
	expectDiagnostic(t, text, validation.CodeValueBuiltinConflict)
}

func TestRegressionBuiltinResolution(t *testing.T) {
	text := "stakeholder W\nrequirement R1\nsystem shall notify W\nstakeholders W\nassignment R1\nW -> privacy\n"
	model := mustModel(t, text)

	if len(model.AssignmentEntries) == 0 {
		t.Fatal("expected assignment")
	}
	a := model.AssignmentEntries[0]
	if len(a.Values) != 1 || a.Values[0].Name != "privacy" {
		t.Fatalf("expected builtin privacy reference, got %+v", a.Values)
	}
}

func TestRegressionQuotedStrings(t *testing.T) {
	text := "stakeholder W\nrequirement R1\n" +
		`system shall track "user location" of W using "GPS system"` + "\n" +
		"stakeholders W\n" + `retention "30 days"` + "\n" + `linked_to "SAFETY-001"` + "\n"

	model := mustModel(t, text)
	diags := validation.Validate(model)
	for _, d := range diags {
		if d.Severity == validation.SeverityError {
			t.Fatalf("unexpected error: %s [%s]", d.Message, d.Code)
		}
	}

	req := model.Requirements[0]
	if req.Action.Object != "user location" {
		t.Fatalf("object = %q, want %q", req.Action.Object, "user location")
	}
	if req.Action.Mechanism != "GPS system" {
		t.Fatalf("mechanism = %q, want %q", req.Action.Mechanism, "GPS system")
	}
	if req.Metadata.Retention.Value != "30 days" {
		t.Fatalf("retention = %q, want %q", req.Metadata.Retention.Value, "30 days")
	}
	if len(req.Traceability) == 0 || req.Traceability[0].Value != "SAFETY-001" {
		t.Fatalf("traceability = %v, want SAFETY-001", req.Traceability)
	}

	formatted, err := formatting.Format(text)
	if err != nil {
		t.Fatal(err)
	}
	second, err := formatting.Format(formatted)
	if err != nil {
		t.Fatal(err)
	}
	if formatted != second {
		t.Fatal("quoted string formatting not idempotent")
	}
}

func TestRegressionValueCategoryBoundaryBins(t *testing.T) {
	cases := []struct {
		angle float64
		want  string
	}{
		{0.00, "power"}, {0.20, "achievement"}, {0.63, "achievement"},
		{0.64, "hedonism"}, {1.26, "hedonism"}, {1.88, "stimulation"},
		{5.65, "security"}, {5.66, "power"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("angle=%.2f", tc.angle), func(t *testing.T) {
			cat, ok := values.ClosestCategory(tc.angle)
			if !ok {
				t.Fatal("expected category")
			}
			if cat.Name != tc.want {
				t.Fatalf("ClosestCategory(%.2f) = %q, want %q", tc.angle, cat.Name, tc.want)
			}
		})
	}
}

// Helpers

func mustModel(tb testing.TB, text string) *semantic.SemanticModel {
	tb.Helper()
	parsed := parser.Parse(text)
	model, err := semantic.Build(parsed.AST)
	if err != nil {
		tb.Fatal(err)
	}
	return model
}

func assertSameModelCounts(t *testing.T, original, formatted string) {
	t.Helper()
	orig := mustModel(t, original)
	fmtd := mustModel(t, formatted)
	if len(orig.Stakeholders) != len(fmtd.Stakeholders) {
		t.Fatalf("stakeholder count: %d -> %d", len(orig.Stakeholders), len(fmtd.Stakeholders))
	}
	if len(orig.Values) != len(fmtd.Values) {
		t.Fatalf("value count: %d -> %d", len(orig.Values), len(fmtd.Values))
	}
	if len(orig.Requirements) != len(fmtd.Requirements) {
		t.Fatalf("requirement count: %d -> %d", len(orig.Requirements), len(fmtd.Requirements))
	}
	if len(orig.AssignmentEntries) != len(fmtd.AssignmentEntries) {
		t.Fatalf("assignment count: %d -> %d", len(orig.AssignmentEntries), len(fmtd.AssignmentEntries))
	}
}

func expectDiagnostic(t *testing.T, text string, code validation.DiagnosticCode) {
	t.Helper()
	result, err := coreanalysis.Run(text, coreanalysis.Options{})
	if err != nil {
		t.Fatal(err)
	}
	diags := result.AllDiagnostics()
	for _, d := range diags {
		if d.Code == code {
			return
		}
	}
	codes := make([]validation.DiagnosticCode, len(diags))
	for i, d := range diags {
		codes[i] = d.Code
	}
	t.Fatalf("expected diagnostic %q, got %v", code, codes)
}

func expectNoDiagnostic(t *testing.T, text string, code validation.DiagnosticCode) {
	t.Helper()
	model := mustModel(t, text)
	diags := validation.Validate(model)
	for _, d := range diags {
		if d.Code == code {
			t.Fatalf("unexpected diagnostic %q", code)
		}
	}
}

func readRepoFile(t *testing.T, relativePath string) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
	data, err := os.ReadFile(filepath.Join(dir, relativePath))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func checkIncludes(t *testing.T, source, needle, format string, args ...any) {
	t.Helper()
	if !strings.Contains(source, needle) {
		t.Fatalf(format, args...)
	}
}

func xrefKindForName(name string) grammar.DeclarationKind {
	switch {
	case strings.HasPrefix(name, "v"):
		return grammar.DeclarationKindValue
	case strings.HasPrefix(name, "R"):
		return grammar.DeclarationKindRequirement
	default:
		return grammar.DeclarationKindStakeholder
	}
}

func parserRefsToIndexRefs(refs []parser.ReferenceOccurrence) []docindex.Reference {
	out := make([]docindex.Reference, 0, len(refs))
	for _, ref := range refs {
		out = append(out, docindex.Reference{Kind: ref.Kind, Name: ref.Name, Line: ref.Line, Range: ref.Range})
	}
	return out
}

func buildIndex(model *semantic.SemanticModel, refs []docindex.Reference) *docindex.Document {
	return docindex.Build(nil, nil, model, nil, nil, nil, refs)
}

func completionWorkload(tb testing.TB, n int) (string, coreanalysis.Result, protocol.Position) {
	tb.Helper()
	text := generateDSL(n) + "\nrequirement ACTIVE\nsystem shall notify "
	result, err := coreanalysis.Run(text, coreanalysis.Options{})
	if err != nil {
		tb.Fatal(err)
	}
	return text, result, protocolPositionAtEnd(text)
}

type lspBenchState struct {
	context *glsp.Context
	handler *protocol.Handler
	uri     protocol.DocumentUri
}

type lspWorkspaceFixture struct {
	uri                protocol.DocumentUri
	openText           string
	bytes              int
	workerPosition     protocol.Position
	completionPosition protocol.Position
}

type lspWorkspaceBenchState struct {
	lspBenchState
	workerPosition     protocol.Position
	completionPosition protocol.Position
}

func newLSPBenchState(b *testing.B, text string) lspBenchState {
	b.Helper()
	server := lsp.NewServer()
	handler := server.Handler()
	state := lspBenchState{
		context: &glsp.Context{Notify: func(string, any) {}},
		handler: handler,
		uri:     "file:///bench.dsl",
	}
	if err := handler.TextDocumentDidOpen(state.context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: state.uri, Version: 1, Text: text},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = handler.Shutdown(state.context) })
	return state
}

func newLSPWorkspaceBenchState(b *testing.B, requirements int) lspWorkspaceBenchState {
	b.Helper()
	workspace := newLSPWorkspaceFixture(b, requirements)
	server := lsp.NewServer()
	handler := server.Handler()
	state := lspWorkspaceBenchState{
		lspBenchState: lspBenchState{
			context: &glsp.Context{Notify: func(string, any) {}},
			handler: handler,
			uri:     workspace.uri,
		},
		workerPosition:     workspace.workerPosition,
		completionPosition: workspace.completionPosition,
	}
	if err := handler.TextDocumentDidOpen(state.context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: workspace.uri, Version: 1, Text: workspace.openText},
	}); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = handler.Shutdown(state.context) })
	return state
}

func newLSPWorkspaceFixture(tb testing.TB, requirements int) lspWorkspaceFixture {
	tb.Helper()
	root := tb.TempDir()
	totalBytes := 0
	write := func(name, text string) {
		tb.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			tb.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			tb.Fatalf("WriteFile() error = %v", err)
		}
		totalBytes += len(text)
	}

	write("main.dsl", "requirement ROOT\nsystem shall notify Worker\nstakeholders Worker, Supervisor\n\n")
	write("stakeholders.dsl", "stakeholder Worker\nstakeholder Supervisor\nstakeholder SafetyOfficer\n")
	write("values/preferences.dsl", "value privacy_pref = 1.58, 0.91\nvalue safety_pref = 0.42, 0.88\n")

	var assignments strings.Builder
	const requirementsPerFile = 25
	fileCount := max(1, (requirements+requirementsPerFile-1)/requirementsPerFile)
	for file := range fileCount {
		var features strings.Builder
		start := file * requirementsPerFile
		end := min(requirements, start+requirementsPerFile)
		for i := start; i < end; i++ {
			fmt.Fprintf(&features, "requirement R%d\nsystem shall notify Worker\nstakeholders Worker, Supervisor\n\n", i)
			fmt.Fprintf(&assignments, "assignment R%d\nWorker -> privacy_pref\nSupervisor -> safety_pref\n\n", i)
		}
		write(fmt.Sprintf("features/feature_%03d.dsl", file), features.String())
	}
	write("assignments.dsl", assignments.String())

	openText := "requirement ACTIVE\nsystem shall notify Worker\nstakeholders Worker, Supervisor\n\nassignment ACTIVE\nWorker -> privacy_pref\nSupervisor -> safety_pref\n"
	openPath := filepath.Join(root, "features", "active.dsl")
	write("features/active.dsl", openText)

	return lspWorkspaceFixture{
		uri:                protocol.DocumentUri((&url.URL{Scheme: "file", Path: openPath}).String()),
		openText:           openText,
		bytes:              totalBytes,
		workerPosition:     protocol.Position{Line: 1, Character: 21},
		completionPosition: protocol.Position{Line: 1, Character: 20},
	}
}

func editorFixtureText() string {
	return "stakeholder Worker\nvalue privacy_pref = 1.58, 0.91\nrequirement R1\nsystem shall notify Worker\nstakeholders Worker\nassignment R1\nWorker -> privacy_pref\n"
}

func protocolPositionAtEnd(text string) protocol.Position {
	line := uint32(strings.Count(text, "\n"))
	lastNewline := strings.LastIndex(text, "\n")
	character := len(text)
	if lastNewline >= 0 {
		character = len(text) - lastNewline - 1
	}
	return protocol.Position{Line: line, Character: uint32(character)}
}

func recordParallelError(firstErr *atomic.Value, format string, args ...any) {
	if firstErr.Load() == nil {
		firstErr.Store(fmt.Sprintf(format, args...))
	}
}

func failOnParallelError(b *testing.B, firstErr *atomic.Value) {
	b.Helper()
	if v := firstErr.Load(); v != nil {
		b.Fatal(v.(string))
	}
}

func waitForDiagnosticVersion(b *testing.B, versions <-chan int32, want int32) {
	b.Helper()
	timeout := time.After(2 * time.Second)
	for {
		select {
		case got := <-versions:
			if got == want {
				return
			}
		case <-timeout:
			b.Fatalf("timed out waiting for diagnostics version %d", want)
		}
	}
}

func collectTopLevelDefinitions(decl grammar.ExportedDecl) []string {
	for i, matcher := range decl.Header {
		hasLeadingKeyword := false
		for _, previous := range decl.Header[:i] {
			if previous.Type == "kw" {
				hasLeadingKeyword = true
				break
			}
		}
		if !hasLeadingKeyword {
			continue
		}
		if (matcher.Type == "ident" || matcher.Type == "list") && matcher.Name != "" {
			return []string{matcher.Name}
		}
	}
	return nil
}

func collectReferences(matchers []grammar.ExportedMatcher) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	var walk func([]grammar.ExportedMatcher)
	walk = func(items []grammar.ExportedMatcher) {
		for _, matcher := range items {
			switch matcher.Type {
			case "ref":
				if matcher.Name != "" {
					if _, ok := seen[matcher.Name]; !ok {
						seen[matcher.Name] = struct{}{}
						out = append(out, matcher.Name)
					}
				}
			case "list":
				if matcher.Item != nil {
					walk([]grammar.ExportedMatcher{*matcher.Item})
				}
			case "opt":
				walk(matcher.Inner)
			}
		}
	}
	walk(matchers)
	return out
}

func collectKeywordTokens(spec grammar.ExportedSpec) []string {
	seen := make(map[string]struct{})
	var walk func([]grammar.ExportedMatcher)
	walk = func(matchers []grammar.ExportedMatcher) {
		for _, matcher := range matchers {
			switch matcher.Type {
			case "kw":
				for token := range strings.FieldsSeq(matcher.Text) {
					seen[token] = struct{}{}
				}
			case "list":
				if matcher.Separator != "" {
					seen[matcher.Separator] = struct{}{}
				}
				if matcher.Item != nil {
					walk([]grammar.ExportedMatcher{*matcher.Item})
				}
			case "opt":
				walk(matcher.Inner)
			}
		}
	}
	for _, decl := range spec.Declarations {
		walk(decl.Header)
		if decl.Body != nil {
			for _, line := range decl.Body.Lines {
				walk(line.Pattern)
			}
		}
	}
	out := make([]string, 0, len(seen))
	for token := range seen {
		out = append(out, token)
	}
	return out
}

func collectStructuralHighlightPatterns(spec grammar.ExportedSpec) []string {
	tokenKindRules := make(map[string]grammar.ExportedTokenKindRule, len(spec.TokenKindRules))
	for _, rule := range spec.TokenKindRules {
		tokenKindRules[rule.Kind] = rule
	}

	out := make([]string, 0)
	var walk func([]grammar.ExportedMatcher, string)
	walk = func(matchers []grammar.ExportedMatcher, nodeName string) {
		for _, matcher := range matchers {
			switch matcher.Type {
			case "tok":
				if matcher.Name == "" {
					continue
				}
				rule, ok := tokenKindRules[matcher.Kind]
				if !ok {
					continue
				}
				if len(rule.Values) > 0 {
					out = append(out, fmt.Sprintf("(%s %s: (%s) @function.builtin)", nodeName, matcher.Name, rule.Rule))
				}
				if rule.Fallback != "" {
					out = append(out, fmt.Sprintf("(%s %s: (%s) @function.builtin)", nodeName, matcher.Name, rule.Fallback))
				} else if len(rule.Values) == 0 {
					out = append(out, fmt.Sprintf("(%s %s: (%s) @function.builtin)", nodeName, matcher.Name, rule.Rule))
				}
			case "enum":
				if matcher.Name != "" {
					out = append(out, fmt.Sprintf("(%s %s: (identifier) @constant)", nodeName, matcher.Name))
				}
			case "list":
				if matcher.Item != nil {
					walk([]grammar.ExportedMatcher{*matcher.Item}, nodeName)
				}
			case "opt":
				walk(matcher.Inner, nodeName)
			}
		}
	}

	for _, decl := range spec.Declarations {
		if decl.Body == nil {
			continue
		}
		for _, line := range decl.Body.Lines {
			if spec.ActionSystem != nil && line.NodeName == spec.ActionSystem.ClauseNodeName {
				continue
			}
			walk(line.Pattern, line.NodeName)
		}
	}
	if spec.ActionSystem != nil {
		for _, form := range spec.ActionSystem.Forms {
			walk(form.Pattern, form.NodeName)
		}
	}
	return out
}

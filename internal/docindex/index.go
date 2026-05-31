// Package docindex stores the single indexed view of one analyzed document.
package docindex

import (
	"sort"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/validation"
)

type Declaration struct {
	Kind  grammar.DeclarationKind
	Name  string
	Line  int
	Range sourcepos.Range
}

type Reference struct {
	Kind    grammar.DeclarationKind
	Name    string
	Line    int
	Range   sourcepos.Range
	Context string
}

type Occurrence struct {
	Kind        grammar.DeclarationKind
	Name        string
	Range       sourcepos.Range
	Declaration bool
}

type SourceFile struct {
	Path      string
	StartLine int
	EndLine   int
}

type symbolKey struct {
	Kind grammar.DeclarationKind
	Name string
}

// Document is the shared semantic/navigation index for one text version.
type Document struct {
	AST              *ast.DocumentNode
	Lines            []grammar.TokenLine
	Model            *semantic.SemanticModel
	ParseDiagnostics []validation.Diagnostic
	Diagnostics      []validation.Diagnostic
	Conflicts        []model.ConflictResult
	Declarations     []Declaration
	References       []Reference
	SourceFiles      []SourceFile

	declByKey   map[symbolKey]Declaration
	refsByKey   map[symbolKey][]Reference
	namesByKind map[grammar.DeclarationKind][]string
	byLine      map[int][]Occurrence
}

func Build(ast *ast.DocumentNode, lines []grammar.TokenLine, model *semantic.SemanticModel, parseDiags, diags []validation.Diagnostic, conflicts []model.ConflictResult, refs []Reference) *Document {
	d := &Document{
		AST:              ast,
		Lines:            lines,
		Model:            model,
		ParseDiagnostics: parseDiags,
		Diagnostics:      diags,
		Conflicts:        conflicts,
		Declarations:     buildDeclarations(model),
		References:       refs,
		declByKey:        map[symbolKey]Declaration{},
		refsByKey:        map[symbolKey][]Reference{},
		namesByKind:      map[grammar.DeclarationKind][]string{},
		byLine:           map[int][]Occurrence{},
	}

	d.indexDeclarations()
	d.indexReferences()

	return d
}

// WithNavigationDataFromFallback keeps current text and diagnostics while
// reusing the last successful semantic/navigation structures for LSP navigation.
func (d *Document) WithNavigationDataFromFallback(fallback *Document) *Document {
	if d == nil {
		return fallback
	}
	if fallback == nil {
		return d
	}

	merged := *d
	merged.Model = fallback.Model
	merged.Declarations = fallback.Declarations
	merged.References = fallback.References
	merged.declByKey = fallback.declByKey
	merged.refsByKey = fallback.refsByKey
	merged.namesByKind = fallback.namesByKind
	merged.byLine = fallback.byLine

	return &merged
}

func (d *Document) AllDiagnostics() []validation.Diagnostic {
	if d == nil {
		return nil
	}

	out := make([]validation.Diagnostic, 0, len(d.ParseDiagnostics)+len(d.Diagnostics))
	return append(append(out, d.ParseDiagnostics...), d.Diagnostics...)
}

func (d *Document) LookupDeclaration(kind grammar.DeclarationKind, name string) (Declaration, bool) {
	if d == nil {
		return Declaration{}, false
	}

	decl, ok := d.declByKey[symbolKey{kind, strings.TrimSpace(name)}]
	return decl, ok
}

func (d *Document) ReferencesFor(kind grammar.DeclarationKind, name string) []Reference {
	if d == nil {
		return nil
	}
	refs := d.refsByKey[symbolKey{kind, strings.TrimSpace(name)}]
	if len(refs) == 0 {
		return nil
	}

	out := make([]Reference, len(refs))
	copy(out, refs)

	return out
}

func (d *Document) Names(kind grammar.DeclarationKind) []string {
	if d == nil {
		return nil
	}

	names := d.namesByKind[kind]
	if len(names) == 0 {
		return nil
	}

	out := make([]string, len(names))
	copy(out, names)
	return out
}

func (d *Document) SourceFileForLine(line int) (SourceFile, bool) {
	if d == nil {
		return SourceFile{}, false
	}
	for _, file := range d.SourceFiles {
		if line >= file.StartLine && line <= file.EndLine {
			return file, true
		}
	}
	return SourceFile{}, false
}

func (d *Document) OccurrenceAt(pos sourcepos.Position) (Occurrence, bool) {
	if d == nil {
		return Occurrence{}, false
	}

	var ref *Occurrence
	for i := range d.byLine[pos.Line] {
		occ := &d.byLine[pos.Line][i]
		if !positionInRange(pos, occ.Range) {
			continue
		}
		if occ.Declaration {
			return *occ, true
		}
		if ref == nil {
			ref = occ
		}
	}
	if ref != nil {
		return *ref, true
	}
	return Occurrence{}, false
}

func (d *Document) indexDeclarations() {
	for _, decl := range d.Declarations {
		key := symbolKey{decl.Kind, decl.Name}
		if _, ok := d.declByKey[key]; !ok {
			d.declByKey[key] = decl
			d.namesByKind[decl.Kind] = append(d.namesByKind[decl.Kind], decl.Name)
		}

		d.byLine[decl.Range.Start.Line] = append(d.byLine[decl.Range.Start.Line], Occurrence{
			Kind: decl.Kind, Name: decl.Name, Range: decl.Range, Declaration: true,
		})
	}

	for kind := range d.namesByKind {
		sort.Strings(d.namesByKind[kind])
	}
}

func (d *Document) indexReferences() {
	for _, ref := range d.References {
		key := symbolKey{ref.Kind, ref.Name}
		d.refsByKey[key] = append(d.refsByKey[key], ref)
		d.byLine[ref.Range.Start.Line] = append(d.byLine[ref.Range.Start.Line], Occurrence{
			Kind: ref.Kind, Name: ref.Name, Range: ref.Range, Declaration: false,
		})
	}
}

func addDecl(out []Declaration, kind grammar.DeclarationKind, name string, line int, rng sourcepos.Range) []Declaration {
	if name = strings.TrimSpace(name); name != "" {
		out = append(out, Declaration{Kind: kind, Name: name, Line: line, Range: rng})
	}

	return out
}

func buildDeclarations(m *semantic.SemanticModel) []Declaration {
	if m == nil {
		return nil
	}

	out := make([]Declaration, 0, len(m.Stakeholders)+len(m.Values)+len(m.Requirements))
	for _, s := range m.Stakeholders {
		out = addDecl(out, grammar.DeclarationKindStakeholder, s.Name, s.Line, s.Range)
	}

	for _, v := range m.Values {
		out = addDecl(out, grammar.DeclarationKindValue, v.Name, v.Line, v.Range)
	}

	for _, r := range m.Requirements {
		out = addDecl(out, grammar.DeclarationKindRequirement, r.ID, r.Line, r.Range)
	}

	return out
}

func positionInRange(pos sourcepos.Position, rng sourcepos.Range) bool {
	if pos.Line != rng.Start.Line || pos.Line != rng.End.Line {
		return false
	}

	return pos.Column >= rng.Start.Column && pos.Column < rng.End.Column
}

// Package semantictokens resolves full-document semantic tokens from the shared declaration/reference graph.
package semantictokens

import (
	"sort"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/values"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func Resolve(text string, result coreanalysis.Result) *protocol.SemanticTokens {
	index := result.Index
	lines := grammar.TokenizeSource(text)
	if index != nil && len(index.Lines) > 0 {
		lines = index.Lines
	}

	return &protocol.SemanticTokens{Data: encode(indexMatches(index, lines))}
}

func indexMatches(index *docindex.Document, lines []grammar.TokenLine) []Match {
	if index == nil {
		return nil
	}
	matches := make([]Match, 0, len(index.Declarations)+len(index.References))
	var file docindex.SourceFile
	hasFile := len(index.SourceFiles) > 0
	if hasFile {
		file = index.SourceFiles[0]
	}

	for _, decl := range index.Declarations {
		if role, ok := declarationRole(decl.Kind); ok && validRange(decl.Range) {
			if hasFile && !rangeInSourceFile(decl.Range, file) {
				continue
			}
			matches = append(matches, matchFromRangeInSourceFile(lines, decl.Range, role, file))
		}
	}
	for _, ref := range index.References {
		if role, ok := referenceRole(index, ref); ok && validRange(ref.Range) {
			if hasFile && !rangeInSourceFile(ref.Range, file) {
				continue
			}
			matches = append(matches, matchFromRangeInSourceFile(lines, ref.Range, role, file))
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Line != matches[j].Line {
			return matches[i].Line < matches[j].Line
		}
		if matches[i].StartColumn != matches[j].StartColumn {
			return matches[i].StartColumn < matches[j].StartColumn
		}
		return matches[i].EndColumn < matches[j].EndColumn
	})

	return matches
}

func matchFromRangeInSourceFile(lines []grammar.TokenLine, rng sourcepos.Range, role Role, file docindex.SourceFile) Match {
	match := matchFromRange(lines, rng, role)
	if file.StartLine > 0 {
		match.Line -= file.StartLine - 1
	}
	return match
}

func rangeInSourceFile(rng sourcepos.Range, file docindex.SourceFile) bool {
	return rng.Start.Line >= file.StartLine && rng.Start.Line <= file.EndLine
}

func declarationRole(kind grammar.DeclarationKind) (Role, bool) {
	role := Role(string(kind) + "_declaration")
	_, ok := SpecForRole(role)
	return role, ok
}

func referenceRole(index *docindex.Document, ref docindex.Reference) (Role, bool) {
	if _, ok := index.LookupDeclaration(ref.Kind, ref.Name); ok {
		role := Role(string(ref.Kind) + "_reference")
		if _, ok := SpecForRole(role); ok {
			return role, true
		}
	}

	if ref.Kind == grammar.DeclarationKindValue {
		if _, ok := values.LookupByName(ref.Name); ok {
			return RoleBuiltinValue, true
		}
	}

	return "", false
}

func matchFromRange(lines []grammar.TokenLine, rng sourcepos.Range, role Role) Match {
	start := max(rng.Start.Column-1, 0)
	end := max(rng.End.Column-1, start)
	if line, ok := tokenLineFor(lines, rng.Start.Line); ok {
		if converted, valid := sourcepos.UTF16OffsetFromByteOK(line.Raw, start); valid {
			start = converted
		}
	}
	if line, ok := tokenLineFor(lines, rng.End.Line); ok {
		if converted, valid := sourcepos.UTF16OffsetFromByteOK(line.Raw, end); valid {
			end = converted
		}
	}

	return Match{
		Line:        rng.Start.Line,
		StartColumn: start + 1,
		EndColumn:   end + 1,
		Role:        role,
	}
}

func validRange(rng sourcepos.Range) bool {
	return rng.Start.Line > 0 && rng.Start.Line == rng.End.Line && rng.Start.Column > 0 && rng.End.Column > rng.Start.Column
}

func encode(matches []Match) []protocol.UInteger {
	if len(matches) == 0 {
		return nil
	}

	typeIndexes := indexMap(LegendTypes())
	modifierIndexes := indexMap(LegendModifiers())
	data := make([]protocol.UInteger, 0, len(matches)*5)
	previousLine := 1
	previousStart := 1

	for _, match := range matches {
		spec, ok := SpecForRole(match.Role)
		if !ok {
			continue
		}

		deltaLine := match.Line - previousLine
		deltaStart := match.StartColumn - previousStart
		if deltaLine != 0 {
			deltaStart = match.StartColumn - 1
		}

		modifiers := 0
		for _, m := range spec.Modifiers {
			if index, ok := modifierIndexes[string(m)]; ok {
				modifiers |= 1 << index
			}
		}
		data = append(data,
			protocol.UInteger(max(deltaLine, 0)),
			protocol.UInteger(max(deltaStart, 0)),
			protocol.UInteger(max(match.EndColumn-match.StartColumn, 0)),
			protocol.UInteger(typeIndexes[string(spec.Type)]),
			protocol.UInteger(modifiers),
		)

		previousLine, previousStart = match.Line, match.StartColumn
	}

	return data
}

func indexMap(values []string) map[string]int {
	out := make(map[string]int, len(values))
	for i, v := range values {
		out[v] = i
	}
	return out
}

func tokenLineFor(lines []grammar.TokenLine, number int) (grammar.TokenLine, bool) {
	if i := number - 1; i >= 0 && i < len(lines) {
		return lines[i], true
	}
	return grammar.TokenLine{}, false
}

package parser

import (
	"strings"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
)

type Diagnostic struct {
	Line   int
	Column int
	Code   string
}

type ParseResult struct {
	AST         *ast.DocumentNode
	Diagnostics []Diagnostic
	References  []ReferenceOccurrence
	Lines       []grammar.TokenLine
}

type ReferenceOccurrence struct {
	Kind  grammar.DeclarationKind
	Name  string
	Line  int
	Range sourcepos.Range
}

func Parse(text string) ParseResult {
	lines := grammar.TokenizeSource(text)
	result := ParseResult{
		AST: &ast.DocumentNode{
			Declarations: make([]ast.DeclarationNode, 0),
			Malformed:    make([]ast.MalformedLineNode, 0),
		},
		Diagnostics: make([]Diagnostic, 0),
		References:  make([]ReferenceOccurrence, 0),
		Lines:       lines,
	}

	for _, block := range parseBlocksFromLines(lines) {
		parseBlock(&result, block)
	}

	return result
}

func parseBlock(result *ParseResult, block structuralBlock) {
	decl, ok := grammar.Compiled.DeclarationByKind[block.Kind]
	if !ok {
		result.AST.Malformed = append(result.AST.Malformed, malformedLineNode(block.Header, "unknown line kind"))
		appendUnclassifiedDiagnostic(&result.Diagnostics, block.Header)
		return
	}

	headerMatch := grammar.MatchDeclaration(block.Kind, block.Header)
	appendGrammarDiagnostics(&result.Diagnostics, block.Header.Line, headerMatch.Diagnostics)
	headerNode := genericNodeFromMatch(headerMatch)

	node := ast.DeclarationNode{
		Kind:      block.Kind,
		Line:      block.Header.Line,
		Range:     headerNode.Range,
		Header:    headerNode,
		Body:      make([]ast.GenericNode, 0, len(block.Body)),
		Malformed: make([]ast.MalformedLineNode, 0),
	}

	appendGenericNodeReferences(&result.References, headerNode, block.Kind)

	observed := make([]grammar.MatchResult, 0, len(block.Body))
	for _, line := range block.Body {
		blm, matched := grammar.MatchBody(decl, line)
		if !matched {
			node.Malformed = append(node.Malformed, malformedLineNode(line, "unknown line kind"))
			appendUnclassifiedDiagnostic(&result.Diagnostics, line)
			continue
		}
		appendGrammarDiagnostics(&result.Diagnostics, line.Line, blm.Diagnostics)
		observed = append(observed, blm)

		bodyNode := genericNodeFromMatch(blm)
		node.Body = append(node.Body, bodyNode)
		appendGenericNodeReferences(&result.References, bodyNode, "")
	}
	appendGrammarDiagnostics(&result.Diagnostics, block.Header.Line, grammar.BodyStructureDiagnostics(decl, observed, block.Header.Line))

	result.AST.Declarations = append(result.AST.Declarations, node)
}

func genericNodeFromMatch(match grammar.MatchResult) ast.GenericNode {
	fields := make(map[string][]ast.FieldValue)
	for name, tokens := range match.Match.Fields {
		for _, token := range tokens {
			value := ast.FieldValue{Text: token.Text, Range: token.Range(match.Line)}
			if token.RefKind != "" {
				value.RefKind = token.RefKind
				value.HasRef = true
			}
			fields[name] = append(fields[name], value)
		}
	}

	for _, keyword := range match.Match.Keywords {
		fields[ast.FieldKeyword] = append(fields[ast.FieldKeyword], ast.FieldValue{Text: keyword.Text, Range: keyword.Range(match.Line)})
	}
	for _, extra := range match.Match.Extra {
		fields[ast.FieldExtra] = append(fields[ast.FieldExtra], ast.FieldValue{Text: extra.Text, Range: extra.Range(match.Line)})
	}
	if match.RawText != "" {
		fields[ast.FieldRawText] = append(fields[ast.FieldRawText], ast.FieldValue{Text: match.RawText, Range: match.RawRange})
	}

	return ast.GenericNode{Kind: match.NodeKind, Line: match.Line, Range: match.Range, Fields: fields}
}

func appendGenericNodeReferences(out *[]ReferenceOccurrence, node ast.GenericNode, skipKind grammar.DeclarationKind) {
	for _, values := range node.Fields {
		for _, value := range values {
			if !value.HasRef || value.RefKind == skipKind || strings.TrimSpace(value.Text) == "" {
				continue
			}
			*out = append(*out, ReferenceOccurrence{
				Kind:  value.RefKind,
				Name:  value.Text,
				Line:  node.Line,
				Range: value.Range,
			})
		}
	}
}

func appendUnclassifiedDiagnostic(diagnostics *[]Diagnostic, line grammar.TokenLine) {
	*diagnostics = append(*diagnostics, Diagnostic{
		Line:   line.Line,
		Column: 1,
		Code:   "parse.line.unclassified",
	})
}

func appendGrammarDiagnostics(diagnostics *[]Diagnostic, fallbackLine int, items []grammar.MatchDiagnostic) {
	for _, item := range items {
		line := item.Line
		if line == 0 {
			line = fallbackLine
		}

		*diagnostics = append(*diagnostics, Diagnostic{Line: line, Column: max(item.Column, 1), Code: item.Code})
	}
}

func malformedLineNode(line grammar.TokenLine, reason string) ast.MalformedLineNode {
	keyword := ""
	if len(line.Tokens) > 0 {
		keyword = line.Tokens[0].Text
	}

	return ast.MalformedLineNode{
		Line: line.Line, Range: line.Range(), Kind: "unknown",
		Keyword: strings.TrimSpace(keyword), RawText: strings.TrimSpace(line.Raw), Reason: strings.TrimSpace(reason),
	}
}

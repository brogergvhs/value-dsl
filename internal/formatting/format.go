// Package formatting parses DSL source into the generic AST and prints it back
// with canonical block spacing, token spacing, quotes, and number rendering.
package formatting

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/parser"
)

func formatNumber(s string) string {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	return s
}

func Format(text string) (string, error) {
	parsed := parser.Parse(text)
	return printDocument(parsed, parseComments(parsed.Lines), len(text)+len(text)/8), nil
}

// FormatForWrite refuses in-place changes when malformed input is present.
// Plain formatting may still return the original text for inspection.
func FormatForWrite(text string) (string, error) {
	parsed := parser.Parse(text)
	malformed := 0

	if parsed.AST != nil {
		malformed = len(parsed.AST.Malformed)
		for _, decl := range parsed.AST.Declarations {
			malformed += len(decl.Malformed)
		}
	}
	if len(parsed.Diagnostics) != 0 || malformed != 0 {
		return "", &unsafeWriteError{ParseDiagnostics: len(parsed.Diagnostics), MalformedLines: malformed}
	}

	return printDocument(parsed, parseComments(parsed.Lines), len(text)+len(text)/8), nil
}

type unsafeWriteError struct {
	ParseDiagnostics int
	MalformedLines   int
}

func (e *unsafeWriteError) Error() string {
	return fmt.Sprintf(
		"format --write refused: input contains %d parser diagnostics and %d malformed lines; run format without --write to inspect preserved output",
		e.ParseDiagnostics,
		e.MalformedLines,
	)
}

type linePlan struct {
	text     string
	raw      bool
	blockTop bool
}

// printDocument orchestrates the three formatting stages: build plans, collect
// comments, then emit in source-line order so comment associations stay stable.
func printDocument(parsed parser.ParseResult, comments commentMap, capacity int) string {
	if parsed.AST == nil {
		return ""
	}

	out := &output{comments: comments}
	out.b.Grow(capacity)
	out.emitDocument(parsed, buildPlans(parsed))
	return out.string()
}

// buildPlans assigns each non-blank source line a linePlan. Recognized AST
// nodes get a rendered canonical text, malformed or unknown lines get raw
// passthrough. A second pass over sorted block starts sets blockTop on the
// first line of each new block group to trigger blank-line separation.
func buildPlans(parsed parser.ParseResult) map[int]linePlan {
	plans := make(map[int]linePlan, len(parsed.Lines))
	type block struct {
		line  int
		group blockGroupKey
	}

	blocks := make([]block, 0, len(parsed.AST.Malformed)+len(parsed.AST.Declarations))
	for _, bad := range parsed.AST.Malformed {
		plans[bad.Line] = linePlan{text: rawLine(parsed.Lines, bad.Line), raw: true}
		blocks = append(blocks, block{line: bad.Line, group: malformedBlockGroup(bad.Line)})
	}

	for _, decl := range parsed.AST.Declarations {
		spec, ok := grammar.Compiled.DeclarationByKind[decl.Kind]
		if !ok {
			plans[decl.Line] = linePlan{text: rawLine(parsed.Lines, decl.Line), raw: true}
			blocks = append(blocks, block{line: decl.Line, group: malformedBlockGroup(decl.Line)})
			continue
		}

		plans[decl.Line] = linePlan{text: render(spec.Header, decl.Header)}
		blocks = append(blocks, block{line: decl.Line, group: blockGroup(decl.Kind, decl.Line)})
		for _, node := range decl.Body {
			text := strings.TrimSpace(node.First(ast.FieldRawText))
			if rule, ok := ruleForNode(spec, node); ok {
				text = render(rule.Pattern, node)
			}
			plans[node.Line] = linePlan{text: text}
		}

		for _, bad := range decl.Malformed {
			plans[bad.Line] = linePlan{text: rawLine(parsed.Lines, bad.Line), raw: true}
		}
	}

	slices.SortFunc(blocks, func(a, b block) int { return a.line - b.line })
	prev := blockGroupKey{}
	for _, block := range blocks {
		plan := plans[block.line]
		plan.blockTop = block.group != prev
		plans[block.line] = plan
		prev = block.group
	}

	return plans
}

type blockGroupKey struct {
	kind      grammar.DeclarationKind
	line      int
	malformed bool
}

// blockGroup returns the grouping key for blank-line separation. DenseGroup
// kinds (e.g. stakeholder) share a single key so consecutive declarations of
// that kind are not separated, all other declarations get a per-line key.
func blockGroup(kind grammar.DeclarationKind, line int) blockGroupKey {
	if decl, ok := grammar.Compiled.DeclarationByKind[kind]; ok && decl.DenseGroup {
		return blockGroupKey{kind: kind}
	}
	return blockGroupKey{kind: kind, line: line}
}

func malformedBlockGroup(line int) blockGroupKey {
	return blockGroupKey{line: line, malformed: true}
}

func rawLine(lines []grammar.TokenLine, line int) string {
	if line <= 0 || line > len(lines) {
		return ""
	}
	return lines[line-1].Raw
}

const maxRenderDepth = 32

// render prints a matched grammar pattern with uniform token normalization.
func render(pattern []grammar.Matcher, node ast.GenericNode) string {
	return renderPattern(pattern, node, 0)
}

func renderPattern(pattern []grammar.Matcher, node ast.GenericNode, depth int) string {
	if depth > maxRenderDepth {
		return strings.TrimSpace(node.First(ast.FieldRawText))
	}

	parts := make([]string, 0, len(pattern))
	for _, matcher := range pattern {
		switch m := matcher.(type) {
		case grammar.Kw:
			if keywordSeen(node, m.Text) {
				parts = append(parts, m.Text)
			}
		case grammar.List:
			values := make([]string, 0, len(node.Fields[m.Name]))
			for _, value := range node.Texts(m.Name) {
				value, quote := renderValue(m.Item, value)
				values = appendScalar(values, value, quote)
			}
			if len(values) > 0 {
				sep := m.Separator
				if sep == "," {
					sep = ", "
				}
				parts = append(parts, strings.Join(values, sep))
			}
		case grammar.Opt:
			if patternHasData(m.Inner, node, depth+1) {
				parts = append(parts, renderPattern(m.Inner, node, depth+1))
			}
		default:
			value, quote := renderCapture(matcher, node)
			parts = appendScalar(parts, value, quote)
		}
	}

	if len(parts) == 0 {
		return strings.TrimSpace(node.First(ast.FieldRawText))
	}

	var b strings.Builder
	for i, part := range parts {
		if i > 0 && part != "," {
			b.WriteByte(' ')
		}
		b.WriteString(part)
	}
	return b.String()
}

func appendScalar(parts []string, value string, quote bool) []string {
	if value = strings.TrimSpace(value); value == "" {
		return parts
	}
	if quote && needsQuotes(value) {
		value = "'" + strings.ReplaceAll(value, "'", `\'`) + "'"
	}

	return append(parts, value)
}

func needsQuotes(value string) bool {
	tokens := grammar.TokenizeRawLine(1, value).Tokens
	return len(tokens) != 1 || tokens[0].Text != value
}

func patternHasData(pattern []grammar.Matcher, node ast.GenericNode, depth int) bool {
	if depth > maxRenderDepth {
		return false
	}

	for _, m := range pattern {
		switch typed := m.(type) {
		case grammar.Kw:
			if keywordSeen(node, typed.Text) {
				return true
			}
		case grammar.List:
			return len(node.Fields[typed.Name]) > 0
		case grammar.Opt:
			if patternHasData(typed.Inner, node, depth+1) {
				return true
			}
		default:
			s, _ := renderCapture(m, node)
			return s != ""
		}
	}

	return false
}

func renderCapture(matcher grammar.Matcher, node ast.GenericNode) (string, bool) {
	var name string
	switch m := matcher.(type) {
	case grammar.Ident:
		name = m.Name
	case grammar.Tok:
		name = m.Name
	case grammar.Ref:
		name = m.Name
	case grammar.Enum:
		name = m.Name
	case grammar.Str:
		name = m.Name
	case grammar.Free:
		name = m.Name
	}

	if name == "" {
		return "", false
	}

	return renderValue(matcher, node.First(name))
}

func renderValue(matcher grammar.Matcher, value string) (string, bool) {
	switch m := matcher.(type) {
	case grammar.Tok:
		if m.Kind == grammar.TokKindNumber {
			return formatNumber(value), false
		}
	case grammar.Str:
		return value, true
	}

	return value, false
}

func keywordSeen(node ast.GenericNode, keyword string) bool {
	for token := range strings.FieldsSeq(keyword) {
		if !node.HasText(ast.FieldKeyword, token) {
			return false
		}
	}

	return keyword != ""
}

// ruleForNode selects the grammar line rule for a body node. It prefers the
// first rule whose pattern's token sets match the node's captured values
// (disambiguating e.g. different action verb groups), falls back to the first
// kind-matching rule so partially parsed nodes still get a canonical rendering.
func ruleForNode(spec grammar.Decl, node ast.GenericNode) (grammar.LineRule, bool) {
	if spec.Body == nil {
		return grammar.LineRule{}, false
	}

	var fallback *grammar.LineRule
	for i := range spec.Body.Lines {
		rule := &spec.Body.Lines[i]
		wantKind := grammar.LineRuleNodeName(*rule)
		if rule.ClauseKind != "" {
			wantKind = string(rule.ClauseKind)
		}
		if node.Kind != wantKind {
			continue
		}
		if tokenSetsMatch(rule.Pattern, node) {
			return *rule, true
		}
		if fallback == nil {
			fallback = rule
		}
	}

	if fallback == nil {
		return grammar.LineRule{}, false
	}

	return *fallback, true
}

func tokenSetsMatch(pattern []grammar.Matcher, node ast.GenericNode) bool {
	for _, m := range pattern {
		switch typed := m.(type) {
		case grammar.Tok:
			if len(typed.Set) > 0 && node.First(typed.Name) != "" && !slices.Contains(typed.Set, node.First(typed.Name)) {
				return false
			}
		case grammar.Opt:
			if !tokenSetsMatch(typed.Inner, node) {
				return false
			}
		}
	}

	return true
}

type commentMap struct {
	leading  map[int][]string
	inline   map[int]string
	trailing []string
}

// parseComments classifies each comment relative to code lines. Comments on
// blank lines are held as "pending" and become leading comments of the next
// non-blank code line; any remaining pending comments at EOF become trailing.
func parseComments(lines []grammar.TokenLine) commentMap {
	leading := make(map[int][]string)
	inline := make(map[int]string)
	var pending []string

	for _, line := range lines {
		hasComment := line.CommentOffset >= 0 && line.CommentOffset+2 <= len(line.Raw)
		var comment string
		if hasComment {
			comment = strings.TrimSpace(line.Raw[line.CommentOffset+2:])
		}

		if strings.TrimSpace(line.Code) == "" {
			if hasComment {
				pending = append(pending, comment)
			}
			continue
		}

		if len(pending) > 0 {
			leading[line.Line] = slices.Clone(pending)
			pending = pending[:0]
		}

		if hasComment {
			inline[line.Line] = comment
		}
	}

	return commentMap{leading: leading, inline: inline, trailing: slices.Clone(pending)}
}

// output is a stateful line writer. wrote/blanked track pending newlines so
// blank() emits at most one empty line and write() never produces a leading
// newline on the first line.
type output struct {
	b        strings.Builder
	wrote    bool
	blanked  bool
	comments commentMap
}

func (o *output) emit(line int, text string, raw bool) {
	for _, comment := range o.comments.leading[line] {
		o.write(lineComment(grammar.Compiled.Spec.LineComment, comment))
	}

	if !raw {
		if comment, ok := o.comments.inline[line]; ok {
			text += lineComment(" "+grammar.Compiled.Spec.LineComment, comment)
		}
	}

	o.write(text)
}

func (o *output) blank() {
	if o.wrote && !o.blanked {
		o.write("")
	}
}

func (o *output) string() string {
	for _, comment := range o.comments.trailing {
		o.write(lineComment(grammar.Compiled.Spec.LineComment, comment))
	}
	if !o.wrote {
		return ""
	}

	o.b.WriteByte('\n')
	return o.b.String()
}

func (o *output) write(text string) {
	if o.wrote {
		o.b.WriteByte('\n')
	}

	o.b.WriteString(strings.TrimRight(text, " \t"))
	o.wrote = true
	o.blanked = text == ""
}

func lineComment(prefix, comment string) string {
	if comment == "" {
		return prefix
	}
	return prefix + " " + comment
}

func (o *output) emitDocument(parsed parser.ParseResult, plans map[int]linePlan) {
	for _, line := range parsed.Lines {
		if line.IsBlank() {
			continue
		}

		if plan, ok := plans[line.Line]; ok {
			if plan.blockTop {
				o.blank()
			}
			o.emit(line.Line, plan.text, plan.raw)
		}
	}
}

package grammar

import (
	"slices"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/sourcepos"
)

// Token is one lossless line token with source columns and lexer metadata.
type Token struct {
	Text        string
	Kind        TokenClass
	StartColumn int32
	EndColumn   int32
	RefKind     DeclarationKind
}

func (t Token) Range(line int) sourcepos.Range {
	return sourcepos.NewRange(sourcepos.NewPosition(line, int(t.StartColumn)), sourcepos.NewPosition(line, int(t.EndColumn)))
}

// TokenClass is the lexer-level token category stored compactly on each token.
type TokenClass uint8

const (
	TokenClassUnknown TokenClass = iota
	TokenClassIdent
	TokenClassNumber
	TokenClassString
	TokenClassWhenVerb
	TokenClassActionVerb
)

const maxMatcherDepth = 32

// TokenLine is the shared tokenized view of one source line.
type TokenLine struct {
	Line          int
	Raw           string
	Code          string
	Tokens        []Token
	CommentOffset int
}

func (l TokenLine) IsBlank() bool { return len(l.Tokens) == 0 }

func (l TokenLine) Range() sourcepos.Range {
	if len(l.Tokens) == 0 {
		return sourcepos.LineRange(l.Line)
	}
	return tokenRange(l, 0, len(l.Tokens))
}

func (l TokenLine) VisibleOffset(offset int) (int, bool) {
	offset = min(max(offset, 0), len(l.Raw))
	if l.CommentOffset >= 0 && offset >= l.CommentOffset {
		return l.CommentOffset, true
	}
	return min(offset, len(l.Code)), false
}

func (l TokenLine) TokenAtOffset(offset int) (Token, bool) {
	for _, token := range l.Tokens {
		start := int(token.StartColumn) - 1
		end := int(token.EndColumn) - 1
		if offset >= start && offset < end {
			return token, true
		}
	}

	return Token{}, false
}

func TokenizeSource(text string) []TokenLine {
	raw := strings.Split(text, "\n")
	lines := make([]TokenLine, 0, len(raw))

	for i, line := range raw {
		lines = append(lines, TokenizeRawLine(i+1, line))
	}
	if len(lines) == 0 {
		return []TokenLine{{Line: 1, CommentOffset: -1}}
	}

	return lines
}

func TokenizeRawLine(line int, raw string) TokenLine {
	tokens, commentOffset := tokenizeRawLine(raw)
	code := raw
	if commentOffset >= 0 {
		code = raw[:commentOffset]
	}

	return TokenLine{
		Line:          line,
		Raw:           raw,
		Code:          code,
		Tokens:        tokens,
		CommentOffset: commentOffset,
	}
}

// LineMatch contains the captures and keywords produced by one rule match.
type LineMatch struct {
	Fields   map[string][]Token
	Keywords []Token
	Extra    []Token
}

func (m LineMatch) FirstCapture() (Token, bool) {
	bestCol := int32(1<<31 - 1)
	var best Token
	ok := false

	for _, tokens := range m.Fields {
		if len(tokens) > 0 && tokens[0].StartColumn > 0 && tokens[0].StartColumn < bestCol {
			bestCol, best, ok = tokens[0].StartColumn, tokens[0], true
		}
	}

	return best, ok
}

// MatchDiagnostic is a grammar-shape diagnostic emitted by matcher primitives.
type MatchDiagnostic struct {
	Line   int
	Code   string
	Column int
}

// MatchResult binds one source line to a grammar rule and its captures.
type MatchResult struct {
	Match       LineMatch
	Rule        LineRule
	NodeKind    string
	Line        int
	Range       sourcepos.Range
	RawText     string
	RawRange    sourcepos.Range
	Diagnostics []MatchDiagnostic
}

const (
	captureIdent = 1 << iota
	captureNumber
	captureString
	captureEnum
)

// MatchDeclaration matches a top-level declaration header and derives its
// canonical node range from the earliest capture.
func MatchDeclaration(kind DeclarationKind, line TokenLine) MatchResult {
	decl, ok := Compiled.DeclarationByKind[kind]
	if !ok {
		return MatchResult{}
	}
	return matchedLine(decl.Header, line, decl.Kind, string(kind), string(kind), -1)
}

// MatchBody selects a compiled line rule, applies it, and derives the body
// node range from the leading grammar keyword sequence.
func MatchBody(decl Decl, line TokenLine) (MatchResult, bool) {
	if decl.Body == nil || len(line.Tokens) == 0 {
		return MatchResult{}, false
	}

	rule, ok := Compiled.BodyLineRule(decl, line.Tokens)
	if !ok {
		return MatchResult{}, false
	}

	result := matchedLine(rule.Pattern, line, "", Compiled.DiagnosticPrefix(decl.Kind, rule), LineRuleKey(rule), leadingKeywordLen(rule.Pattern, line.Tokens))
	result.Rule = rule
	return result, true
}

// BodyStructureDiagnostics derives missing, duplicate, and order diagnostics
// from already matched body lines.
func BodyStructureDiagnostics(decl Decl, observed []MatchResult, fallbackLine int) []MatchDiagnostic {
	if decl.Body == nil {
		return nil
	}

	rules := make(map[string]LineRule, len(decl.Body.Lines))
	counts := make(map[string]int, len(decl.Body.Lines))
	repeatable := make(map[string]bool, len(decl.Body.Lines))
	required := make(map[string]bool, len(decl.Body.Lines))
	order := make([]string, 0, len(decl.Body.Lines))

	for _, rule := range decl.Body.Lines {
		key := LineRuleKey(rule)
		if _, ok := rules[key]; !ok {
			order = append(order, key)
			rules[key] = rule
		}
		required[key] = required[key] || rule.Required
		repeatable[key] = repeatable[key] || rule.Repeatable
	}

	diags := make([]MatchDiagnostic, 0)
	lastOrder := -1
	for _, seen := range observed {
		key := LineRuleKey(seen.Rule)
		if _, ok := rules[key]; !ok {
			continue
		}

		if seen.Rule.Order < lastOrder {
			diags = append(diags, matchDiagnostic(seen.Line, seen.Range.Start.Column, Compiled.DiagnosticCode(Compiled.DiagnosticPrefix(decl.Kind, seen.Rule), "order", "invalid")))
		} else if seen.Rule.Order > 0 {
			lastOrder = seen.Rule.Order
		}

		counts[key]++
		if counts[key] > 1 && !repeatable[key] {
			diags = append(diags, matchDiagnostic(seen.Line, seen.Range.Start.Column, Compiled.DiagnosticCode(Compiled.DiagnosticPrefix(decl.Kind, seen.Rule), "duplicate")))
		}
	}

	for _, key := range order {
		if required[key] && counts[key] == 0 {
			diags = append(diags, matchDiagnostic(fallbackLine, 1, Compiled.DiagnosticCode(Compiled.DiagnosticPrefix(decl.Kind, rules[key]), "missing")))
		}
	}

	return diags
}

func matchedLine(pattern []Matcher, line TokenLine, identRefKind DeclarationKind, prefix, nodeKind string, leading int) MatchResult {
	result := MatchResult{
		Line:  line.Line,
		Range: line.Range(),
		Match: LineMatch{Fields: make(map[string][]Token)},
	}
	index := captureMatchers(pattern, line.Tokens, 0, identRefKind, prefix, &result, 0)
	if index < len(line.Tokens) {
		result.Match.Extra = append(result.Match.Extra, line.Tokens[index:]...)
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, "extra_tokens"), int(line.Tokens[index].StartColumn))
	}
	result.NodeKind = nodeKind
	result.RawText, result.RawRange = rawText(line, leading)
	if leading < 0 {
		if token, ok := result.Match.FirstCapture(); ok {
			result.Range = token.Range(line.Line)
		} else {
			result.Range = tokenRange(line, 0, len(result.Match.Keywords))
		}
	} else {
		result.Range = tokenRange(line, 0, leading)
	}
	return result
}

func captureMatchers(pattern []Matcher, tokens []Token, index int, identRefKind DeclarationKind, prefix string, result *MatchResult, depth int) int {
	if depth > maxMatcherDepth {
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, "pattern", "depth", "exceeded"), missingColumn(tokens, index))
		return index
	}

	for i, matcher := range pattern {
		switch part := matcher.(type) {
		case Kw:
			index = consumeKeyword(part.Text, tokens, index, prefix, result)
		case Ident:
			index = captureValue(pattern, i, part.Name, identRefKind, nil, captureIdent, tokens, index, prefix, result)
		case Tok:
			mode := 0
			if part.Kind == TokKindNumber {
				mode = captureNumber
			}
			index = captureValue(pattern, i, part.Name, "", part.Set, mode, tokens, index, prefix, result)
		case Ref:
			index = captureValue(pattern, i, part.Name, part.Kind, nil, captureIdent, tokens, index, prefix, result)
		case Enum:
			index = captureValue(pattern, i, part.Name, "", part.Values, captureEnum, tokens, index, prefix, result)
		case Str:
			index = captureValue(pattern, i, part.Name, "", nil, captureString, tokens, index, prefix, result)
		case List:
			index = captureList(pattern, i, part, tokens, index, prefix, result)
		case Free:
			if index < len(tokens) {
				addCaptureToken(&result.Match, part.Name, joinTokens(tokens[index:]), "")
				index = len(tokens)
			} else if part.Name != "" {
				result.addDiagnostic(Compiled.DiagnosticCode(prefix, part.Name, "missing"), missingColumn(tokens, index))
			}
		case Opt:
			if shouldEnterOptional(part.Inner, tokens, index) {
				index = captureMatchers(part.Inner, tokens, index, "", prefix, result, depth+1)
			}
		}
	}

	return index
}

func consumeKeyword(keyword string, tokens []Token, index int, prefix string, result *MatchResult) int {
	for word := range strings.FieldsSeq(keyword) {
		if index >= len(tokens) || tokens[index].Text != word {
			result.addDiagnostic(Compiled.DiagnosticCode(prefix, "keyword", diagnosticPart(word), "missing"), missingColumn(tokens, index))
			return index
		}

		result.Match.Keywords = append(result.Match.Keywords, tokens[index])
		index++
	}

	return index
}

func captureValue(pattern []Matcher, position int, name string, refKind DeclarationKind, set []string, mode int, tokens []Token, index int, prefix string, result *MatchResult) int {
	if name == "" {
		return index
	}

	gate := nextKeywordGate(pattern, position+1)
	if index >= len(tokens) || (gate != "" && tokens[index].Text == gate) {
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, name, "missing"), missingColumn(tokens, index))
		return index
	}

	token := tokens[index]
	if mode&captureString != 0 && len(token.Text) >= 2 && (token.Text[0] == '"' || token.Text[0] == '\'') && token.Text[len(token.Text)-1] == token.Text[0] {
		token.Text = token.Text[1 : len(token.Text)-1]
	}
	addCaptureToken(&result.Match, name, token, refKind)

	if mode&captureIdent != 0 && !Compiled.IsValidIdentifier(token.Text) {
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, "identifier", "invalid"), int(token.StartColumn))
	}
	if mode&captureNumber != 0 && !Compiled.IsValidNumber(token.Text) {
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, name, "invalid"), int(token.StartColumn))
	}

	if len(set) > 0 && !slices.Contains(set, token.Text) {
		if mode&captureEnum != 0 {
			result.addDiagnostic(Compiled.DiagnosticCode(prefix, name, "invalid"), int(token.StartColumn))
		} else {
			result.addDiagnostic(Compiled.DiagnosticCode(prefix, name, "unsupported"), int(token.StartColumn))
		}
	}

	return index + 1
}

func captureList(pattern []Matcher, position int, list List, tokens []Token, index int, prefix string, result *MatchResult) int {
	if list.Name == "" {
		return index
	}

	gate := nextKeywordGate(pattern, position+1)
	if index >= len(tokens) || (gate != "" && tokens[index].Text == gate) {
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, list.Name, "missing"), missingColumn(tokens, index))
		return index
	}

	end := len(tokens)
	if gate != "" {
		for current := index; current < len(tokens); current++ {
			if tokens[current].Text == gate {
				end = current
				break
			}
		}
	}

	refKind := DeclarationKind("")
	checkIdentifier := false
	switch item := list.Item.(type) {
	case Ref:
		checkIdentifier = true
		refKind = item.Kind
	case Ident:
		checkIdentifier = true
	}

	values := splitListValues(tokens[index:end], list.Separator)
	if len(values) == 0 {
		result.addDiagnostic(Compiled.DiagnosticCode(prefix, list.Name, "missing"), missingColumn(tokens, index))
	}

	for _, value := range values {
		if checkIdentifier && !Compiled.IsValidIdentifier(value.Text) {
			result.addDiagnostic(Compiled.DiagnosticCode(prefix, "identifier", "invalid"), int(value.StartColumn))
		}
		addCaptureToken(&result.Match, list.Name, value, refKind)
	}

	return end
}

func tokenizeRawLine(line string) ([]Token, int) {
	tokens := make([]Token, 0)
	commentOffset := -1

	for index := 0; index < len(line); {
		switch {
		case line[index] == ' ' || line[index] == '\t':
			index++
		case line[index] == '/' && index+1 < len(line) && line[index+1] == '/':
			return tokens, index
		case line[index] == '"' || line[index] == '\'':
			start := index
			index = stringTokenEnd(line, index)

			tokens = append(tokens, Token{Text: line[start:index], Kind: TokenClassString, StartColumn: int32(start + 1), EndColumn: int32(index + 1)})
		case punctuationAt(line, index) != "":
			punct := punctuationAt(line, index)
			tokens = append(tokens, Token{Text: punct, Kind: TokenClassUnknown, StartColumn: int32(index + 1), EndColumn: int32(index + len(punct) + 1)})
			index += len(punct)
		default:
			start := index
			for index < len(line) {
				switch {
				case line[index] == ' ' || line[index] == '\t':
					goto tokenDone
				case line[index] == '"' || line[index] == '\'':
					goto tokenDone
				case line[index] == '/' && index+1 < len(line) && line[index+1] == '/':
					commentOffset = index
					goto tokenDone
				case punctuationAt(line, index) != "":
					goto tokenDone
				default:
					index++
				}
			}
		tokenDone:
			if start < index {
				text := line[start:index]
				tokens = append(tokens, Token{Text: text, Kind: classifyToken(text), StartColumn: int32(start + 1), EndColumn: int32(index + 1)})
			}
			if commentOffset >= 0 {
				return tokens, commentOffset
			}
		}
	}

	return tokens, commentOffset
}

func stringTokenEnd(line string, start int) int {
	quote := line[start]
	index := start + 1
	for index < len(line) {
		if line[index] == quote && !escapedByOddBackslashes(line, index) {
			return index + 1
		}
		index++
	}
	return index
}

func escapedByOddBackslashes(line string, index int) bool {
	count := 0
	for i := index - 1; i >= 0 && line[i] == '\\'; i-- {
		count++
	}
	return count%2 == 1
}

func classifyToken(text string) TokenClass {
	_, isActionVerb := Compiled.actionVerbSet[text]
	_, isWhenVerb := Compiled.whenVerbSet[text]

	if len(text) >= 2 && (text[0] == '"' || text[0] == '\'') {
		return TokenClassString
	}
	if Compiled.IsValidNumber(text) {
		return TokenClassNumber
	}
	if isActionVerb {
		return TokenClassActionVerb
	}
	if isWhenVerb {
		return TokenClassWhenVerb
	}
	if Compiled.IsValidIdentifier(text) {
		return TokenClassIdent
	}

	return TokenClassUnknown
}

func punctuationAt(line string, index int) string {
	for _, punct := range Compiled.PunctuationTokens {
		if strings.HasPrefix(line[index:], punct) {
			return punct
		}
	}

	return ""
}

func (r *MatchResult) addDiagnostic(code string, column int) {
	r.Diagnostics = append(r.Diagnostics, matchDiagnostic(r.Line, column, code))
}

func matchDiagnostic(line, column int, code string) MatchDiagnostic {
	if column <= 0 {
		column = 1
	}
	return MatchDiagnostic{Line: line, Code: code, Column: column}
}

func shouldEnterOptional(pattern []Matcher, tokens []Token, index int) bool {
	gate := LeadingKeyword(pattern)
	if gate == "" {
		return index < len(tokens)
	}

	return index < len(tokens) && tokens[index].Text == gate
}

func nextKeywordGate(pattern []Matcher, start int) string {
	for i := start; i < len(pattern); i++ {
		switch part := pattern[i].(type) {
		case Kw:
			return part.Text
		case Opt:
			if gate := LeadingKeyword(part.Inner); gate != "" {
				return gate
			}
		}
	}

	return ""
}

func splitListValues(tokens []Token, separator string) []Token {
	if len(tokens) == 0 {
		return nil
	}
	if separator == "" {
		return []Token{joinTokens(tokens)}
	}

	var values, current []Token
	for _, token := range tokens {
		if token.Text == separator {
			if len(current) > 0 {
				values = append(values, joinTokens(current))
				current = current[:0]
			}
			continue
		}
		current = append(current, token)
	}

	if len(current) > 0 {
		values = append(values, joinTokens(current))
	}

	return values
}

func addCaptureToken(match *LineMatch, name string, token Token, refKind DeclarationKind) {
	token.Text = strings.TrimSpace(token.Text)
	if name == "" || token.Text == "" {
		return
	}

	token.RefKind = refKind
	match.Fields[name] = append(match.Fields[name], token)
}

func joinTokens(tokens []Token) Token {
	if len(tokens) == 0 {
		return Token{}
	}

	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = t.Text
	}

	return Token{Text: strings.Join(parts, " "), StartColumn: tokens[0].StartColumn, EndColumn: tokens[len(tokens)-1].EndColumn}
}

func rawText(line TokenLine, skip int) (string, sourcepos.Range) {
	if skip <= 0 {
		text := strings.TrimSpace(line.Code)
		if text == "" {
			return "", sourcepos.Range{}
		}
		return text, line.Range()
	}

	if skip >= len(line.Tokens) {
		return "", sourcepos.Range{}
	}

	start := int(line.Tokens[skip].StartColumn) - 1
	end := int(line.Tokens[len(line.Tokens)-1].EndColumn) - 1
	if start < 0 || start >= len(line.Code) || end <= start {
		return "", sourcepos.Range{}
	}

	return strings.TrimSpace(line.Code[start:end]), tokenRange(line, skip, len(line.Tokens))
}

func tokenRange(line TokenLine, start, end int) sourcepos.Range {
	if start < 0 {
		start = 0
	}
	if end > len(line.Tokens) {
		end = len(line.Tokens)
	}
	if start >= end || len(line.Tokens) == 0 {
		return line.Range()
	}

	return sourcepos.NewRange(
		sourcepos.NewPosition(line.Line, int(line.Tokens[start].StartColumn)),
		sourcepos.NewPosition(line.Line, int(line.Tokens[end-1].EndColumn)),
	)
}

func missingColumn(tokens []Token, index int) int {
	if index < len(tokens) {
		return int(tokens[index].StartColumn)
	}
	if len(tokens) > 0 {
		return int(tokens[len(tokens)-1].EndColumn)
	}

	return 1
}

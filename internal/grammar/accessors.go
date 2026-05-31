package grammar

import (
	"slices"
	"strings"
	"sync"
)

// CompiledSpec is the runtime view of Spec: precomputed lookup tables and
// derived vocabularies that parser, formatter, LSP, and exporters can share.
type CompiledSpec struct {
	Spec                   Spec
	TopLevelKeywords       []string
	DeclarationByKeyword   map[string]Decl
	DeclarationByKind      map[DeclarationKind]Decl
	requirementClauseRules []RequirementClauseRule
	clauseKindByKeyword    map[string]RequirementClauseKind
	clauseRankByKind       map[RequirementClauseKind]int
	lineRulesByClause      map[RequirementClauseKind][]LineRule
	AssignmentEntryRule    *LineRule
	MetadataKeywords       []string
	ReservedKeywords       []string
	PunctuationTokens      []string
	ActionSpecs            []ActionSpec
	ActionSpecByVerb       map[string]ActionSpec
	ActionVerbs            []string
	WhenVerbs              []string
	actionVerbSet          map[string]struct{}
	whenVerbSet            map[string]struct{}
}

// Compiled is the singleton runtime grammar used by parser, formatter, and LSP.
var Compiled = Compile(Grammar)

// Compile turns the declarative grammar into cheap lookup tables. Runtime
// packages should query Compiled instead of walking Grammar directly.
func Compile(spec Spec) CompiledSpec {
	compiled := CompiledSpec{
		Spec:                 spec,
		DeclarationByKeyword: make(map[string]Decl, len(spec.Declarations)),
		DeclarationByKind:    make(map[DeclarationKind]Decl, len(spec.Declarations)),
		clauseKindByKeyword:  map[string]RequirementClauseKind{},
		clauseRankByKind:     make(map[RequirementClauseKind]int),
		lineRulesByClause:    make(map[RequirementClauseKind][]LineRule),
		ActionSpecByVerb:     map[string]ActionSpec{},
		ActionSpecs:          make([]ActionSpec, 0, len(actionFamilies)),
		ActionVerbs:          make([]string, 0, len(actionFamilies)*6),
	}

	compiled.TopLevelKeywords = make([]string, 0, len(spec.Declarations))
	var rawClauseRules []LineRule

	for _, decl := range spec.Declarations {
		compiled.TopLevelKeywords = append(compiled.TopLevelKeywords, decl.Keyword)
		compiled.DeclarationByKeyword[decl.Keyword] = decl
		compiled.DeclarationByKind[decl.Kind] = decl

		if decl.Body != nil {
			for i := range decl.Body.Lines {
				line := &decl.Body.Lines[i]
				if line.ClauseKind != "" {
					rawClauseRules = append(rawClauseRules, *line)
					compiled.lineRulesByClause[line.ClauseKind] = append(compiled.lineRulesByClause[line.ClauseKind], *line)
				}
				if compiled.AssignmentEntryRule == nil && len(line.Semantics) > 0 {
					compiled.AssignmentEntryRule = line
				}
			}
		}
	}

	compiled.requirementClauseRules = dedupeRequirementClauseRules(rawClauseRules)
	for idx, rule := range compiled.requirementClauseRules {
		compiled.clauseRankByKind[rule.Kind] = idx
		compiled.clauseKindByKeyword[rule.Keyword] = rule.Kind
	}

	compiled.MetadataKeywords = clauseKeywordsByKinds(spec, RequirementClausePriority, RequirementClauseRetention, RequirementClauseAccess, RequirementClauseLinkedTo, RequirementClauseStakeholders)
	compiled.WhenVerbs = slices.Clone(whenVerbs)
	compiled.whenVerbSet = stringSet(compiled.WhenVerbs)

	for _, f := range actionFamilies {
		form := ActionFormObjectOfTarget
		if f.DirectTarget {
			form = ActionFormDirectTarget
		}

		spec := ActionSpec{
			Kind:            f.Kind,
			Form:            form,
			AllowsMechanism: f.AllowsMechanism,
			Description:     f.Description,
			Verbs:           slices.Clone(f.Verbs),
		}

		compiled.ActionSpecs = append(compiled.ActionSpecs, spec)
		compiled.ActionVerbs = append(compiled.ActionVerbs, f.Verbs...)
		for _, verb := range f.Verbs {
			compiled.ActionSpecByVerb[verb] = spec
		}
	}

	slices.Sort(compiled.ActionVerbs)
	compiled.actionVerbSet = stringSet(compiled.ActionVerbs)

	// Compile literal sets: keywords, token sets, separators — one pass.
	{
		kws := map[string]struct{}{}
		tokSets := map[string]struct{}{}
		seps := map[string]struct{}{}
		for _, decl := range spec.Declarations {
			collectFromPattern(decl.Header, kws, tokSets, seps)
			if decl.Body != nil {
				for _, line := range decl.Body.Lines {
					collectFromPattern(line.Pattern, kws, tokSets, seps)
				}
			}
		}

		resSet := map[string]struct{}{}
		punctSet := map[string]struct{}{}
		for k := range kws {
			if isWordKeyword(k) {
				resSet[k] = struct{}{}
			} else {
				punctSet[k] = struct{}{}
			}
		}

		for k := range tokSets {
			resSet[k] = struct{}{}
		}

		for k := range seps {
			if !isWordKeyword(k) {
				punctSet[k] = struct{}{}
			}
		}

		compiled.ReservedKeywords = sortedKeys(resSet)
		compiled.PunctuationTokens = sortedKeys(punctSet)
	}

	slices.SortFunc(compiled.PunctuationTokens, func(a, b string) int {
		if len(a) != len(b) {
			return len(b) - len(a)
		}
		return strings.Compare(a, b)
	})

	return compiled
}

// RequirementClauseRule is the compiled scheduling metadata for a clause kind.
type RequirementClauseRule struct {
	Kind       RequirementClauseKind
	Keyword    string
	Optional   bool
	Repeatable bool
}

func dedupeRequirementClauseRules(rules []LineRule) []RequirementClauseRule {
	seen := map[RequirementClauseKind]*RequirementClauseRule{}
	order := make([]RequirementClauseKind, 0, len(rules))

	for _, line := range rules {
		if entry, ok := seen[line.ClauseKind]; ok {
			if line.Required {
				entry.Optional = false
			}
			if line.Repeatable {
				entry.Repeatable = true
			}
			continue
		}

		order = append(order, line.ClauseKind)
		seen[line.ClauseKind] = &RequirementClauseRule{
			Kind: line.ClauseKind, Keyword: LeadingKeyword(line.Pattern),
			Optional: !line.Required, Repeatable: line.Repeatable,
		}
	}

	out := make([]RequirementClauseRule, 0, len(order))
	for _, k := range order {
		out = append(out, *seen[k])
	}
	return out
}

func (c CompiledSpec) DeclarationKindForLine(line string) (DeclarationKind, bool) {
	trimmed := strings.TrimSpace(line)
	if i := strings.IndexAny(trimmed, " \t"); i >= 0 {
		trimmed = trimmed[:i]
	}
	if trimmed == "" {
		return "", false
	}

	decl, ok := c.DeclarationByKeyword[trimmed]
	return decl.Kind, ok
}

func (c CompiledSpec) DeclarationKindForTokens(tokens []Token) (DeclarationKind, bool) {
	if len(tokens) == 0 {
		return "", false
	}

	decl, ok := c.DeclarationByKeyword[tokens[0].Text]
	return decl.Kind, ok
}

func (c CompiledSpec) RequirementClauseKindForLine(line string) (RequirementClauseKind, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", false
	}

	for _, rule := range c.requirementClauseRules {
		if trimmed == rule.Keyword || strings.HasPrefix(trimmed, rule.Keyword+" ") {
			return c.clauseKindByKeyword[rule.Keyword], true
		}
	}
	return "", false
}

func (c CompiledSpec) NextRequirementClauseKeywords(seen []RequirementClauseKind) []string {
	counts := map[RequirementClauseKind]int{}
	lastIdx, haveSeen := 0, false
	for _, k := range seen {
		idx, ok := c.clauseRankByKind[k]
		if !ok {
			continue
		}
		counts[k]++
		if !haveSeen || idx > lastIdx {
			lastIdx, haveSeen = idx, true
		}
	}

	start := 0
	if haveSeen {
		start = lastIdx
	}

	out := make([]string, 0, len(c.requirementClauseRules)-start)
	for i := start; i < len(c.requirementClauseRules); i++ {
		r := c.requirementClauseRules[i]
		if counts[r.Kind] > 0 && !r.Repeatable {
			continue
		}
		out = append(out, r.Keyword)
		if !r.Optional {
			break
		}
	}
	return out
}

func clauseKeywordsByKinds(spec Spec, kinds ...RequirementClauseKind) []string {
	set := map[RequirementClauseKind]struct{}{}
	for _, k := range kinds {
		set[k] = struct{}{}
	}

	var out []string
	for _, decl := range spec.Declarations {
		if decl.Body == nil {
			continue
		}
		for _, line := range decl.Body.Lines {
			if _, ok := set[line.ClauseKind]; !ok {
				continue
			}
			if kw := LeadingKeyword(line.Pattern); kw != "" && !slices.Contains(out, kw) {
				out = append(out, kw)
			}
		}
	}

	return out
}

func collectFromPattern(pattern []Matcher, kws, tokSets, separators map[string]struct{}) {
	for _, m := range pattern {
		switch v := m.(type) {
		case Kw:
			if kws != nil {
				kws[v.Text] = struct{}{}
			}
		case Tok:
			if tokSets != nil {
				for _, w := range v.Set {
					tokSets[w] = struct{}{}
				}
			}
		case List:
			if separators != nil && v.Separator != "" {
				separators[v.Separator] = struct{}{}
			}
		case Opt:
			collectFromPattern(v.Inner, kws, tokSets, separators)
		}
	}
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	slices.Sort(out)
	return out
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func isWordKeyword(s string) bool {
	if s == "" || !isIdentStart(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !isIdentStart(c) && (c < '0' || c > '9') && c != '-' && c != ' ' {
			return false
		}
	}
	return true
}

func (c CompiledSpec) IsValidNumber(value string) bool {
	if value == "" {
		return false
	}
	if value[0] == '-' {
		value = value[1:]
	}
	if value == "" {
		return false
	}

	i, digits := 0, 0
	for i < len(value) && value[i] >= '0' && value[i] <= '9' {
		i++
		digits++
	}

	if i == len(value) {
		return digits > 0
	}
	if value[i] != '.' {
		return false
	}

	i++
	frac := 0
	for i < len(value) && value[i] >= '0' && value[i] <= '9' {
		i++
		frac++
	}

	return i == len(value) && frac > 0
}

func (c CompiledSpec) IsValidIdentifier(value string) bool {
	if value == "" || !isIdentStart(value[0]) {
		return false
	}

	for i := 1; i < len(value); i++ {
		if !IsIdentifierChar(value[i]) {
			return false
		}
	}
	return true
}

func IsIdentifierChar(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9') || c == '-'
}

func isIdentStart(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' }

func (c CompiledSpec) BodyLineRule(decl Decl, tokens []Token) (LineRule, bool) {
	if decl.Body == nil || len(tokens) == 0 {
		return LineRule{}, false
	}

	var fallback, best *LineRule
	bestLen := 0
	for i := range decl.Body.Lines {
		rule := &decl.Body.Lines[i]
		n := leadingKeywordLen(rule.Pattern, tokens)
		if n == 0 {
			if LeadingKeyword(rule.Pattern) == "" && fallback == nil {
				fallback = rule
			}
			continue
		}
		if n > bestLen {
			best, bestLen = rule, n
		}
	}

	if best == nil {
		if fallback == nil {
			return LineRule{}, false
		}
		return *fallback, true
	}

	if best.ClauseKind == "" {
		return *best, true
	}
	return c.clauseRule(best.ClauseKind, tokens)
}

func (c CompiledSpec) clauseRule(kind RequirementClauseKind, tokens []Token) (LineRule, bool) {
	var first *LineRule
	body := tokens
	rules := c.lineRulesByClause[kind]

	for i := range rules {
		rule := &rules[i]
		if first == nil {
			first = rule
			body = tokens[leadingKeywordLen(rule.Pattern, tokens):]
		}
		if kind != RequirementClauseSystemShall {
			return *rule, true
		}
		if len(body) > 0 && actionVerb(rule.Pattern, body[0].Text) {
			return *rule, true
		}
	}
	if first == nil {
		return LineRule{}, false
	}

	wantOf := containsToken(body, "of")
	for _, rule := range rules {
		if patternHasKeyword(rule.Pattern, "of") == wantOf {
			return rule, true
		}
	}

	return *first, true
}

// DiagnosticPrefix is the stable base for grammar-derived diagnostics.
func (c CompiledSpec) DiagnosticPrefix(decl DeclarationKind, rule LineRule) string {
	return string(decl) + "." + LineRuleKey(rule)
}

// DiagnosticCode appends normalized parts to a grammar diagnostic prefix.
func (c CompiledSpec) DiagnosticCode(prefix string, parts ...string) string {
	suffix := strings.Join(parts, ".")
	if prefix == "" {
		return suffix
	}
	if suffix == "" {
		return prefix
	}
	return prefix + "." + suffix
}

func diagnosticPart(text string) string {
	switch text {
	case ",":
		return "comma"
	case "=":
		return "equals"
	case "->":
		return "arrow"
	default:
		return strings.ReplaceAll(text, " ", "_")
	}
}

func (r LineRule) SemanticField(role SemanticRole) (string, bool) {
	for _, b := range r.Semantics {
		if b.Role == role {
			return b.Field, true
		}
	}
	return "", false
}

type ActionForm string

const (
	ActionFormObjectOfTarget ActionForm = "object_of_target"
	ActionFormDirectTarget   ActionForm = "direct_target"
)

type ActionSpec struct {
	Kind            ActionKind
	Form            ActionForm
	AllowsMechanism bool
	Description     string
	Verbs           []string
}

func (a ActionSpec) Template() string {
	tmpl := "<verb> <object> of <target>"
	if a.Form == ActionFormDirectTarget {
		tmpl = "<verb> <target>"
	}
	if a.AllowsMechanism {
		tmpl += " [using <mechanism>]"
	}

	return tmpl
}

func LineRuleNodeName(rule LineRule) string {
	switch rule.ClauseKind {
	case RequirementClauseSystemShall:
		return "action_clause"
	case "":
		return LineRuleKey(rule)
	default:
		return string(rule.ClauseKind) + "_clause"
	}
}

func LineRuleKey(rule LineRule) string {
	if rule.ClauseKind != "" {
		return string(rule.ClauseKind)
	}

	var buf strings.Builder
	for i, c := range rule.Name {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				buf.WriteByte('_')
			}
			buf.WriteRune(c + ('a' - 'A'))
		} else {
			buf.WriteRune(c)
		}
	}

	return buf.String()
}

func LeadingKeyword(pattern []Matcher) string {
	var parts []string
	for _, m := range pattern {
		if kw, ok := m.(Kw); ok {
			parts = append(parts, kw.Text)
		} else {
			break
		}
	}

	return strings.Join(parts, " ")
}

func leadingKeywordLen(pattern []Matcher, tokens []Token) int {
	i := 0

	for _, m := range pattern {
		kw, ok := m.(Kw)
		if !ok {
			break
		}

		for word := range strings.FieldsSeq(kw.Text) {
			if i >= len(tokens) || tokens[i].Text != word {
				return 0
			}
			i++
		}
	}

	return i
}

func containsToken(tokens []Token, value string) bool {
	for _, token := range tokens {
		if token.Text == value {
			return true
		}
	}

	return false
}

func actionVerb(pattern []Matcher, verb string) bool {
	for _, m := range pattern {
		tok, ok := m.(Tok)
		if ok && tok.Kind == TokKindActionVerb {
			return slices.Contains(tok.Set, verb)
		}
	}

	return false
}

func patternHasKeyword(pattern []Matcher, keyword string) bool {
	for _, m := range pattern {
		if kw, ok := m.(Kw); ok && slices.Contains(strings.Fields(kw.Text), keyword) {
			return true
		}
	}

	return false
}

// Token registry — lazy-initialised on first query.

type TokenCategory string

const (
	TokenCategoryKeyword  TokenCategory = "keyword"
	TokenCategoryClause   TokenCategory = "clause"
	TokenCategoryVerb     TokenCategory = "verb"
	TokenCategoryMetadata TokenCategory = "metadata"
)

type TokenInfo struct {
	Text          string
	Category      TokenCategory
	Documentation string
}

var (
	tokenInfoOnce   sync.Once
	tokenInfoByText map[string]TokenInfo
)

func TokenInfoFor(text string) (TokenInfo, bool) {
	tokenInfoOnce.Do(initTokenInfo)
	info, ok := tokenInfoByText[text]
	return info, ok
}

func exportedTokenInfo() []TokenInfo {
	tokenInfoOnce.Do(initTokenInfo)
	out := make([]TokenInfo, 0, len(tokenInfoByText))
	for _, keyword := range Compiled.ReservedKeywords {
		if info, ok := tokenInfoByText[keyword]; ok {
			out = append(out, info)
		}
	}
	return out
}

var topLevelKeywordDocs = map[string]string{
	"stakeholder": "Declares a stakeholder that can be referenced in requirements and assignments.",
	"value":       "Declares a custom value with explicit geometry: `value <name> = <rad angle>, <radius>`.",
	"requirement": "Starts a requirement block with EARS context clauses, one `system shall` action, and follow-up clauses.",
	"assignment":  "Starts an assignment block: `assignment <RequirementID>` followed by `<Stakeholder> -> <Value>` lines.",
}

var clauseKeywordDocs = map[string]string{
	"while": "EARS state-driven clause: `while <actor> <state/condition>`.",
	"when":  "EARS event-driven clause: `when <actor> <verb> <target>`.",
	"if":    "EARS unwanted-behavior or conditional clause: `if <condition>`.",
	"where": "EARS optional-feature clause: `where <feature condition>`.",
}

var actionKeywordDocs = map[string]string{
	"system": "Introduces the system response line of a requirement: `system shall ...`.",
	"shall":  "Introduces the system response line of a requirement: `system shall ...`.",
	"of":     "Separates the action object from the action target in targeted actions.",
	"using":  "Introduces the mechanism used by an action when that action kind allows a mechanism.",
}

var metadataKeywordDocs = map[string]string{
	"priority":     "Optional requirement metadata describing importance or urgency.",
	"retention":    "Optional requirement metadata describing how long related data should be kept.",
	"access":       "Optional requirement metadata describing who may access the related information or capability.",
	"linked_to":    "Optional traceability clause linking the requirement to another artifact or identifier.",
	"stakeholders": "Lists the declared stakeholders affected by the requirement.",
}

const whenVerbDocumentation = "Allowed event verb in a `when <actor> <verb> <target>` clause."

func initTokenInfo() {
	tokenInfoByText = make(map[string]TokenInfo)

	registerTokens(Compiled.TopLevelKeywords, TokenCategoryKeyword, topLevelKeywordDocs)
	for kw, doc := range clauseKeywordDocs {
		registerToken(kw, TokenCategoryClause, doc)
	}
	for kw, doc := range actionKeywordDocs {
		registerToken(kw, TokenCategoryKeyword, doc)
	}
	registerTokens(Compiled.MetadataKeywords, TokenCategoryMetadata, metadataKeywordDocs)

	for _, spec := range Compiled.ActionSpecs {
		for _, verb := range spec.Verbs {
			registerToken(verb, TokenCategoryVerb, actionVerbDoc(spec))
		}
	}
	for _, verb := range Compiled.WhenVerbs {
		registerToken(verb, TokenCategoryVerb, whenVerbDocumentation)
	}
}

func registerToken(token string, category TokenCategory, documentation string) {
	tokenInfoByText[token] = TokenInfo{Text: token, Category: category, Documentation: documentation}
}

func registerTokens(tokens []string, category TokenCategory, docs map[string]string) {
	for _, token := range tokens {
		registerToken(token, category, docs[token])
	}
}

func actionVerbDoc(spec ActionSpec) string {
	return spec.Description + "\n\nCanonical form: `" + spec.Template() + "`."
}

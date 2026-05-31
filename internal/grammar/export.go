package grammar

import (
	"encoding/json"

	"github.com/brogergvhs/value-dsl/internal/values"
)

const ExportVersion = "2"

type ExportedSpec struct {
	Version           string                  `json:"version"`
	IdentifierPattern string                  `json:"identifier_pattern"`
	NumberPattern     string                  `json:"number_pattern"`
	StringPattern     string                  `json:"string_pattern"`
	LineCommentPrefix string                  `json:"line_comment_prefix"`
	ReservedKeywords  []string                `json:"reserved_keywords"`
	PunctuationTokens []string                `json:"punctuation_tokens"`
	TokenKindRules    []ExportedTokenKindRule `json:"token_kind_rules"`
	ActionSystem      *ExportedActionSystem   `json:"action_system,omitempty"`
	Declarations      []ExportedDecl          `json:"declarations"`
	Tokens            []ExportedTokenInfo     `json:"tokens"`
	Actions           []ExportedActionSpec    `json:"actions"`
	ValueCategories   []ExportedValueCategory `json:"value_categories"`
}

type ExportedDecl struct {
	Kind     string            `json:"kind"`
	NodeName string            `json:"node_name"`
	Keyword  string            `json:"keyword"`
	Header   []ExportedMatcher `json:"header"`
	Body     *ExportedBlock    `json:"body,omitempty"`
}

type ExportedBlock struct {
	Lines []ExportedLineRule `json:"lines"`
}

type ExportedLineRule struct {
	Name       string            `json:"name"`
	NodeName   string            `json:"node_name"`
	ClauseKind string            `json:"clause_kind,omitempty"`
	Required   bool              `json:"required"`
	Repeatable bool              `json:"repeatable"`
	Order      int               `json:"order"`
	Pattern    []ExportedMatcher `json:"pattern"`
}

type ExportedMatcher struct {
	Type      string            `json:"type"`
	Name      string            `json:"name,omitempty"`
	Text      string            `json:"text,omitempty"`
	Kind      string            `json:"kind,omitempty"`
	Values    []string          `json:"values,omitempty"`
	Separator string            `json:"separator,omitempty"`
	Item      *ExportedMatcher  `json:"item,omitempty"`
	Inner     []ExportedMatcher `json:"inner,omitempty"`
}

type ExportedTokenInfo struct {
	Text          string        `json:"text"`
	Category      TokenCategory `json:"category"`
	Documentation string        `json:"documentation,omitempty"`
}

type ExportedTokenKindRule struct {
	Kind     string   `json:"kind"`
	Rule     string   `json:"rule"`
	Fallback string   `json:"fallback,omitempty"`
	Values   []string `json:"values,omitempty"`
}

type ExportedActionSystem struct {
	ClauseNodeName string               `json:"clause_node_name"`
	ExprNodeName   string               `json:"expr_node_name"`
	Prefix         []ExportedMatcher    `json:"prefix"`
	Forms          []ExportedActionForm `json:"forms"`
}

type ExportedActionForm struct {
	Form     string            `json:"form"`
	NodeName string            `json:"node_name"`
	Pattern  []ExportedMatcher `json:"pattern"`
}

type ExportedActionSpec struct {
	Kind            string   `json:"kind"`
	Form            string   `json:"form"`
	AllowsMechanism bool     `json:"allows_mechanism"`
	Description     string   `json:"description"`
	Template        string   `json:"template"`
	Verbs           []string `json:"verbs"`
}

type ExportedValueCategory struct {
	Name        string  `json:"name"`
	Angle       float64 `json:"angle"`
	Description string  `json:"description"`
}

func Export() ExportedSpec {
	return ExportedSpec{
		Version:           ExportVersion,
		IdentifierPattern: Compiled.Spec.Identifier,
		NumberPattern:     Compiled.Spec.NumberPattern,
		StringPattern:     Compiled.Spec.StringPattern,
		LineCommentPrefix: Compiled.Spec.LineComment,
		ReservedKeywords:  Compiled.ReservedKeywords,
		PunctuationTokens: Compiled.PunctuationTokens,
		TokenKindRules: []ExportedTokenKindRule{
			{Kind: string(TokKindIdent), Rule: "identifier"},
			{Kind: string(TokKindNumber), Rule: "number"},
			{Kind: string(TokKindString), Rule: "string_literal"},
			{Kind: string(TokKindWhenVerb), Rule: "when_verb", Fallback: "identifier", Values: Compiled.WhenVerbs},
			{Kind: string(TokKindActionVerb), Rule: "identifier"},
		},
		ActionSystem: exportActionSystem(),
		Declarations: exportDeclarations(),
		Tokens:       mapSlice(exportedTokenInfo(), func(t TokenInfo) ExportedTokenInfo { return ExportedTokenInfo(t) }),
		Actions: mapSlice(Compiled.ActionSpecs, func(s ActionSpec) ExportedActionSpec {
			return ExportedActionSpec{Kind: string(s.Kind), Form: string(s.Form), AllowsMechanism: s.AllowsMechanism, Description: s.Description, Template: s.Template(), Verbs: s.Verbs}
		}),
		ValueCategories: mapSlice(values.Categories(), func(c values.Category) ExportedValueCategory {
			return ExportedValueCategory{Name: c.Name, Angle: c.Angle, Description: c.Description}
		}),
	}
}

func ExportJSON() ([]byte, error) {
	return json.MarshalIndent(Export(), "", "  ")
}

func exportActionSystem() *ExportedActionSystem {
	if len(actionFamilies) == 0 {
		return nil
	}

	// Compute the shared leading Kw prefix from the first family's pattern.
	prefix := func() []Matcher {
		p := shallPattern(actionFamilies[0])
		for i, m := range p {
			if _, ok := m.(Kw); !ok {
				return p[:i]
			}
		}

		return p
	}()

	// One exported form per distinct (DirectTarget) value; prefer AllowsMechanism.
	seenForms := map[ActionForm]struct{}{}
	var exported []ExportedActionForm
	for _, f := range actionFamilies {
		form := ActionFormObjectOfTarget
		if f.DirectTarget {
			form = ActionFormDirectTarget
		}

		if _, ok := seenForms[form]; ok {
			continue
		}

		seenForms[form] = struct{}{}
		var best *actionFamily

		for i := range actionFamilies {
			cf := &actionFamilies[i]
			if cf.DirectTarget == f.DirectTarget && (best == nil || (cf.AllowsMechanism && !best.AllowsMechanism)) {
				best = cf
			}
		}

		var pattern []Matcher
		if best != nil {
			full := shallPattern(*best)
			if len(full) > len(prefix) {
				pattern = full[len(prefix):]
			}
		}

		nodeName := "targeted_action"
		if form == ActionFormDirectTarget {
			nodeName = "direct_action"
		}
		exported = append(exported, ExportedActionForm{
			Form:     string(form),
			NodeName: nodeName,
			Pattern:  exportMatchers(pattern),
		})
	}

	return &ExportedActionSystem{
		ClauseNodeName: "action_clause",
		ExprNodeName:   "action_expression",
		Prefix:         exportMatchers(prefix),
		Forms:          exported,
	}
}

func exportDeclarations() []ExportedDecl {
	declarations := Compiled.Spec.Declarations
	out := make([]ExportedDecl, 0, len(declarations))

	for _, decl := range declarations {
		var body *ExportedBlock
		if decl.Body != nil {
			body = &ExportedBlock{Lines: mapSlice(decl.Body.Lines, func(r LineRule) ExportedLineRule {
				return ExportedLineRule{Name: r.Name, NodeName: LineRuleNodeName(r), ClauseKind: string(r.ClauseKind), Required: r.Required, Repeatable: r.Repeatable, Order: r.Order, Pattern: exportMatchers(r.Pattern)}
			})}
		}

		out = append(out, ExportedDecl{
			Kind:     string(decl.Kind),
			NodeName: string(decl.Kind) + "_decl",
			Keyword:  decl.Keyword,
			Header:   exportMatchers(decl.Header),
			Body:     body,
		})
	}

	return out
}

func exportMatchers(in []Matcher) []ExportedMatcher {
	out := make([]ExportedMatcher, 0, len(in))
	for _, m := range in {
		switch v := m.(type) {
		case Kw:
			out = append(out, ExportedMatcher{Type: "kw", Text: v.Text})
		case Ident:
			out = append(out, ExportedMatcher{Type: "ident", Name: v.Name})
		case Tok:
			out = append(out, ExportedMatcher{Type: "tok", Name: v.Name, Kind: string(v.Kind), Values: v.Set})
		case Ref:
			out = append(out, ExportedMatcher{Type: "ref", Name: v.Name, Kind: string(v.Kind)})
		case Enum:
			out = append(out, ExportedMatcher{Type: "enum", Name: v.Name, Values: v.Values})
		case Str:
			out = append(out, ExportedMatcher{Type: "str", Name: v.Name})
		case List:
			item := exportMatchers([]Matcher{v.Item})[0]
			out = append(out, ExportedMatcher{Type: "list", Name: v.Name, Separator: v.Separator, Item: &item})
		case Opt:
			out = append(out, ExportedMatcher{Type: "opt", Inner: exportMatchers(v.Inner)})
		case Free:
			out = append(out, ExportedMatcher{Type: "free", Name: v.Name})
		}
	}

	return out
}

func mapSlice[A, B any](in []A, f func(A) B) []B {
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

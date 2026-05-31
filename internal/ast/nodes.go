// Package ast defines syntax-level nodes produced by the DSL parser
package ast

import (
	"sort"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
)

const (
	FieldRawText = "_raw"
	FieldKeyword = "_keyword"
	FieldExtra   = "_extra"
)

type FieldValue struct {
	Text    string
	Range   sourcepos.Range
	RefKind grammar.DeclarationKind
	HasRef  bool
}

type GenericNode struct {
	Kind   string
	Line   int
	Range  sourcepos.Range
	Fields map[string][]FieldValue
}

func (n GenericNode) First(name string) string {
	if v := n.Fields[name]; len(v) > 0 {
		return v[0].Text
	}
	return ""
}

func (n GenericNode) Values(name string) []FieldValue {
	return n.Fields[name]
}

func (n GenericNode) Texts(name string) []string {
	values := n.Fields[name]
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.Text)
	}
	return out
}

func (n GenericNode) Joined(name, sep string) string {
	if values := n.Texts(name); len(values) > 0 {
		return strings.Join(values, sep)
	}
	return ""
}

func (n GenericNode) HasText(name, text string) bool {
	for _, value := range n.Fields[name] {
		if value.Text == text {
			return true
		}
	}
	return false
}

func (n GenericNode) FirstNonMetaField() string {
	names := make([]string, 0, len(n.Fields))
	for name := range n.Fields {
		if name != FieldRawText && name != FieldKeyword && name != FieldExtra && len(n.Fields[name]) > 0 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) > 0 {
		return n.Fields[names[0]][0].Text
	}
	return ""
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// DocumentNode is the root AST node for a DSL specification
type DocumentNode struct {
	Declarations []DeclarationNode
	Malformed    []MalformedLineNode
}

// DeclarationsOfKind returns all declarations matching the given kind.
func (d *DocumentNode) DeclarationsOfKind(kind grammar.DeclarationKind) []DeclarationNode {
	if d == nil {
		return nil
	}

	var out []DeclarationNode
	for _, decl := range d.Declarations {
		if decl.Kind == kind {
			out = append(out, decl)
		}
	}

	return out
}

// DeclarationNode represents a parsed declaration block of any kind.
type DeclarationNode struct {
	Kind      grammar.DeclarationKind
	Line      int
	Range     sourcepos.Range
	Header    GenericNode
	Body      []GenericNode
	Malformed []MalformedLineNode
}

// MalformedLineNode preserves a source line that could be structurally classified
// but not fully parsed into a known AST construct.
type MalformedLineNode struct {
	Line    int
	Range   sourcepos.Range
	Kind    string
	Keyword string
	RawText string
	Reason  string
}

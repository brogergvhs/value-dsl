package completion

import (
	"fmt"
	"sort"

	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/values"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func keywordItems(keywords []string) []protocol.CompletionItem {
	return simpleItems(keywords, protocol.CompletionItemKindKeyword, "keyword")
}

func identifierItems(index *docindex.Document) []protocol.CompletionItem {
	if index == nil {
		return nil
	}
	unique := map[string]protocol.CompletionItem{}
	for _, kind := range []grammar.DeclarationKind{grammar.DeclarationKindRequirement, grammar.DeclarationKindStakeholder, grammar.DeclarationKindValue} {
		for _, name := range index.Names(kind) {
			itemKind := protocol.CompletionItemKindReference
			detail := string(kind)
			switch kind {
			case grammar.DeclarationKindStakeholder:
				itemKind = protocol.CompletionItemKindVariable
			case grammar.DeclarationKindValue:
				definition, ok := values.LookupByName(name)
				if !ok {
					definition = values.Definition{Name: name}
				}
				unique[name] = makeValueItem(definition)
				continue
			}
			unique[name] = makeItem(name, itemKind, detail)
		}
	}
	return sortedItems(unique)
}

func valueItems(defs map[string]values.Definition) []protocol.CompletionItem {
	m := make(map[string]protocol.CompletionItem, len(defs))
	for n, def := range defs {
		m[n] = makeValueItem(def)
	}
	return sortedItems(m)
}

func sortedItems(m map[string]protocol.CompletionItem) []protocol.CompletionItem {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	items := make([]protocol.CompletionItem, 0, len(names))
	for _, n := range names {
		items = append(items, m[n])
	}
	return items
}

func simpleItems(vals []string, kind protocol.CompletionItemKind, detail string) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, len(vals))
	for _, v := range vals {
		items = append(items, makeItem(v, kind, detail))
	}
	return items
}

func makeItem(label string, kind protocol.CompletionItemKind, detail string) protocol.CompletionItem {
	return protocol.CompletionItem{
		Label:  label,
		Kind:   &kind,
		Detail: &detail,
	}
}

func makeValueItem(def values.Definition) protocol.CompletionItem {
	detail := fmt.Sprintf("value (%s)", def.Category)
	item := makeItem(def.Name, protocol.CompletionItemKindValue, detail)
	item.Documentation = protocol.MarkupContent{Kind: protocol.MarkupKindMarkdown, Value: formatValueDocumentation(def)}
	return item
}

func appendUniqueItems(base []protocol.CompletionItem, extras ...protocol.CompletionItem) []protocol.CompletionItem {
	seen := make(map[string]struct{}, len(base))
	for _, item := range base {
		seen[item.Label] = struct{}{}
	}
	for _, item := range extras {
		if _, ok := seen[item.Label]; !ok {
			base = append(base, item)
			seen[item.Label] = struct{}{}
		}
	}
	return base
}

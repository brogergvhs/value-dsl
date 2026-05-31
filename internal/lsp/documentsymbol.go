package lsp

import (
	"fmt"
	"sort"
	"strings"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func buildDocumentSymbols(result coreanalysis.Result) []protocol.DocumentSymbol {
	if result.Index == nil || result.Index.Model == nil {
		return nil
	}

	return buildOutlineSymbols(result.Index.Lines, result.Index.Model.Stakeholders, result.Index.Model.Values, result.Index.Model.Requirements)
}

func newRangeSymbol(lines []grammar.TokenLine, name string, kind protocol.SymbolKind, sourceRange sourcepos.Range, detail *string, children []protocol.DocumentSymbol) protocol.DocumentSymbol {
	r := toProtocolRangeInLines(lines, sourceRange)
	symbol := protocol.DocumentSymbol{Name: name, Kind: kind, Range: r, SelectionRange: r}

	if detail != nil {
		symbol.Detail = detail
	}
	if len(children) > 0 {
		symbol.Children = children
	}

	return symbol
}

func buildOutlineSymbols(lines []grammar.TokenLine, stakeholders []model.Stakeholder, values []model.Value, requirements []model.Requirement) []protocol.DocumentSymbol {
	items := make([]protocol.DocumentSymbol, 0, len(stakeholders)+len(values)+len(requirements))

	for _, stakeholder := range stakeholders {
		items = append(items, newRangeSymbol(lines, stakeholder.Name, protocol.SymbolKindVariable, stakeholder.Range, nil, nil))
	}
	for _, value := range values {
		detail := fmt.Sprintf("angle=%0.2f, radius=%0.2f", value.Angle, value.Radius)
		items = append(items, newRangeSymbol(lines, value.Name, protocol.SymbolKindConstant, value.Range, &detail, nil))
	}
	for _, requirement := range requirements {
		items = append(items, requirementDocumentSymbol(lines, requirement))
	}

	sortSymbols(items)
	return items
}

func requirementDocumentSymbol(lines []grammar.TokenLine, requirement model.Requirement) protocol.DocumentSymbol {
	children := make([]protocol.DocumentSymbol, 0, len(requirement.Context.Clauses)+4)
	for _, clause := range requirement.Context.Clauses {
		detail := clause.RawText
		children = append(children, newRangeSymbol(lines, string(clause.Type), protocol.SymbolKindEvent, clause.Range, &detail, nil))
	}

	if requirement.Action.RawText != "" {
		detail := requirement.Action.RawText
		children = append(children, newRangeSymbol(lines, "system shall", protocol.SymbolKindFunction, requirement.Action.Range, &detail, nil))
	}
	if len(requirement.Stakeholders) > 0 {
		names := make([]string, 0, len(requirement.Stakeholders))
		for _, s := range requirement.Stakeholders {
			names = append(names, s.Name)
		}
		detail := strings.Join(names, ", ")
		children = append(children, newRangeSymbol(lines, "stakeholders", protocol.SymbolKindArray, requirement.StakeholdersRange, &detail, nil))
	}

	if detail, metadataRange, ok := requirementMetadataDetail(requirement); ok {
		children = append(children, newRangeSymbol(lines, "metadata", protocol.SymbolKindStruct, metadataRange, &detail, nil))
	}
	if detail, traceabilityRange, ok := requirementTraceabilityDetail(requirement); ok {
		children = append(children, newRangeSymbol(lines, "linked_to", protocol.SymbolKindKey, traceabilityRange, &detail, nil))
	}

	sortSymbols(children)
	detail := requirement.Action.RawText
	return newRangeSymbol(lines, requirement.ID, protocol.SymbolKindObject, requirement.Range, &detail, children)
}

func requirementMetadataDetail(requirement model.Requirement) (string, sourcepos.Range, bool) {
	var parts []string
	var rng sourcepos.Range
	meta := requirement.Metadata

	if meta.Priority.Value != "" {
		parts = append(parts, "priority="+meta.Priority.Value)
		rng = firstNonZeroRange(rng, meta.Priority.Range)
	}
	if meta.Retention.Value != "" {
		parts = append(parts, "retention="+meta.Retention.Value)
		rng = firstNonZeroRange(rng, meta.Retention.Range)
	}
	if meta.Access.Value != "" {
		parts = append(parts, "access="+meta.Access.Value)
		rng = firstNonZeroRange(rng, meta.Access.Range)
	}
	if len(parts) == 0 {
		return "", sourcepos.Range{}, false
	}

	return strings.Join(parts, ", "), rng, true
}

func requirementTraceabilityDetail(requirement model.Requirement) (string, sourcepos.Range, bool) {
	if len(requirement.Traceability) == 0 {
		return "", sourcepos.Range{}, false
	}

	var values []string
	var rng sourcepos.Range
	for _, link := range requirement.Traceability {
		values = append(values, link.Value)
		rng = firstNonZeroRange(rng, link.Range)
	}

	return strings.Join(values, ", "), rng, true
}

func sortSymbols(symbols []protocol.DocumentSymbol) {
	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].Range.Start.Line == symbols[j].Range.Start.Line {
			return symbols[i].Name < symbols[j].Name
		}
		return symbols[i].Range.Start.Line < symbols[j].Range.Start.Line
	})
}

func firstNonZeroRange(vals ...sourcepos.Range) sourcepos.Range {
	for _, v := range vals {
		if v.Start.Line != 0 || v.Start.Column != 0 || v.End.Line != 0 || v.End.Column != 0 {
			return v
		}
	}

	return sourcepos.Range{}
}

// Package completion provides context-aware completion items for the DSL LSP.
package completion

import (
	"slices"
	"strings"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/values"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func Items(text string, position protocol.Position, result coreanalysis.Result) []protocol.CompletionItem {
	lines := bestLines(text, result)
	lineIndex, prefix, trimmedPrefix, inComment := cursorContext(lines, position)
	if inComment {
		return nil
	}

	valueDefinitions := bestValueDefinitions(result)
	var items []protocol.CompletionItem
	if kind, start := surroundingBlock(lines, lineIndex); kind == grammar.DeclarationKindAssignment && start >= 0 {
		items = assignmentItems(lines, lineIndex, prefix, trimmedPrefix, result.Index, valueDefinitions)
	} else {
		items = contextItems(lines, lineIndex, prefix, trimmedPrefix, result.Index)
	}

	return applyTextEdits(lines, position, items)
}

func bestLines(text string, result coreanalysis.Result) []grammar.TokenLine {
	if result.Index != nil && len(result.Index.Lines) > 0 {
		return result.Index.Lines
	}
	return grammar.TokenizeSource(text)
}

func cursorContext(lines []grammar.TokenLine, position protocol.Position) (int, string, string, bool) {
	if len(lines) == 0 {
		return 0, "", "", true
	}
	lineIndex := min(max(int(position.Line), 0), len(lines)-1)
	byteOffset, ok := sourcepos.ByteOffsetFromUTF16OK(lines[lineIndex].Raw, int(position.Character))
	if !ok {
		return lineIndex, "", "", true
	}
	visibleOffset, inComment := lines[lineIndex].VisibleOffset(byteOffset)
	prefix := lines[lineIndex].Code[:visibleOffset]
	return lineIndex, prefix, strings.TrimSpace(prefix), inComment
}

func contextItems(lines []grammar.TokenLine, lineIndex int, prefix string, trimmedPrefix string, index *docindex.Document) []protocol.CompletionItem {
	blockKind, blockStart := surroundingBlock(lines, lineIndex)
	if blockKind == grammar.DeclarationKindRequirement {
		for _, items := range [][]protocol.CompletionItem{
			planItems(completeActionPrefix(prefix), index),
			planItems(completeWhenPrefix(prefix), index),
			planItems(completeActorClausePrefix(prefix, "while"), index),
			planItems(completeActorClausePrefix(prefix, "if"), index),
			planItems(completeStakeholdersPrefix(prefix), index),
			planItems(completeMetadataPrefix(prefix), index),
		} {
			if len(items) > 0 {
				return items
			}
		}
		return requirementItems(lines, blockStart, lineIndex, trimmedPrefix, index)
	}
	if isTopLevelBoundary(lines, lineIndex) {
		return keywordItems(grammar.Compiled.TopLevelKeywords)
	}
	return nil
}

func planItems(plan completionPlan, index *docindex.Document) []protocol.CompletionItem {
	if !plan.inside || !plan.valid() {
		return nil
	}
	items := keywordItems(plan.keywords)
	if plan.expectStakeholder {
		items = appendUniqueItems(items, stakeholderItems(index)...)
	}
	if plan.expectVerb {
		items = appendUniqueItems(items, simpleItems(grammar.Compiled.WhenVerbs, protocol.CompletionItemKindMethod, "verb")...)
	}
	if plan.expectIdentifier {
		items = appendUniqueItems(items, identifierItems(index)...)
	}
	return items
}

func assignmentItems(lines []grammar.TokenLine, lineIndex int, prefix string, trimmedPrefix string, index *docindex.Document, valueDefinitions map[string]values.Definition) []protocol.CompletionItem {
	_, blockStart := surroundingBlock(lines, lineIndex)
	if lineIndex == blockStart {
		return assignmentPlanItems(completeAssignmentPrefix(prefix, true), index, valueDefinitions)
	}
	if trimmedPrefix == "" || isTopLevelKeywordPrefix(trimmedPrefix) {
		return appendUniqueItems(keywordItems(grammar.Compiled.TopLevelKeywords), assignmentPlanItems(completeAssignmentPrefix(prefix, false), index, valueDefinitions)...)
	}
	return assignmentPlanItems(completeAssignmentPrefix(prefix, false), index, valueDefinitions)
}

func requirementItems(lines []grammar.TokenLine, blockStart, lineIndex int, trimmedPrefix string, index *docindex.Document) []protocol.CompletionItem {
	seen := seenRequirementClauses(lines, blockStart, lineIndex)
	nextKeywords := grammar.Compiled.NextRequirementClauseKeywords(seen)
	items := keywordItems(nextKeywords)

	if slices.Contains(nextKeywords, "stakeholders") {
		items = appendUniqueItems(items, stakeholderItems(index)...)
	}
	if (trimmedPrefix == "" || isTopLevelKeywordPrefix(trimmedPrefix)) && slices.Contains(seen, grammar.RequirementClauseStakeholders) {
		items = appendUniqueItems(items, keywordItems(grammar.Compiled.TopLevelKeywords)...)
	}
	return items
}

func assignmentPlanItems(plan completionPlan, index *docindex.Document, valueDefinitions map[string]values.Definition) []protocol.CompletionItem {
	if !plan.inside || !plan.valid() {
		return nil
	}
	var items []protocol.CompletionItem
	if plan.expectRequirement {
		items = append(items, simpleItems(index.Names(grammar.DeclarationKindRequirement), protocol.CompletionItemKindReference, "requirement")...)
	}
	if plan.expectStakeholder {
		items = append(items, stakeholderItems(index)...)
	}
	if plan.expectValue {
		items = append(items, valueItems(valueDefinitions)...)
	}
	return items
}

func stakeholderItems(index *docindex.Document) []protocol.CompletionItem {
	if index == nil {
		return nil
	}
	return simpleItems(index.Names(grammar.DeclarationKindStakeholder), protocol.CompletionItemKindVariable, "stakeholder")
}

func surroundingBlock(lines []grammar.TokenLine, lineIndex int) (grammar.DeclarationKind, int) {
	for idx := lineIndex; idx >= 0; idx-- {
		if kind, ok := grammar.Compiled.DeclarationKindForLine(lines[idx].Code); ok {
			switch kind {
			case grammar.DeclarationKindRequirement, grammar.DeclarationKindAssignment:
				return kind, idx
			default:
				return "", idx
			}
		}
	}
	return "", -1
}

func seenRequirementClauses(lines []grammar.TokenLine, blockStart, lineIndex int) []grammar.RequirementClauseKind {
	seen := make([]grammar.RequirementClauseKind, 0, lineIndex-blockStart)
	for idx := blockStart + 1; idx < lineIndex; idx++ {
		if kind, ok := grammar.Compiled.RequirementClauseKindForLine(lines[idx].Code); ok {
			seen = append(seen, kind)
		}
	}

	return seen
}

func isTopLevelBoundary(lines []grammar.TokenLine, lineIndex int) bool {
	for idx := lineIndex - 1; idx >= 0; idx-- {
		if !lines[idx].IsBlank() {
			return isTopLevelBoundaryLine(lines[idx].Code)
		}
	}
	return true
}

func bestValueDefinitions(result coreanalysis.Result) map[string]values.Definition {
	defs := make(map[string]values.Definition)

	for _, b := range values.All() {
		defs[b.Name] = b
	}
	if result.Index == nil || result.Index.Model == nil {
		return defs
	}
	for _, v := range result.Index.Model.Values {
		defs[v.Name] = values.Definition{Name: v.Name, Category: v.Category, Angle: v.Angle, Radius: v.Radius}
	}

	return defs
}

package completion

import (
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func applyTextEdits(lines []grammar.TokenLine, position protocol.Position, items []protocol.CompletionItem) []protocol.CompletionItem {
	if len(items) == 0 {
		return nil
	}

	editRange := replacementRange(lines, position)
	for idx := range items {
		item := &items[idx]
		item.TextEdit = protocol.TextEdit{
			Range:   editRange,
			NewText: item.Label,
		}
		filterText := item.Label
		item.FilterText = &filterText
	}

	return items
}

func replacementRange(lines []grammar.TokenLine, position protocol.Position) protocol.Range {
	lineIndex := int(position.Line)
	if lineIndex < 0 || lineIndex >= len(lines) {
		return protocol.Range{
			Start: position,
			End:   position,
		}
	}

	line := lines[lineIndex].Code
	charIndex, ok := sourcepos.ByteOffsetFromUTF16OK(line, int(position.Character))
	if !ok {
		return protocol.Range{Start: position, End: position}
	}

	start := charIndex
	for start > 0 && grammar.IsIdentifierChar(line[start-1]) {
		start--
	}

	end := charIndex
	for end < len(line) && grammar.IsIdentifierChar(line[end]) {
		end++
	}

	startUTF16, ok := sourcepos.UTF16OffsetFromByteOK(line, start)
	if !ok {
		return protocol.Range{Start: position, End: position}
	}
	endUTF16, ok := sourcepos.UTF16OffsetFromByteOK(line, end)
	if !ok {
		return protocol.Range{Start: position, End: position}
	}

	return protocol.Range{
		Start: protocol.Position{
			Line:      position.Line,
			Character: protocol.UInteger(startUTF16),
		},
		End: protocol.Position{
			Line:      position.Line,
			Character: protocol.UInteger(endUTF16),
		},
	}
}

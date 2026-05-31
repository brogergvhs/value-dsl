package lsp

import (
	"net/url"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func toProtocolUint(value int) protocol.UInteger {
	if value <= 0 {
		return 0
	}
	return protocol.UInteger(value - 1)
}

func tokenLines(text string, result coreanalysis.Result) []grammar.TokenLine {
	if result.Index != nil && len(result.Index.Lines) > 0 {
		return result.Index.Lines
	}
	return grammar.TokenizeSource(text)
}

// lineAndByteOffset resolves the TokenLine and UTF-8 byte offset for a protocol
// position. Returns false if the line is out of range or the character cannot be
// converted from UTF-16.
func lineAndByteOffset(lines []grammar.TokenLine, position protocol.Position) (grammar.TokenLine, int, bool) {
	i := int(position.Line)
	if i < 0 || i >= len(lines) {
		return grammar.TokenLine{}, 0, false
	}
	offset, ok := sourcepos.ByteOffsetFromUTF16OK(lines[i].Raw, int(position.Character))
	return lines[i], offset, ok
}

func protocolPositionInLines(lines []grammar.TokenLine, position protocol.Position) sourcepos.Position {
	_, offset, ok := lineAndByteOffset(lines, position)
	if !ok {
		return sourcepos.NewPosition(int(position.Line)+1, 1)
	}
	return sourcepos.NewPosition(int(position.Line)+1, offset+1)
}

func toProtocolRangeInLines(lines []grammar.TokenLine, sourceRange sourcepos.Range) protocol.Range {
	start := toProtocolUint(sourceRange.Start.Column)
	end := toProtocolUint(sourceRange.End.Column)
	if line, ok := lineForNumber(lines, sourceRange.Start.Line); ok {
		if offset, valid := sourcepos.UTF16OffsetFromByteOK(line.Raw, max(sourceRange.Start.Column-1, 0)); valid {
			start = protocol.UInteger(offset)
		}
	}
	if line, ok := lineForNumber(lines, sourceRange.End.Line); ok {
		if offset, valid := sourcepos.UTF16OffsetFromByteOK(line.Raw, max(sourceRange.End.Column-1, 0)); valid {
			end = protocol.UInteger(offset)
		}
	}

	return protocol.Range{
		Start: protocol.Position{
			Line:      toProtocolUint(sourceRange.Start.Line),
			Character: start,
		},
		End: protocol.Position{
			Line:      toProtocolUint(sourceRange.End.Line),
			Character: end,
		},
	}
}

func toProtocolRangeInSourceFile(lines []grammar.TokenLine, file docindex.SourceFile, sourceRange sourcepos.Range) protocol.Range {
	r := toProtocolRangeInLines(lines, sourceRange)
	if file.StartLine > 0 {
		offset := protocol.UInteger(file.StartLine - 1)
		if r.Start.Line >= offset {
			r.Start.Line -= offset
		}
		if r.End.Line >= offset {
			r.End.Line -= offset
		}
	}
	return r
}

func locationForRange(defaultURI protocol.DocumentUri, result coreanalysis.Result, sourceRange sourcepos.Range) protocol.Location {
	location := protocol.Location{
		URI:   defaultURI,
		Range: toProtocolRangeInLines(result.Index.Lines, sourceRange),
	}
	if file, ok := result.Index.SourceFileForLine(sourceRange.Start.Line); ok {
		location.URI = fileURI(file.Path)
		location.Range = toProtocolRangeInSourceFile(result.Index.Lines, file, sourceRange)
	}
	return location
}

func fileURI(path string) protocol.DocumentUri {
	return protocol.DocumentUri((&url.URL{Scheme: "file", Path: path}).String())
}

func lineForNumber(lines []grammar.TokenLine, number int) (grammar.TokenLine, bool) {
	if i := number - 1; i >= 0 && i < len(lines) {
		return lines[i], true
	}
	return grammar.TokenLine{}, false
}

func occurrenceAtPosition(result coreanalysis.Result, position protocol.Position) (docindex.Occurrence, bool) {
	if result.Index == nil {
		return docindex.Occurrence{}, false
	}
	return result.Index.OccurrenceAt(protocolPositionInLines(result.Index.Lines, position))
}

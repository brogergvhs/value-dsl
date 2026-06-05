package lsp

import (
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// selectionRange handles the textDocument/selectionRange request.
func (s *Server) selectionRange(_ *glsp.Context, params *protocol.SelectionRangeParams) ([]protocol.SelectionRange, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.SelectionRange{}, nil
	}
	if result.Index == nil {
		return []protocol.SelectionRange{}, nil
	}

	file, hasFile := sourceFileForURI(protocol.DocumentUri(document.URI), result.Index.SourceFiles)
	folds := []sourceFold(nil)
	if result.Index.AST != nil {
		folds = sourceFolds(result.Index.AST.Declarations)
	}

	toRange := func(rng sourcepos.Range) protocol.Range {
		if hasFile {
			return toProtocolRangeInSourceFile(result.Index.Lines, file, rng)
		}
		return toProtocolRangeInLines(result.Index.Lines, rng)
	}

	ranges := make([]protocol.SelectionRange, 0, len(params.Positions))
	for _, position := range params.Positions {
		sourceLine := int(position.Line) + 1
		if hasFile {
			sourceLine += file.StartLine - 1
		}

		leaf := protocol.Range{Start: position, End: position}
		if line, ok := lineForNumber(result.Index.Lines, sourceLine); ok {
			leaf = toRange(sourcepos.NewRange(
				sourcepos.NewPosition(sourceLine, 1),
				sourcepos.NewPosition(sourceLine, len(line.Raw)+1),
			))
		}
		sourcePosition := protocol.Position{Line: toProtocolUint(sourceLine), Character: position.Character}
		if line, byteOffset, ok := lineAndByteOffset(result.Index.Lines, sourcePosition); ok {
			if visibleOffset, inComment := line.VisibleOffset(byteOffset); !inComment {
				if token, ok := line.TokenAtOffset(visibleOffset); ok {
					leaf = toRange(token.Range(line.Line))
				}
			}
		}

		var parent *protocol.SelectionRange
		var parentSize int
		for _, fold := range folds {
			if sourceLine < fold.startLine || sourceLine > fold.endLine {
				continue
			}
			if hasFile && (fold.startLine < file.StartLine || fold.endLine > file.EndLine) {
				continue
			}
			start, ok := lineForNumber(result.Index.Lines, fold.startLine)
			if !ok {
				continue
			}
			end, ok := lineForNumber(result.Index.Lines, fold.endLine)
			if !ok || end.Line < start.Line {
				continue
			}
			rng := toRange(sourcepos.NewRange(
				sourcepos.NewPosition(fold.startLine, 1),
				sourcepos.NewPosition(fold.endLine, len(end.Raw)+1),
			))
			size := fold.endLine - fold.startLine
			if rng != leaf && (parent == nil || size < parentSize) {
				parentSize = size
				parent = &protocol.SelectionRange{Range: rng}
			}
		}

		ranges = append(ranges, protocol.SelectionRange{Range: leaf, Parent: parent})
	}
	return ranges, nil
}

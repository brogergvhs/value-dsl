package lsp

import (
	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// documentHighlight handles the textDocument/documentHighlight request.
func (s *Server) documentHighlight(_ *glsp.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.DocumentHighlight{}, nil
	}
	return resolveDocumentHighlights(params.Position, s.navigationAnalysis(document.URI, result)), nil
}

func resolveDocumentHighlights(position protocol.Position, result coreanalysis.Result) []protocol.DocumentHighlight {
	occurrence, ok := occurrenceAtPosition(result, position)
	if !ok {
		return nil
	}

	highlights := make([]protocol.DocumentHighlight, 0, 1+len(result.Index.ReferencesFor(occurrence.Kind, occurrence.Name)))
	writeKind := protocol.DocumentHighlightKindWrite
	if decl, ok := result.Index.LookupDeclaration(occurrence.Kind, occurrence.Name); ok {
		highlights = append(highlights, protocol.DocumentHighlight{
			Range: toProtocolRangeInLines(result.Index.Lines, decl.Range),
			Kind:  &writeKind,
		})
	}

	readKind := protocol.DocumentHighlightKindRead
	for _, ref := range result.Index.ReferencesFor(occurrence.Kind, occurrence.Name) {
		highlights = append(highlights, protocol.DocumentHighlight{
			Range: toProtocolRangeInLines(result.Index.Lines, ref.Range),
			Kind:  &readKind,
		})
	}

	return highlights
}

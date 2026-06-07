package lsp

import (
	"github.com/brogergvhs/value-dsl/internal/lsp/completion"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// complete handles the textDocument/completion request.
func (s *Server) complete(_ *glsp.Context, params *protocol.CompletionParams) (any, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return protocol.CompletionList{IsIncomplete: false}, nil
	}
	return protocol.CompletionList{IsIncomplete: false, Items: completion.Items(document.Text, params.Position, s.navigationAnalysis(document.URI, result))}, nil
}

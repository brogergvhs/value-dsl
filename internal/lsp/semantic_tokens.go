package lsp

import (
	"github.com/brogergvhs/value-dsl/internal/lsp/semantictokens"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// semanticTokens handles the textDocument/semanticTokens/full request.
func (s *Server) semanticTokens(_ *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return &protocol.SemanticTokens{}, nil
	}
	return semantictokens.Resolve(document.Text, result), nil
}

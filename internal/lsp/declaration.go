package lsp

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// declaration handles the textDocument/declaration request.
// for current state, semantically equivalent to definition.
func (s *Server) declaration(context *glsp.Context, params *protocol.DeclarationParams) (any, error) {
	return s.definition(context, &protocol.DefinitionParams{
		TextDocumentPositionParams: params.TextDocumentPositionParams,
		WorkDoneProgressParams:     params.WorkDoneProgressParams,
		PartialResultParams:        params.PartialResultParams,
	})
}

package lsp

import (
	"fmt"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/formatting"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// format handles the textDocument/formatting request.
func (s *Server) format(_ *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	document, ok := s.documents.Get(params.TextDocument.URI)
	if !ok {
		return []protocol.TextEdit{}, nil
	}
	formatted, err := formatting.Format(document.Text)
	if err != nil {
		return nil, fmt.Errorf("format document: %w", err)
	}
	if formatted == document.Text {
		return []protocol.TextEdit{}, nil
	}
	return []protocol.TextEdit{{Range: fullDocumentRange(document.Text), NewText: formatted}}, nil
}

func fullDocumentRange(text string) protocol.Range {
	lastLine := strings.Count(text, "\n")
	lastLineText := text
	if i := strings.LastIndex(text, "\n"); i >= 0 {
		lastLineText = text[i+1:]
	}
	lastCharacter, _ := sourcepos.UTF16OffsetFromByteOK(lastLineText, len(lastLineText))
	return protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: uint32(lastLine), Character: uint32(lastCharacter)},
	}
}

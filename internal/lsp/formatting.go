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

// rangeFormat handles the textDocument/rangeFormatting request.
func (s *Server) rangeFormat(_ *glsp.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	document, ok := s.documents.Get(params.TextDocument.URI)
	if !ok {
		return []protocol.TextEdit{}, nil
	}

	lines := strings.SplitAfter(document.Text, "\n")
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return []protocol.TextEdit{}, nil
	}

	start := min(max(int(params.Range.Start.Line), 0), len(lines)-1)
	end := int(params.Range.End.Line)
	if params.Range.End.Character > 0 {
		end++
	}
	end = min(max(end, start+1), len(lines))

	fragment := strings.Join(lines[start:end], "")
	formatted, err := formatting.Format(fragment)
	if err != nil {
		return nil, fmt.Errorf("format range: %w", err)
	}
	if formatted == fragment {
		return []protocol.TextEdit{}, nil
	}

	rng := protocol.Range{
		Start: protocol.Position{Line: uint32(start), Character: 0},
		End:   protocol.Position{Line: uint32(end), Character: 0},
	}
	if last := lines[end-1]; !strings.HasSuffix(last, "\n") {
		character, _ := sourcepos.UTF16OffsetFromByteOK(last, len(last))
		rng.End = protocol.Position{Line: uint32(end - 1), Character: uint32(character)}
	}
	return []protocol.TextEdit{{Range: rng, NewText: formatted}}, nil
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

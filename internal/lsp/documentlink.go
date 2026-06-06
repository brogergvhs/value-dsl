package lsp

import (
	"net/url"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// documentLink handles the textDocument/documentLink request.
func (s *Server) documentLink(_ *glsp.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.DocumentLink{}, nil
	}
	result = s.navigationAnalysis(document.URI, result)
	if result.Index == nil || result.Index.Model == nil {
		return []protocol.DocumentLink{}, nil
	}

	file, hasFile := sourceFileForURI(protocol.DocumentUri(document.URI), result.Index.SourceFiles)
	links := []protocol.DocumentLink{}
	for _, req := range result.Index.Model.Requirements {
		for _, link := range req.Traceability {
			target, ok := documentLinkTarget(link.Value)
			if !ok || hasFile && (link.Line < file.StartLine || link.Line > file.EndLine) {
				continue
			}
			sourceRange, ok := traceabilityTargetRange(result.Index.Lines, link)
			if !ok {
				continue
			}
			rng := toProtocolRangeInLines(result.Index.Lines, sourceRange)
			if hasFile {
				rng = toProtocolRangeInSourceFile(result.Index.Lines, file, sourceRange)
			}
			links = append(links, protocol.DocumentLink{Range: rng, Target: &target})
		}
	}
	return links, nil
}

func traceabilityTargetRange(lines []grammar.TokenLine, link model.TraceabilityLink) (sourcepos.Range, bool) {
	line, ok := lineForNumber(lines, link.Line)
	if !ok || len(line.Tokens) < 2 {
		return sourcepos.Range{}, false
	}
	return line.Tokens[1].Range(line.Line), true
}

func documentLinkTarget(value string) (protocol.DocumentUri, bool) {
	value = strings.TrimSpace(value)
	u, err := url.Parse(value)
	if err == nil {
		switch u.Scheme {
		case "http", "https", "file":
			return protocol.DocumentUri(value), true
		}
	}
	if strings.Contains(value, ".") && !strings.ContainsAny(value, " \t") {
		return protocol.DocumentUri("https://" + value), true
	}
	return "", false
}

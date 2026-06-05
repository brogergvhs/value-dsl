package lsp

import (
	"strings"

	"github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type workspaceSymbolKey struct {
	name      string
	kind      protocol.SymbolKind
	uri       protocol.DocumentUri
	startLine protocol.UInteger
	startChar protocol.UInteger
}

// workspaceSymbol handles the workspace/symbol request.
func (s *Server) workspaceSymbol(_ *glsp.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	seen := make(map[workspaceSymbolKey]struct{})
	symbols := []protocol.SymbolInformation{}

	for _, document := range s.documents.All() {
		result := s.navigationAnalysis(document.URI, s.analyzeDocument(document))
		for _, symbol := range buildWorkspaceSymbols(protocol.DocumentUri(document.URI), result, params.Query) {
			key := workspaceSymbolDedupKey(symbol)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

func buildWorkspaceSymbols(uri protocol.DocumentUri, result analysis.Result, query string) []protocol.SymbolInformation {
	if result.Index == nil {
		return nil
	}

	symbols := make([]protocol.SymbolInformation, 0, len(result.Index.Declarations))
	for _, decl := range result.Index.Declarations {
		if !matchesWorkspaceSymbolQuery(decl, query) {
			continue
		}
		symbols = append(symbols, protocol.SymbolInformation{
			Name:     decl.Name,
			Kind:     symbolKindForDeclarationKind(decl.Kind),
			Location: locationForRange(uri, result, decl.Range),
		})
	}

	return symbols
}

func matchesWorkspaceSymbolQuery(decl docindex.Declaration, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(decl.Name), query) ||
		strings.Contains(strings.ToLower(string(decl.Kind)), query)
}

func workspaceSymbolDedupKey(symbol protocol.SymbolInformation) workspaceSymbolKey {
	return workspaceSymbolKey{
		name:      symbol.Name,
		kind:      symbol.Kind,
		uri:       symbol.Location.URI,
		startLine: symbol.Location.Range.Start.Line,
		startChar: symbol.Location.Range.Start.Character,
	}
}

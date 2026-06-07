package lsp

import (
	"fmt"
	"slices"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/grammar"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// rename handles the textDocument/rename request.
func (s *Server) rename(_ *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return &protocol.WorkspaceEdit{}, nil
	}
	return resolveRename(params.TextDocument.URI, params.Position, params.NewName, s.navigationAnalysis(document.URI, result))
}

func resolveRename(uri protocol.DocumentUri, position protocol.Position, newName string, result coreanalysis.Result) (*protocol.WorkspaceEdit, error) {
	if !grammar.Compiled.IsValidIdentifier(newName) {
		return nil, fmt.Errorf("rename target %q is not a valid DSL identifier", newName)
	}
	if slices.Contains(grammar.Compiled.ReservedKeywords, newName) {
		return nil, fmt.Errorf("rename target %q must not be a reserved keyword", newName)
	}

	occurrence, ok := occurrenceAtPosition(result, position)
	if !ok {
		return nil, fmt.Errorf("no renamable symbol at cursor")
	}
	if existing, ok := result.Index.LookupDeclaration(occurrence.Kind, newName); ok && existing.Name != occurrence.Name {
		return nil, fmt.Errorf("rename target %q would conflict with an existing %s", newName, occurrence.Kind)
	}

	locations := resolveReferences(uri, position, true, result)
	if len(locations) == 0 {
		return nil, fmt.Errorf("no references found for %q", occurrence.Name)
	}

	changes := make(map[protocol.DocumentUri][]protocol.TextEdit)
	for _, location := range locations {
		changes[location.URI] = append(changes[location.URI], protocol.TextEdit{
			Range:   location.Range,
			NewText: newName,
		})
	}

	return &protocol.WorkspaceEdit{
		Changes: changes,
	}, nil
}

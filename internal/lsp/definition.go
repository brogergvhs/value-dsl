package lsp

import (
	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func resolveDefinition(uri protocol.DocumentUri, position protocol.Position, result coreanalysis.Result) []protocol.Location {
	occurrence, ok := occurrenceAtPosition(result, position)
	if !ok {
		return nil
	}
	declaration, ok := result.Index.LookupDeclaration(occurrence.Kind, occurrence.Name)
	if !ok {
		return nil
	}
	return []protocol.Location{locationForRange(uri, result, declaration.Range)}
}

package lsp

import (
	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

func resolveReferences(uri protocol.DocumentUri, position protocol.Position, includeDeclaration bool, result coreanalysis.Result) []protocol.Location {
	occurrence, ok := occurrenceAtPosition(result, position)
	if !ok {
		return nil
	}

	var locations []protocol.Location
	if includeDeclaration {
		if decl, ok := result.Index.LookupDeclaration(occurrence.Kind, occurrence.Name); ok {
			locations = append(locations, locationForRange(uri, result, decl.Range))
		}
	}
	for _, ref := range result.Index.ReferencesFor(occurrence.Kind, occurrence.Name) {
		locations = append(locations, locationForRange(uri, result, ref.Range))
	}

	return locations
}

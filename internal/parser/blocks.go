package parser

import "github.com/brogergvhs/value-dsl/internal/grammar"

type structuralBlock struct {
	Kind   grammar.DeclarationKind
	Header grammar.TokenLine
	Body   []grammar.TokenLine
}

func parseBlocksFromLines(lines []grammar.TokenLine) []structuralBlock {
	blocks := make([]structuralBlock, 0)
	current := -1
	for _, source := range lines {
		if source.IsBlank() {
			continue
		}

		if kind, ok := grammar.Compiled.DeclarationKindForTokens(source.Tokens); ok {
			blocks = append(blocks, structuralBlock{Kind: kind, Header: source})

			if decl, found := grammar.Compiled.DeclarationByKind[kind]; found && decl.Body != nil {
				current = len(blocks) - 1
			} else {
				current = -1
			}
			continue
		}

		if current >= 0 {
			blocks[current].Body = append(blocks[current].Body, source)
			continue
		}

		blocks = append(blocks, structuralBlock{Header: source})
	}

	return blocks
}

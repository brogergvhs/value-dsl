package parser

import (
	"testing"

	"github.com/brogergvhs/value-dsl/internal/grammar"
)

func TestParseBlocksGroupsRequirementAndAssignmentBodies(t *testing.T) {
	blocks := parseBlocksFromLines(grammar.TokenizeSource(`stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`))

	if len(blocks) != 4 {
		t.Fatalf("expected 4 structural blocks, got %+v", blocks)
	}
	if blocks[0].Kind != grammar.DeclarationKindStakeholder {
		t.Fatalf("expected stakeholder block, got %+v", blocks[0])
	}
	if blocks[1].Kind != grammar.DeclarationKindValue {
		t.Fatalf("expected value block, got %+v", blocks[1])
	}
	if blocks[2].Kind != grammar.DeclarationKindRequirement || len(blocks[2].Body) != 3 {
		t.Fatalf("expected requirement block with 3 body lines, got %+v", blocks[2])
	}
	if blocks[3].Kind != grammar.DeclarationKindAssignment || len(blocks[3].Body) != 1 {
		t.Fatalf("expected assignment block with 1 body line, got %+v", blocks[3])
	}
}

func TestParseBlocksPreservesUnknownTopLevelLines(t *testing.T) {
	blocks := parseBlocksFromLines(grammar.TokenizeSource(`stakeholder Worker
system shall notify Worker

requirement R1
when Worker enters DangerousArea
`))

	if len(blocks) != 3 {
		t.Fatalf("expected 3 structural blocks, got %+v", blocks)
	}
	if blocks[1].Kind != "" {
		t.Fatalf("expected unknown top-level block, got %+v", blocks[1])
	}
	if blocks[1].Header.Line != 2 {
		t.Fatalf("expected unknown line 2, got %+v", blocks[1])
	}
}

func TestParseBlocksStripsCommentsForClassificationButPreservesRawLine(t *testing.T) {
	blocks := parseBlocksFromLines(grammar.TokenizeSource(`requirement R1 // header
when Worker enters DangerousArea // note
`))

	if len(blocks) != 1 {
		t.Fatalf("expected one block, got %+v", blocks)
	}
	if blocks[0].Header.Raw != `requirement R1 // header` {
		t.Fatalf("expected raw header to be preserved, got %+v", blocks[0].Header)
	}
	if blocks[0].Header.Code != `requirement R1 ` {
		t.Fatalf("expected stripped code header, got %+v", blocks[0].Header)
	}
	if len(blocks[0].Body) != 1 || blocks[0].Body[0].Code != `when Worker enters DangerousArea ` {
		t.Fatalf("expected stripped body line, got %+v", blocks[0].Body)
	}
}

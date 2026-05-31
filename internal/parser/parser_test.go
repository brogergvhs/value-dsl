package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
)

func requirements(doc *ast.DocumentNode) []ast.DeclarationNode {
	return doc.DeclarationsOfKind(grammar.DeclarationKindRequirement)
}

func assignments(doc *ast.DocumentNode) []ast.DeclarationNode {
	return doc.DeclarationsOfKind(grammar.DeclarationKindAssignment)
}

func stakeholders(doc *ast.DocumentNode) []ast.DeclarationNode {
	return doc.DeclarationsOfKind(grammar.DeclarationKindStakeholder)
}

func vals(doc *ast.DocumentNode) []ast.DeclarationNode {
	return doc.DeclarationsOfKind(grammar.DeclarationKindValue)
}

func requireSingleWhenClause(t *testing.T, body []ast.GenericNode) ast.GenericNode {
	t.Helper()

	clauses := requirementEntries(ast.DeclarationNode{Body: body}, grammar.RequirementClauseWhen)
	if len(clauses) != 1 {
		t.Fatalf("expected exactly one context clause, got %+v", clauses)
	}
	whenClause := clauses[0]
	if whenClause.Kind != string(grammar.RequirementClauseWhen) {
		t.Fatalf("expected when clause, got %+v", whenClause)
	}
	return whenClause
}

func firstRequirementEntry(t *testing.T, req ast.DeclarationNode, kind grammar.RequirementClauseKind) ast.GenericNode {
	t.Helper()
	for _, entry := range req.Body {
		if entry.Kind == string(kind) {
			return entry
		}
	}
	t.Fatalf("expected requirement entry %q, got %+v", kind, req.Body)
	return ast.GenericNode{}
}

func requirementEntries(req ast.DeclarationNode, kind grammar.RequirementClauseKind) []ast.GenericNode {
	out := make([]ast.GenericNode, 0)
	for _, entry := range req.Body {
		if entry.Kind == string(kind) {
			out = append(out, entry)
		}
	}
	return out
}

func TestParseStakeholderDeclarations(t *testing.T) {
	input := `stakeholder Worker
stakeholder Manager`

	doc := Parse(input).AST

	decls := stakeholders(doc)
	if len(decls) != 2 {
		t.Fatalf("expected 2 stakeholder declarations, got %d", len(decls))
	}
	if decls[0].Header.First("names") != "Worker" || decls[1].Header.First("names") != "Manager" {
		t.Fatalf("unexpected stakeholders: %+v, %+v", decls[0].Header, decls[1].Header)
	}
	if decls[0].Line != 1 {
		t.Fatalf("expected first stakeholder line 1, got %d", decls[0].Line)
	}
}

func TestParseValueDeclarations(t *testing.T) {
	input := `value privacy_pref = 1.58, 0.91
value authority_pref = 0.02, 0.88`

	doc := Parse(input).AST

	decls := vals(doc)
	if len(decls) != 2 {
		t.Fatalf("expected 2 values, got %d", len(decls))
	}
	if decls[0].Header.First("name") != "privacy_pref" || decls[0].Header.First("angle") != "1.58" || decls[0].Header.First("radius") != "0.91" {
		t.Fatalf("unexpected first value: %+v", decls[0].Header)
	}
	if !decls[0].Header.HasText(ast.FieldKeyword, ",") {
		t.Fatalf("expected value comma keyword, got %+v", decls[0].Header)
	}
	if decls[0].Line != 1 {
		t.Fatalf("expected first value line 1, got %d", decls[0].Line)
	}
}

func TestParseIncompleteValueDeclarationsWithoutAbortingDocument(t *testing.T) {
	input := `stakeholder Worker

value privacy_pref = 10.8,

requirement R1
system shall notify Worker
stakeholders Worker
`

	doc := Parse(input).AST
	decls := vals(doc)
	if len(decls) != 1 {
		t.Fatalf("expected one value, got %+v", decls)
	}
	if decls[0].Header.First("angle") == "" || decls[0].Header.First("radius") != "" || !decls[0].Header.HasText(ast.FieldKeyword, ",") {
		t.Fatalf("expected partial value geometry, got %+v", decls[0].Header)
	}
	reqs := requirements(doc)
	if len(reqs) != 1 || reqs[0].Header.First("id") != "R1" {
		t.Fatalf("expected requirement to still parse, got %+v", reqs)
	}
}

func TestParseRequirement(t *testing.T) {
	input := `stakeholder Worker
stakeholder Manager

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager
`

	doc := Parse(input).AST

	reqs := requirements(doc)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(reqs))
	}
	req := reqs[0]
	if req.Header.First("id") != "R1" {
		t.Fatalf("expected requirement id R1, got %q", req.Header.First("id"))
	}
	if req.Line != 4 {
		t.Fatalf("expected requirement line 4, got %d", req.Line)
	}
	whenClause := requireSingleWhenClause(t, req.Body)
	if whenClause.First(ast.FieldRawText) != "Worker enters DangerousArea" {
		t.Fatalf("unexpected condition: %q", whenClause.First(ast.FieldRawText))
	}
	if whenClause.Line != 5 {
		t.Fatalf("expected condition line 5, got %d", whenClause.Line)
	}
	if whenClause.First("actor") != "Worker" || whenClause.First("verb") != "enters" || whenClause.First("target") != "DangerousArea" {
		t.Fatalf("unexpected condition fields: %+v", whenClause)
	}
	action := firstRequirementEntry(t, req, grammar.RequirementClauseSystemShall)
	if action.First("verb") != "track" {
		t.Fatalf("expected verb track, got %q", action.First("verb"))
	}
	if action.First("object") != "location" {
		t.Fatalf("expected object location, got %q", action.First("object"))
	}
	if action.First("target") != "Worker" {
		t.Fatalf("expected target Worker, got %q", action.First("target"))
	}
	if action.First("mechanism") != "Camera" {
		t.Fatalf("expected mechanism Camera, got %q", action.First("mechanism"))
	}
	if action.First(ast.FieldRawText) != "track location of Worker using Camera" {
		t.Fatalf("unexpected action raw text: %q", action.First(ast.FieldRawText))
	}
	if action.Line != 6 {
		t.Fatalf("expected action line 6, got %d", action.Line)
	}
	sh := firstRequirementEntry(t, req, grammar.RequirementClauseStakeholders)
	if sh.Line != 7 {
		t.Fatalf("expected stakeholders line 7, got %d", sh.Line)
	}
	if len(sh.Texts("actors")) != 2 || sh.Texts("actors")[0] != "Worker" || sh.Texts("actors")[1] != "Manager" {
		t.Fatalf("unexpected requirement stakeholders: %+v", sh)
	}
}

func TestParseCanonicalActionKinds(t *testing.T) {
	tests := []struct {
		name          string
		actionLine    string
		wantVerb      string
		wantObject    string
		wantTarget    string
		wantMechanism string
	}{
		{
			name:          "monitoring",
			actionLine:    "system shall track location of Worker using Camera",
			wantVerb:      "track",
			wantObject:    "location",
			wantTarget:    "Worker",
			wantMechanism: "Camera",
		},
		{
			name:          "notification",
			actionLine:    "system shall notify Worker using AudioAlarm",
			wantVerb:      "notify",
			wantTarget:    "Worker",
			wantMechanism: "AudioAlarm",
		},
		{
			name:          "persistence",
			actionLine:    "system shall log location of Worker using Database",
			wantVerb:      "log",
			wantObject:    "location",
			wantTarget:    "Worker",
			wantMechanism: "Database",
		},
		{
			name:       "restriction",
			actionLine: "system shall anonymize location of Worker",
			wantVerb:   "anonymize",
			wantObject: "location",
			wantTarget: "Worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "stakeholder Worker\n\nrequirement R1\nwhen Worker enters DangerousArea\n" + tt.actionLine + "\nstakeholders Worker\n"

			doc := Parse(input).AST

			req := requirements(doc)[0]
			action := firstRequirementEntry(t, req, grammar.RequirementClauseSystemShall)
			if action.First("verb") != tt.wantVerb {
				t.Fatalf("expected verb %q, got %q", tt.wantVerb, action.First("verb"))
			}
			if action.First("object") != tt.wantObject {
				t.Fatalf("expected object %q, got %q", tt.wantObject, action.First("object"))
			}
			if action.First("target") != tt.wantTarget {
				t.Fatalf("expected target %q, got %q", tt.wantTarget, action.First("target"))
			}
			if action.First("mechanism") != tt.wantMechanism {
				t.Fatalf("expected mechanism %q, got %q", tt.wantMechanism, action.First("mechanism"))
			}
		})
	}
}

func TestMissingClausesArePreservedForValidation(t *testing.T) {
	input := `requirement R1
stakeholders Worker
`

	doc := Parse(input).AST
	reqs := requirements(doc)
	if len(reqs) != 1 {
		t.Fatalf("expected requirement to be preserved, got %+v", reqs)
	}
	if len(requirementEntries(reqs[0], grammar.RequirementClauseSystemShall)) != 0 {
		t.Fatalf("expected empty action to be preserved via missing action entry, got %+v", reqs[0].Body)
	}
}

func TestInvalidSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "invalid stakeholder identifier",
			input: `stakeholder 123Worker
`,
		},
		{
			name: "unknown keyword",
			input: `stakeholders Worker
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse(tt.input).AST
			if doc == nil {
				t.Fatal("expected document, got nil")
			}
		})
	}
}

func TestParseWithComments(t *testing.T) {
	input := `// top-level comment
stakeholder Worker

// value comment
value privacy_pref = 1.58, 0.91 // builtin category

// requirement comment
requirement R1 // header comment
when Worker enters DangerousArea
system shall track location of Worker using Camera // action comment
stakeholders Worker
`

	doc := Parse(input).AST

	sh := stakeholders(doc)
	if len(sh) != 1 || sh[0].Header.First("names") != "Worker" {
		t.Fatalf("unexpected stakeholders: %+v", sh)
	}
	v := vals(doc)
	if len(v) != 1 || v[0].Header.First("name") != "privacy_pref" {
		t.Fatalf("unexpected values: %+v", v)
	}
	reqs := requirements(doc)
	if len(reqs) != 1 || reqs[0].Header.First("id") != "R1" {
		t.Fatalf("unexpected requirements: %+v", reqs)
	}
}

func TestParseRequirementPreservesUnsupportedActionVerb(t *testing.T) {
	input := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall inspect location of Worker using Camera
stakeholders Worker
`

	doc := Parse(input).AST
	reqs := requirements(doc)
	if len(reqs) != 1 || firstRequirementEntry(t, reqs[0], grammar.RequirementClauseSystemShall).First("verb") != "inspect" {
		t.Fatalf("expected unsupported action verb to be preserved, got %+v", reqs)
	}
}

func TestParsePreservesSemanticallyInvalidActionShapes(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{
			name:   "notification using of",
			action: "system shall notify Worker of AudioAlarm",
		},
		{
			name:   "persistence using without mechanism",
			action: "system shall log location of Worker using",
		},
		{
			name:   "restriction with using",
			action: "system shall anonymize location of Worker using Camera",
		},
		{
			name:   "reserved keyword operand",
			action: "system shall track using of Worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "stakeholder Worker\n\nrequirement R1\nwhen Worker enters DangerousArea\n" + tt.action + "\nstakeholders Worker\n"

			doc := Parse(input).AST
			if len(requirements(doc)) != 1 {
				t.Fatalf("expected requirement to be preserved, got %+v", requirements(doc))
			}
		})
	}
}

func TestParseRejectsStructurallyInvalidActionShapes(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		wantErr string
	}{
		{
			name:    "monitoring missing of",
			action:  "system shall track location Worker using Camera",
			wantErr: "parse DSL:",
		},
		{
			name:    "persistence missing of",
			action:  "system shall log location Worker using Database",
			wantErr: "parse DSL:",
		},
		{
			name:    "multi token object",
			action:  "system shall track protective equipment of Worker using Camera",
			wantErr: "parse DSL:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "stakeholder Worker\n\nrequirement R1\nwhen Worker enters DangerousArea\n" + tt.action + "\nstakeholders Worker\n"

			doc := Parse(input).AST
			if len(requirements(doc)) != 1 {
				t.Fatalf("expected requirement to be preserved, got %+v", requirements(doc))
			}
		})
	}
}

func TestParseValidActionExamples(t *testing.T) {
	files := []string{
		"01_minimal_valid.dsl",
		"02_multi_requirement_valid.dsl",
		"03_with_comments.dsl",
		"07_valid_notification.dsl",
		"08_valid_persistence.dsl",
		"09_valid_restriction.dsl",
		"17_valid_inline_assignments.dsl",
		"18_invalid_assignment_duplicate.dsl",
		"19_valid_metadata.dsl",
		"20_valid_traceability.dsl",
		"21_valid_ubiquitous.dsl",
		"22_valid_while.dsl",
		"23_valid_if.dsl",
		"24_valid_where.dsl",
		"25_valid_while_when.dsl",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			input := readExampleDSL(t, file)
			if Parse(input).AST == nil {
				t.Fatalf("Parse(%s).AST returned nil AST", file)
			}
		})
	}
}

func TestParseInvalidActionExamples(t *testing.T) {
	tests := []struct {
		file    string
		wantErr string
	}{
		{"14_invalid_multi_token_object.dsl", "parse DSL:"},
		{"15_invalid_multi_token_mechanism.dsl", "parse DSL:"},
		{"30_invalid_malformed_when.dsl", "parse DSL:"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			input := readExampleDSL(t, tt.file)
			doc := Parse(input).AST
			if len(requirements(doc)) == 0 {
				t.Fatalf("expected malformed example %s to preserve requirements", tt.file)
			}
		})
	}
}

func TestParseSemanticallyInvalidActionExamples(t *testing.T) {
	files := []string{
		"10_invalid_verb_class_mismatch.dsl",
		"11_invalid_restriction_using.dsl",
		"12_invalid_notification_of.dsl",
		"13_invalid_unknown_action_verb.dsl",
		"16_invalid_reserved_keyword_operand.dsl",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			input := readExampleDSL(t, file)
			if Parse(input).AST == nil {
				t.Fatalf("Parse(%s).AST returned nil AST", file)
			}
		})
	}
}

func TestParseStructuredCondition(t *testing.T) {
	input := `stakeholder Worker

requirement R1
when Worker reports Hazard
system shall notify Worker using AudioAlarm
stakeholders Worker
`

	doc := Parse(input).AST

	req := requirements(doc)[0]
	whenClause := requireSingleWhenClause(t, req.Body)
	if whenClause.First("actor") != "Worker" || whenClause.First("verb") != "reports" || whenClause.First("target") != "Hazard" {
		t.Fatalf("unexpected condition fields: %+v", whenClause)
	}
	if whenClause.First(ast.FieldRawText) != "Worker reports Hazard" {
		t.Fatalf("unexpected condition raw text: %q", whenClause.First(ast.FieldRawText))
	}
}

func TestParseRejectsInvalidConditionShapes(t *testing.T) {
	tests := []struct {
		name    string
		when    string
		wantErr string
	}{
		{
			name:    "unknown condition verb",
			when:    "when Worker handles HazardousMaterial",
			wantErr: "parse DSL:",
		},
		{
			name:    "multi token target",
			when:    "when Worker enters Danger Zone",
			wantErr: "parse DSL:",
		},
		{
			name:    "extra tokens",
			when:    "when Worker enters DangerZone immediately",
			wantErr: "parse DSL:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := "stakeholder Worker\n\nrequirement R1\n" + tt.when + "\nsystem shall track location of Worker using Camera\nstakeholders Worker\n"

			doc := Parse(input).AST
			if len(requirements(doc)) != 1 {
				t.Fatalf("expected requirement to be preserved, got %+v", requirements(doc))
			}
		})
	}
}

func TestParseEARSPatterns(t *testing.T) {
	tests := []struct {
		name      string
		dsl       string
		wantTypes []grammar.RequirementClauseKind
	}{
		{
			name: "ubiquitous",
			dsl: `stakeholder Worker

requirement R1
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
		},
		{
			name: "while",
			dsl: `stakeholder Worker

requirement R1
while Worker is in DangerousArea
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
			wantTypes: []grammar.RequirementClauseKind{grammar.RequirementClauseWhile},
		},
		{
			name: "if",
			dsl: `stakeholder Worker

requirement R1
if Worker has_no_consent
system shall mask location of Worker
stakeholders Worker
`,
			wantTypes: []grammar.RequirementClauseKind{grammar.RequirementClauseIf},
		},
		{
			name: "where",
			dsl: `stakeholder Worker

requirement R1
where SafetyMode enabled
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
			wantTypes: []grammar.RequirementClauseKind{grammar.RequirementClauseWhere},
		},
		{
			name: "while plus when",
			dsl: `stakeholder Worker

requirement R1
while Worker is in DangerousArea
when Worker enters RestrictedZone
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
			wantTypes: []grammar.RequirementClauseKind{grammar.RequirementClauseWhile, grammar.RequirementClauseWhen},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse(tt.dsl).AST

			clauses := requirementContextNodes(requirements(doc)[0])
			if len(clauses) != len(tt.wantTypes) {
				t.Fatalf("expected %d clauses, got %+v", len(tt.wantTypes), clauses)
			}
			for i, wantType := range tt.wantTypes {
				if clauses[i].Kind != string(wantType) {
					t.Fatalf("expected clause %d type %q, got %q", i, wantType, clauses[i].Kind)
				}
			}
		})
	}
}

func TestParsePreservesInvalidEARSContext(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantClauseCount int
	}{
		{
			name: "duplicate when",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
when Worker reports Hazard
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
			wantClauseCount: 2,
		},
		{
			name: "when after if",
			input: `stakeholder Worker

requirement R1
if Worker has_no_consent
when Worker enters DangerousArea
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
			wantClauseCount: 2,
		},
		{
			name: "where before if",
			input: `stakeholder Worker

requirement R1
where SafetyMode enabled
if Worker has_no_consent
system shall notify Worker using AudioAlarm
stakeholders Worker
`,
			wantClauseCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse(tt.input).AST
			reqs := requirements(doc)
			if len(reqs) != 1 || len(requirementContextNodes(reqs[0])) != tt.wantClauseCount {
				t.Fatalf("expected %d preserved clauses, got %+v", tt.wantClauseCount, reqs)
			}
		})
	}
}

func TestParsePreservesInvalidRequirementClauseOrder(t *testing.T) {
	input := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
while Worker is in DangerousArea
system shall notify Worker using AudioAlarm
stakeholders Worker
`

	doc := Parse(input).AST
	clauses := requirementContextNodes(requirements(doc)[0])
	if len(clauses) != 2 {
		t.Fatalf("expected 2 clauses, got %+v", clauses)
	}
	if clauses[0].Kind != string(grammar.RequirementClauseWhen) || clauses[1].Kind != string(grammar.RequirementClauseWhile) {
		t.Fatalf("expected parser to preserve invalid order, got %+v", clauses)
	}
}

func TestParsePreservesIncompleteRequirementClauses(t *testing.T) {
	input := `stakeholder Worker

requirement R1
where
when Worker enters
system shall notify Worker using AudioAlarm
stakeholders Worker
`

	doc := Parse(input).AST
	reqs := requirements(doc)
	if len(reqs) != 1 {
		t.Fatalf("expected one requirement, got %+v", reqs)
	}
	clauses := requirementContextNodes(reqs[0])
	if len(clauses) != 2 {
		t.Fatalf("expected preserved incomplete clauses, got %+v", clauses)
	}
	if clauses[0].Kind != string(grammar.RequirementClauseWhere) || strings.TrimSpace(clauses[0].First(ast.FieldRawText)) != "" {
		t.Fatalf("expected empty where clause to be preserved, got %+v", clauses[0])
	}
	if clauses[1].Kind != string(grammar.RequirementClauseWhen) || strings.TrimSpace(clauses[1].First(ast.FieldRawText)) != "Worker enters" {
		t.Fatalf("expected partial when clause to be preserved, got %+v", clauses[1])
	}
}

func TestParsePreservesIncompleteActionAndStakeholdersLines(t *testing.T) {
	input := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall notify
stakeholders
`

	doc := Parse(input).AST
	reqs := requirements(doc)
	if len(reqs) != 1 {
		t.Fatalf("expected one requirement, got %+v", reqs)
	}
	req := reqs[0]
	action := firstRequirementEntry(t, req, grammar.RequirementClauseSystemShall)
	if action.First("verb") != "notify" || action.First("target") != "" {
		t.Fatalf("expected partial action to be preserved, got %+v", action)
	}
	sh := firstRequirementEntry(t, req, grammar.RequirementClauseStakeholders)
	if sh.Line == 0 || len(sh.Texts("actors")) != 0 {
		t.Fatalf("expected empty stakeholders clause to be preserved, got %+v", sh)
	}
}

func TestParseAssignmentBlocks(t *testing.T) {
	input := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91
value authority_pref = 0.02, 0.88

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager

requirement R2
when Manager reports Incident
system shall notify Worker using AudioAlarm
stakeholders Worker, Manager

assignment R1
Worker -> privacy_pref
Manager -> authority_pref

assignment R2
Worker -> privacy_pref
`

	doc := Parse(input).AST

	assigns := assignments(doc)
	if len(assigns) != 2 {
		t.Fatalf("expected 2 assignment blocks, got %d", len(assigns))
	}

	first := assigns[0]
	if first.Line != 17 {
		t.Fatalf("expected first assignment block on line 17, got %d", first.Line)
	}
	if first.Header.First("requirement") != "R1" {
		t.Fatalf("expected first assignment requirement R1, got %q", first.Header.First("requirement"))
	}
	if len(first.Body) != 2 {
		t.Fatalf("expected 2 entries in first assignment block, got %d", len(first.Body))
	}
	if first.Body[0].Line != 18 || first.Body[0].Joined("stakeholders", ", ") != "Worker" || first.Body[0].Joined("values", ", ") != "privacy_pref" {
		t.Fatalf("unexpected first assignment entry: %+v", first.Body[0])
	}
	if first.Body[1].Line != 19 || first.Body[1].Joined("stakeholders", ", ") != "Manager" || first.Body[1].Joined("values", ", ") != "authority_pref" {
		t.Fatalf("unexpected second assignment entry: %+v", first.Body[1])
	}

	second := assigns[1]
	if second.Header.First("requirement") != "R2" {
		t.Fatalf("expected second assignment requirement R2, got %q", second.Header.First("requirement"))
	}
	if len(second.Body) != 1 {
		t.Fatalf("expected 1 entry in second assignment block, got %d", len(second.Body))
	}
	if second.Body[0].Joined("stakeholders", ", ") != "Worker" || second.Body[0].Joined("values", ", ") != "privacy_pref" {
		t.Fatalf("unexpected third assignment entry: %+v", second.Body[0])
	}
}

func TestParseRejectsInvalidAssignmentBlocks(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name: "missing arrow",
			input: `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker

assignment R1
Worker privacy
`,
			wantErr: "parse DSL:",
		},
		{
			name: "extra tokens after value",
			input: `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker

assignment R1
Worker -> privacy_pref immediately
`,
			wantErr: "parse DSL:",
		},
		{
			name: "empty assignment block",
			input: `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker

assignment R1
`,
			wantErr: "parse DSL:",
		},
		{
			name: "reserved keyword stakeholder",
			input: `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker

assignment R1
when -> privacy_pref
`,
			wantErr: `assignment R1 stakeholder "when" must not use a reserved keyword`,
		},
		{
			name: "reserved keyword value",
			input: `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker

assignment R1
Worker -> using
`,
			wantErr: `assignment R1 value "using" must not use a reserved keyword`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse(tt.input).AST
			if len(assignments(doc)) != 1 {
				t.Fatalf("expected assignment block to be preserved, got %+v", assignments(doc))
			}
		})
	}
}

func TestParseRequirementMetadata(t *testing.T) {
	input := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
priority high
retention short_term
access supervisor_only
`

	doc := Parse(input).AST

	req := requirements(doc)[0]
	if entry := firstRequirementEntry(t, req, grammar.RequirementClausePriority); entry.Line != 7 || entry.First("level") != "high" {
		t.Fatalf("unexpected priority metadata: %+v", entry)
	}
	if entry := firstRequirementEntry(t, req, grammar.RequirementClauseRetention); entry.Line != 8 || entry.First("value") != "short_term" {
		t.Fatalf("unexpected retention metadata: %+v", entry)
	}
	if entry := firstRequirementEntry(t, req, grammar.RequirementClauseAccess); entry.Line != 9 || entry.First("value") != "supervisor_only" {
		t.Fatalf("unexpected access metadata: %+v", entry)
	}
}

func TestParsePreservesInvalidMetadataClauses(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "metadata on stakeholders line",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker priority high
`,
		},
		{
			name: "metadata before stakeholders",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
priority high
stakeholders Worker
`,
		},
		{
			name: "duplicate metadata clause",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
priority high
priority urgent
`,
		},
		{
			name: "multi token metadata value",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
priority very high
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse(tt.input).AST
			if len(requirements(doc)) != 1 {
				t.Fatalf("expected requirement to be preserved, got %+v", requirements(doc))
			}
		})
	}
}

func TestParseRequirementTraceability(t *testing.T) {
	input := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
priority high
retention short_term
access supervisor_only
linked_to HazardAnalysis
linked_to UseCase_42
`

	doc := Parse(input).AST

	req := requirements(doc)[0]
	traceability := requirementEntries(req, grammar.RequirementClauseLinkedTo)
	if len(traceability) != 2 {
		t.Fatalf("expected 2 traceability links, got %+v", traceability)
	}
	if traceability[0].Line != 10 || traceability[0].First("target") != "HazardAnalysis" {
		t.Fatalf("unexpected first traceability link: %+v", traceability[0])
	}
	if traceability[1].Line != 11 || traceability[1].First("target") != "UseCase_42" {
		t.Fatalf("unexpected second traceability link: %+v", traceability[1])
	}
}

func requirementContextNodes(req ast.DeclarationNode) []ast.GenericNode {
	out := make([]ast.GenericNode, 0)
	for _, node := range req.Body {
		switch node.Kind {
		case string(grammar.RequirementClauseWhile), string(grammar.RequirementClauseWhen), string(grammar.RequirementClauseIf), string(grammar.RequirementClauseWhere):
			out = append(out, node)
		}
	}
	return out
}

func TestParsePreservesInvalidTraceabilityClauses(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "linked_to before stakeholders",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
linked_to HazardAnalysis
stakeholders Worker
`,
		},
		{
			name: "linked_to before later metadata",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
linked_to HazardAnalysis
priority high
`,
		},
		{
			name: "multi token linked_to value",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
linked_to Hazard Analysis
`,
		},
		{
			name: "linked_to on access line",
			input: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
access supervisor_only linked_to HazardAnalysis
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := Parse(tt.input).AST
			if len(requirements(doc)) != 1 {
				t.Fatalf("expected requirement to be preserved, got %+v", requirements(doc))
			}
		})
	}
}

func readExampleDSL(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "examples", "dsl", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}

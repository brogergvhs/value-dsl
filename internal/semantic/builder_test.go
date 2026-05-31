package semantic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/parser"
)

func fieldValue(text string) ast.FieldValue {
	return ast.FieldValue{Text: text}
}

func stakeholderDecl(line int, names ...string) ast.DeclarationNode {
	fvs := make([]ast.FieldValue, 0, len(names))
	for _, name := range names {
		fvs = append(fvs, fieldValue(name))
	}
	return ast.DeclarationNode{
		Kind: grammar.DeclarationKindStakeholder,
		Line: line,
		Header: ast.GenericNode{
			Kind: string(grammar.DeclarationKindStakeholder),
			Line: line,
			Fields: map[string][]ast.FieldValue{
				"names": fvs,
			},
		},
	}
}

func valueDecl(line int, name, angle, radius string) ast.DeclarationNode {
	fields := map[string][]ast.FieldValue{
		"name": {fieldValue(name)},
	}
	if angle != "" {
		fields["angle"] = []ast.FieldValue{fieldValue(angle)}
	}
	if radius != "" {
		fields["radius"] = []ast.FieldValue{fieldValue(radius)}
	}
	if angle != "" || radius != "" {
		fields[ast.FieldKeyword] = []ast.FieldValue{fieldValue("="), fieldValue(",")}
	}
	return ast.DeclarationNode{
		Kind: grammar.DeclarationKindValue,
		Line: line,
		Header: ast.GenericNode{
			Kind:   string(grammar.DeclarationKindValue),
			Line:   line,
			Fields: fields,
		},
	}
}

func requirementDecl(line int, id string, body ...ast.GenericNode) ast.DeclarationNode {
	return ast.DeclarationNode{
		Kind: grammar.DeclarationKindRequirement,
		Line: line,
		Header: ast.GenericNode{
			Kind: string(grammar.DeclarationKindRequirement),
			Line: line,
			Fields: map[string][]ast.FieldValue{
				"id": {fieldValue(id)},
			},
		},
		Body: body,
	}
}

func assignmentDecl(line int, reqID string, entries ...ast.GenericNode) ast.DeclarationNode {
	return ast.DeclarationNode{
		Kind: grammar.DeclarationKindAssignment,
		Line: line,
		Header: ast.GenericNode{
			Kind: string(grammar.DeclarationKindAssignment),
			Line: line,
			Fields: map[string][]ast.FieldValue{
				"requirement": {fieldValue(reqID)},
			},
		},
		Body: entries,
	}
}

func whenClauseNode(line int, actor, verb, target string) ast.GenericNode {
	return ast.GenericNode{
		Kind: string(grammar.RequirementClauseWhen),
		Line: line,
		Fields: map[string][]ast.FieldValue{
			"actor":          {fieldValue(actor)},
			"verb":           {fieldValue(verb)},
			"target":         {fieldValue(target)},
			ast.FieldRawText: {fieldValue(actor + " " + verb + " " + target)},
		},
	}
}

func actionEntry(line int, verb, object, target, mechanism, raw string) ast.GenericNode {
	fields := map[string][]ast.FieldValue{
		"verb":           {fieldValue(verb)},
		ast.FieldRawText: {fieldValue(raw)},
	}
	if object != "" {
		fields["object"] = []ast.FieldValue{fieldValue(object)}
		fields[ast.FieldKeyword] = append(fields[ast.FieldKeyword], fieldValue("of"))
	}
	if target != "" {
		fields["target"] = []ast.FieldValue{fieldValue(target)}
	}
	if mechanism != "" {
		fields["mechanism"] = []ast.FieldValue{fieldValue(mechanism)}
		fields[ast.FieldKeyword] = append(fields[ast.FieldKeyword], fieldValue("using"))
	}
	return ast.GenericNode{
		Kind:   string(grammar.RequirementClauseSystemShall),
		Line:   line,
		Fields: fields,
	}
}

func stakeholdersEntry(line int, names ...string) ast.GenericNode {
	fields := make([]ast.FieldValue, 0, len(names))
	for _, name := range names {
		fields = append(fields, fieldValue(name))
	}
	return ast.GenericNode{
		Kind: string(grammar.RequirementClauseStakeholders),
		Line: line,
		Fields: map[string][]ast.FieldValue{
			"actors": fields,
		},
	}
}

func metadataEntry(kind grammar.RequirementClauseKind, line int, value string) ast.GenericNode {
	fieldName := "value"
	if kind == grammar.RequirementClausePriority {
		fieldName = "level"
	}
	if kind == grammar.RequirementClauseLinkedTo {
		fieldName = "target"
	}
	return ast.GenericNode{
		Kind: string(kind),
		Line: line,
		Fields: map[string][]ast.FieldValue{
			fieldName: {fieldValue(value)},
		},
	}
}

func TestBuildSemanticModel(t *testing.T) {
	doc := &ast.DocumentNode{
		Declarations: []ast.DeclarationNode{
			stakeholderDecl(1, "Worker"),
			stakeholderDecl(2, "Manager"),
			valueDecl(4, "privacy_pref", "1.57", "0.92"),
			valueDecl(5, "authority_pref", "0.00", "0.90"),
			requirementDecl(7, "R1",
				actionEntry(9, "track", "location", "Worker", "Camera", "track location of Worker using Camera"),
				stakeholdersEntry(10, "Worker", "Manager"),
				whenClauseNode(8, "Worker", "enters", "DangerousArea"),
			),
			assignmentDecl(12, "R1",
				ast.GenericNode{Line: 13, Fields: map[string][]ast.FieldValue{
					"stakeholders":   {fieldValue("Worker")},
					"values":         {fieldValue("privacy_pref")},
					ast.FieldKeyword: {fieldValue("->")},
				}},
			),
		},
	}

	semanticModel, err := Build(doc)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	req := semanticModel.Requirements[0]
	if req.ID != "R1" || len(req.Context.Clauses) != 1 {
		t.Fatalf("unexpected requirement: %+v", req)
	}
	if req.Context.Clauses[0].Actor != "Worker" || req.Context.Clauses[0].Verb != "enters" {
		t.Fatalf("unexpected context clause: %+v", req.Context.Clauses[0])
	}
	if req.Action.Verb != "track" || req.Action.Object != "location" || req.Action.Target != "Worker" || req.Action.Mechanism != "Camera" {
		t.Fatalf("unexpected action: %+v", req.Action)
	}
	if req.StakeholdersLine != 10 || len(req.Stakeholders) != 2 {
		t.Fatalf("unexpected stakeholders: %+v", req.Stakeholders)
	}
	if len(semanticModel.AssignmentEntries) != 1 || semanticModel.AssignmentEntries[0].HeaderLine != 12 {
		t.Fatalf("unexpected assignments: %+v", semanticModel.AssignmentEntries)
	}
}

func TestBuildUnknownStakeholderReferencePreservesPlaceholder(t *testing.T) {
	doc := &ast.DocumentNode{
		Declarations: []ast.DeclarationNode{
			stakeholderDecl(1, "Worker"),
			requirementDecl(3, "R1",
				whenClauseNode(4, "Worker", "enters", "DangerousArea"),
				actionEntry(5, "track", "location", "Worker", "", "track location of Worker"),
				stakeholdersEntry(6, "Manager"),
			),
		},
	}

	semanticModel, err := Build(doc)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(semanticModel.Requirements[0].Stakeholders) != 1 || semanticModel.Requirements[0].Stakeholders[0].Name != "Manager" {
		t.Fatalf("expected unresolved stakeholder placeholder, got %+v", semanticModel.Requirements[0].Stakeholders)
	}
}

func TestBuildRequirementMetadataAndTraceability(t *testing.T) {
	doc := &ast.DocumentNode{
		Declarations: []ast.DeclarationNode{
			stakeholderDecl(1, "Worker"),
			requirementDecl(3, "R1",
				whenClauseNode(4, "Worker", "enters", "DangerousArea"),
				actionEntry(5, "track", "location", "Worker", "Camera", "track location of Worker using Camera"),
				stakeholdersEntry(6, "Worker"),
				metadataEntry(grammar.RequirementClausePriority, 7, "high"),
				metadataEntry(grammar.RequirementClauseRetention, 8, "short_term"),
				metadataEntry(grammar.RequirementClauseAccess, 9, "supervisor_only"),
				metadataEntry(grammar.RequirementClauseLinkedTo, 10, "HazardAnalysis"),
				metadataEntry(grammar.RequirementClauseLinkedTo, 11, "UseCase_42"),
			),
		},
	}

	semanticModel, err := Build(doc)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	req := semanticModel.Requirements[0]
	if req.Metadata.Priority.Value != "high" || req.Metadata.Retention.Value != "short_term" || req.Metadata.Access.Value != "supervisor_only" {
		t.Fatalf("unexpected metadata: %+v", req.Metadata)
	}
	if len(req.Traceability) != 2 || req.Traceability[0].Value != "HazardAnalysis" || req.Traceability[1].Value != "UseCase_42" {
		t.Fatalf("unexpected traceability: %+v", req.Traceability)
	}
}

func TestBuildNilDocument(t *testing.T) {
	_, err := Build(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "semantic error: document is nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildInvalidExampleUnknownContextActorProducesPartialModel(t *testing.T) {
	input := readBuilderExampleDSL(t, "29_invalid_unknown_context_actor.dsl")

	doc := parser.Parse(input).AST

	semanticModel, err := Build(doc)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(semanticModel.Requirements) != 1 || semanticModel.Requirements[0].Context.Clauses[0].Actor != "Manager" {
		t.Fatalf("expected invalid example to preserve unresolved actor, got %+v", semanticModel.Requirements)
	}
}

func TestBuildAssignmentEntriesStayGrouped(t *testing.T) {
	doc := &ast.DocumentNode{
		Declarations: []ast.DeclarationNode{
			stakeholderDecl(1, "Worker"),
			stakeholderDecl(2, "Manager"),
			valueDecl(4, "privacy_pref", "1.57", "0.92"),
			valueDecl(5, "authority_pref", "0.00", "0.90"),
			requirementDecl(7, "R1",
				actionEntry(9, "track", "location", "Worker", "", "track location of Worker"),
				stakeholdersEntry(10, "Worker", "Manager"),
			),
			assignmentDecl(12, "R1",
				ast.GenericNode{Line: 13, Fields: map[string][]ast.FieldValue{
					"stakeholders":   {fieldValue("Worker"), fieldValue("Manager")},
					"values":         {fieldValue("privacy_pref"), fieldValue("authority_pref")},
					ast.FieldKeyword: {fieldValue("->")},
				}},
			),
		},
	}

	m, err := Build(doc)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(m.AssignmentEntries) != 1 {
		t.Fatalf("expected 1 grouped assignment entry, got %d: %+v", len(m.AssignmentEntries), m.AssignmentEntries)
	}
	entry := m.AssignmentEntries[0]
	if len(entry.Stakeholders) != 2 || entry.Stakeholders[0].Name != "Worker" || entry.Stakeholders[1].Name != "Manager" {
		t.Fatalf("unexpected stakeholders: %+v", entry.Stakeholders)
	}
	if len(entry.Values) != 2 || entry.Values[0].Name != "privacy_pref" || entry.Values[1].Name != "authority_pref" {
		t.Fatalf("unexpected values: %+v", entry.Values)
	}
}

func readBuilderExampleDSL(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "examples", "dsl", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}

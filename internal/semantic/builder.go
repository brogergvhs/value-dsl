// Package semantic converts AST nodes into domain model objects.
// It is the only place that should interpret generic AST entries as DSL domain
// concepts such as requirements, actions, stakeholders, values, and assignments.
package semantic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/values"
)

// SemanticModel is the domain-level representation produced from AST input
type SemanticModel struct {
	Requirements      []model.Requirement
	Stakeholders      []model.Stakeholder
	Values            []model.Value
	AssignmentEntries []model.AssignmentEntry

	StakeholderByName map[string]model.Stakeholder
	ValueByName       map[string]model.Value
	RequirementByID   map[string]model.Requirement
}

func (m *SemanticModel) LookupRequirement(id string) (model.Requirement, bool) {
	if m.RequirementByID != nil {
		r, ok := m.RequirementByID[id]
		return r, ok
	}
	for _, r := range m.Requirements {
		if r.ID == id {
			return r, true
		}
	}
	return model.Requirement{}, false
}

// Build converts AST into a semantic model and resolves references.
// It preserves as much structure as possible and leaves user-facing
// invalidity to validation instead of returning recoverable build errors.
func Build(doc *ast.DocumentNode) (*SemanticModel, error) {
	if doc == nil {
		return nil, fmt.Errorf("semantic error: document is nil")
	}
	if err := values.RegistryError(); err != nil {
		return nil, fmt.Errorf("semantic error: values registry: %w", err)
	}

	stakeholders, stakeholderByName := buildStakeholders(doc.DeclarationsOfKind(grammar.DeclarationKindStakeholder))
	vals, valueByName := buildValues(doc.DeclarationsOfKind(grammar.DeclarationKindValue))
	requirements, requirementByID := buildRequirements(doc.DeclarationsOfKind(grammar.DeclarationKindRequirement), stakeholderByName)
	assignments := buildAssignments(doc.DeclarationsOfKind(grammar.DeclarationKindAssignment))

	return &SemanticModel{
		Requirements:      requirements,
		Stakeholders:      stakeholders,
		Values:            vals,
		AssignmentEntries: assignments,
		StakeholderByName: stakeholderByName,
		ValueByName:       valueByName,
		RequirementByID:   requirementByID,
	}, nil
}

// buildStakeholders expands multi-name declarations into individual model objects
// and builds a name lookup map (first occurrence wins).
func buildStakeholders(decls []ast.DeclarationNode) ([]model.Stakeholder, map[string]model.Stakeholder) {
	stakeholders := make([]model.Stakeholder, 0, len(decls))
	byName := make(map[string]model.Stakeholder, len(decls))

	for _, decl := range decls {
		fieldValues := decl.Header.Values("names")
		if len(fieldValues) == 0 {
			stakeholders = append(stakeholders, model.Stakeholder{
				Line:  decl.Line,
				Range: decl.Range,
			})
			continue
		}

		for _, fv := range fieldValues {
			name := strings.TrimSpace(fv.Text)
			s := model.Stakeholder{Line: decl.Line, Range: fv.Range, Name: name}
			if name != "" {
				if _, exists := byName[name]; !exists {
					byName[name] = s
				}
			}
			stakeholders = append(stakeholders, s)
		}
	}

	return stakeholders, byName
}

// buildValues converts value declarations into model objects, resolving angle/radius
// geometry and mapping angles to the closest value category.
func buildValues(decls []ast.DeclarationNode) ([]model.Value, map[string]model.Value) {
	outValues := make([]model.Value, 0, len(decls))
	byName := make(map[string]model.Value, len(decls))

	for _, decl := range decls {
		name := strings.TrimSpace(decl.Header.First("name"))
		value := valueFromDecl(decl, name)
		if value.Name != "" {
			if _, exists := byName[value.Name]; !exists {
				byName[value.Name] = value
			}
		}
		outValues = append(outValues, value)
	}

	return outValues, byName
}

func valueFromDecl(decl ast.DeclarationNode, name string) model.Value {
	angleText := decl.Header.First("angle")
	radiusText := decl.Header.First("radius")
	hasComma := decl.Header.HasText(ast.FieldKeyword, ",")

	angle, hasAngle := parseOptionalNumber(angleText)
	radius, hasRadius := parseOptionalNumber(radiusText)

	categoryName := ""
	if hasAngle {
		if category, ok := values.ClosestCategory(angle); ok {
			categoryName = category.Name
		}
	}

	return model.Value{
		Line:          decl.Line,
		Range:         decl.Range,
		Name:          name,
		Category:      categoryName,
		Angle:         angle,
		Radius:        radius,
		HasAngle:      hasAngle,
		HasRadius:     hasRadius,
		HasValueComma: hasComma,
	}
}

// buildRequirements assembles requirements from their header (id) and body clauses
// (context, action, stakeholders, metadata, traceability).
func buildRequirements(decls []ast.DeclarationNode, stakeholderByName map[string]model.Stakeholder) ([]model.Requirement, map[string]model.Requirement) {
	requirements := make([]model.Requirement, 0, len(decls))
	byID := make(map[string]model.Requirement, len(decls))
	for _, decl := range decls {
		r := buildRequirement(decl, stakeholderByName)
		requirements = append(requirements, r)
		if r.ID != "" {
			if _, exists := byID[r.ID]; !exists {
				byID[r.ID] = r
			}
		}
	}
	return requirements, byID
}

func buildRequirement(decl ast.DeclarationNode, stakeholderByName map[string]model.Stakeholder) model.Requirement {
	requirement := model.Requirement{
		Line:  decl.Line,
		Range: decl.Range,
		ID:    strings.TrimSpace(decl.Header.First("id")),
	}
	context := make([]model.Clause, 0, len(decl.Body))
	stakeholderNames := []string(nil)

	for _, node := range decl.Body {
		switch grammar.RequirementClauseKind(node.Kind) {
		case grammar.RequirementClauseWhen, grammar.RequirementClauseWhile, grammar.RequirementClauseIf, grammar.RequirementClauseWhere:
			context = append(context, clauseFromNode(node))
		case grammar.RequirementClauseSystemShall:
			requirement.Action = model.Action{
				Line:      node.Line,
				Range:     node.Range,
				Verb:      strings.TrimSpace(node.First("verb")),
				Object:    strings.TrimSpace(node.First("object")),
				Target:    strings.TrimSpace(node.First("target")),
				Mechanism: strings.TrimSpace(node.First("mechanism")),
				RawText:   strings.TrimSpace(node.First(ast.FieldRawText)),
			}
		case grammar.RequirementClauseStakeholders:
			requirement.StakeholdersLine = node.Line
			requirement.StakeholdersRange = node.Range
			stakeholderNames = node.Texts("actors")
		case grammar.RequirementClausePriority:
			requirement.Metadata.Priority = metadataField(node)
		case grammar.RequirementClauseRetention:
			requirement.Metadata.Retention = metadataField(node)
		case grammar.RequirementClauseAccess:
			requirement.Metadata.Access = metadataField(node)
		case grammar.RequirementClauseLinkedTo:
			requirement.Traceability = append(requirement.Traceability,
				model.TraceabilityLink{Line: node.Line, Range: node.Range, Value: strings.TrimSpace(node.FirstNonMetaField())})
		}
	}

	if len(context) > 0 {
		requirement.Context = model.RequirementContext{Clauses: context}
	}

	requirement.Stakeholders = stakeholderRefs(stakeholderNames, stakeholderByName)
	return requirement
}

func metadataField(node ast.GenericNode) model.MetadataField {
	return model.MetadataField{
		Line: node.Line, Range: node.Range,
		Value: strings.TrimSpace(ast.FirstNonEmpty(node.First("level"), node.First("value"), node.First("target"))),
	}
}

// buildAssignments keeps assignment lines grouped instead of eagerly expanding
// cartesian stakeholder/value pairs into individual rows.
func buildAssignments(decls []ast.DeclarationNode) []model.AssignmentEntry {
	sourceField, targetField := "", ""
	if rule := grammar.Compiled.AssignmentEntryRule; rule != nil {
		sourceField, _ = rule.SemanticField(grammar.SemanticRoleAssignmentSource)
		targetField, _ = rule.SemanticField(grammar.SemanticRoleAssignmentTarget)
	}

	assignments := make([]model.AssignmentEntry, 0, len(decls))

	for _, decl := range decls {
		reqID := strings.TrimSpace(decl.Header.First("requirement"))
		if len(decl.Body) == 0 {
			assignments = append(assignments, model.AssignmentEntry{
				HeaderLine:    decl.Line,
				RequirementID: reqID,
			})
			continue
		}

		for _, entry := range decl.Body {
			assignments = append(assignments, model.AssignmentEntry{
				Line:          entry.Line,
				HeaderLine:    decl.Line,
				RequirementID: reqID,
				Stakeholders:  refs(entry.Texts(sourceField)),
				Values:        refs(entry.Texts(targetField)),
			})
		}
	}

	return assignments
}

func stakeholderRefs(names []string, byName map[string]model.Stakeholder) []model.Stakeholder {
	out := make([]model.Stakeholder, 0, len(names))
	for _, name := range names {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		if s, ok := byName[name]; ok {
			out = append(out, s)
		} else {
			out = append(out, model.Stakeholder{Name: name, Invalid: true})
		}
	}

	return out
}

func refs(names []string) []model.Ref {
	out := make([]model.Ref, 0, len(names))
	for _, name := range names {
		if name = strings.TrimSpace(name); name != "" {
			out = append(out, model.Ref{Name: name})
		}
	}

	return out
}

func clauseFromNode(node ast.GenericNode) model.Clause {
	return model.Clause{
		Line:    node.Line,
		Range:   node.Range,
		Type:    grammar.RequirementClauseKind(node.Kind),
		Actor:   strings.TrimSpace(node.First("actor")),
		Verb:    strings.TrimSpace(node.First("verb")),
		Target:  strings.TrimSpace(ast.FirstNonEmpty(node.First("target"), node.First("rest"))),
		RawText: strings.TrimSpace(node.First(ast.FieldRawText)),
	}
}

func parseOptionalNumber(text string) (float64, bool) {
	if text == "" {
		return 0, false
	}
	if v, err := strconv.ParseFloat(text, 64); err == nil {
		return v, true
	}
	return 0, false
}

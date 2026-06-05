package lsp

import (
	"fmt"
	"strings"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/values"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// hover handles the textDocument/hover request.
func (s *Server) hover(_ *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	return resolveHover(document.Text, params.Position, s.navigationAnalysis(document.URI, result)), nil
}

func resolveHover(text string, position protocol.Position, result coreanalysis.Result) *protocol.Hover {
	lines := tokenLines(text, result)
	token, hoverRange, ok := tokenAtPosition(lines, position)
	if !ok {
		return nil
	}

	line := ""
	if idx := int(position.Line); idx >= 0 && idx < len(lines) {
		line = lines[idx].Code
	}
	if contents, ok := symbolHoverDocumentation(token, result.Index); ok {
		return newHover(contents, hoverRange)
	}
	if contents, ok := contextualHoverDocumentation(token, line); ok {
		return newHover(contents, hoverRange)
	}
	if contents, ok := tokenHoverDocumentation(token); ok {
		return newHover(contents, hoverRange)
	}

	return nil
}

func symbolHoverDocumentation(token string, index *docindex.Document) (string, bool) {
	if index == nil || index.Model == nil {
		return "", false
	}

	model := index.Model
	if stakeholder, ok := model.StakeholderByName[token]; ok {
		return hoverForStakeholder(stakeholder, model.Requirements, model.AssignmentEntries), true
	}
	if valueDecl, ok := model.ValueByName[token]; ok {
		return hoverForValue(valueDecl), true
	}
	if requirement, ok := model.LookupRequirement(token); ok {
		return hoverForRequirement(requirement), true
	}

	return "", false
}

func newHover(contents string, hoverRange protocol.Range) *protocol.Hover {
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: contents,
		},
		Range: &hoverRange,
	}
}

func tokenAtPosition(lines []grammar.TokenLine, position protocol.Position) (string, protocol.Range, bool) {
	line, byteOffset, ok := lineAndByteOffset(lines, position)
	if !ok {
		return "", protocol.Range{}, false
	}
	visibleOffset, inComment := line.VisibleOffset(byteOffset)
	if inComment {
		return "", protocol.Range{}, false
	}
	token, ok := line.TokenAtOffset(visibleOffset)
	if !ok || !grammar.Compiled.IsValidIdentifier(token.Text) {
		return "", protocol.Range{}, false
	}
	startUTF16, ok := sourcepos.UTF16OffsetFromByteOK(line.Raw, int(token.StartColumn)-1)
	if !ok {
		return "", protocol.Range{}, false
	}
	endUTF16, ok := sourcepos.UTF16OffsetFromByteOK(line.Raw, int(token.EndColumn)-1)
	if !ok {
		return "", protocol.Range{}, false
	}
	return token.Text, protocol.Range{
		Start: protocol.Position{Line: position.Line, Character: protocol.UInteger(startUTF16)},
		End:   protocol.Position{Line: position.Line, Character: protocol.UInteger(endUTF16)},
	}, true
}

func tokenHoverDocumentation(token string) (string, bool) {
	if info, ok := grammar.TokenInfoFor(token); ok && info.Documentation != "" {
		return info.Documentation, true
	}

	if value, ok := values.LookupByName(token); ok {
		return fmt.Sprintf(
			"Built-in Schwartz value.\n\nCategory: `%s`\n\nAngle: `%.2f rad`.\n\nRadius: `%.2f`.",
			value.Category,
			value.Angle,
			value.Radius,
		), true
	}

	for _, category := range values.Categories() {
		if category.Name == token {
			return fmt.Sprintf(
				"Schwartz category.\n\n%s\n\nAngle: `%.2f rad`.",
				category.Description,
				category.Angle,
			), true
		}
	}

	return "", false
}

func hoverForStakeholder(stakeholder model.Stakeholder, requirements []model.Requirement, assignments []model.AssignmentEntry) string {
	var requirementRefs, contextRefs, assignmentRefs []string

	for _, requirement := range requirements {
		for _, reqStakeholder := range requirement.Stakeholders {
			if reqStakeholder.Name == stakeholder.Name {
				requirementRefs = append(requirementRefs, requirement.ID)
				break
			}
		}
		for _, clause := range requirement.Context.Clauses {
			if clause.Actor == stakeholder.Name {
				contextRefs = append(contextRefs, requirement.ID)
				break
			}
		}
	}

	for _, assignment := range assignments {
		for _, r := range assignment.Stakeholders {
			if r.Name == stakeholder.Name {
				assignmentRefs = append(assignmentRefs, assignment.RequirementID)
				break
			}
		}
	}

	return fmt.Sprintf(
		"**stakeholder** `%s`\n\nRequirement clauses: %s\n\nContext actor in: %s\n\nAssignments in: %s",
		stakeholder.Name,
		formatListOrNone(uniqueStrings(requirementRefs)),
		formatListOrNone(uniqueStrings(contextRefs)),
		formatListOrNone(uniqueStrings(assignmentRefs)),
	)
}

func hoverForValue(v model.Value) string {
	hasGeom := func() bool {
		if v.HasAngle || v.HasRadius || v.HasValueComma {
			return v.HasAngle && v.HasRadius
		}
		return v.Angle != 0 || v.Radius != 0 || strings.TrimSpace(v.Category) != ""
	}()

	switch {
	case v.Builtin:
		return fmt.Sprintf("**built-in value** `%s`\n\nCategory: `%s`\n\nAngle: `%.2f rad`\n\nRadius: `%.2f`", v.Name, v.Category, v.Angle, v.Radius)
	case v.Invalid || !hasGeom:
		return fmt.Sprintf("**invalid custom value** `%s`\n\nThis value declaration is incomplete or invalid and cannot be used until both angle and radius are defined correctly.", v.Name)
	default:
		return fmt.Sprintf("**custom value** `%s`\n\nDerived Schwartz category: `%s`\n\nAngle: `%.2f rad`\n\nRadius: `%.2f`", v.Name, v.Category, v.Angle, v.Radius)
	}
}

func hoverForRequirement(requirement model.Requirement) string {
	contextKinds := make([]string, 0, len(requirement.Context.Clauses))
	for _, clause := range requirement.Context.Clauses {
		contextKinds = append(contextKinds, string(clause.Type))
	}

	stakeholders := make([]string, 0, len(requirement.Stakeholders))
	for _, s := range requirement.Stakeholders {
		stakeholders = append(stakeholders, s.Name)
	}

	contextSummary := "ubiquitous"
	if len(contextKinds) > 0 {
		contextSummary = strings.Join(contextKinds, ", ")
	}

	meta := requirement.Metadata
	var metadata []string
	if meta.Priority.Value != "" {
		metadata = append(metadata, "priority="+meta.Priority.Value)
	}
	if meta.Retention.Value != "" {
		metadata = append(metadata, "retention="+meta.Retention.Value)
	}
	if meta.Access.Value != "" {
		metadata = append(metadata, "access="+meta.Access.Value)
	}

	traceability := make([]string, 0, len(requirement.Traceability))
	for _, link := range requirement.Traceability {
		traceability = append(traceability, link.Value)
	}

	return fmt.Sprintf(
		"**requirement** `%s`\n\nContext: `%s`\n\nAction: `%s`\n\nStakeholders: %s\n\nMetadata: %s\n\nTraceability: %s",
		requirement.ID,
		contextSummary,
		requirement.Action.RawText,
		strings.Join(stakeholders, ", "),
		formatListOrNone(metadata),
		formatListOrNone(traceability),
	)
}

var metadataHoverFormats = []struct{ prefix, format string }{
	{"priority ", "Priority value `%s` for the requirement."},
	{"retention ", "Retention value `%s` for the requirement."},
	{"access ", "Access-scope value `%s` for the requirement."},
	{"linked_to ", "Traceability target `%s` linked from this requirement."},
}

func contextualHoverDocumentation(token, line string) (string, bool) {
	t := strings.TrimSpace(line)
	for _, entry := range metadataHoverFormats {
		if strings.HasPrefix(t, entry.prefix) {
			return fmt.Sprintf(entry.format, token), true
		}
	}
	return "", false
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; s != "" && !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}

	return out
}

func formatListOrNone(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

package lsp

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/values"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// codeAction handles the textDocument/codeAction request.
func (s *Server) codeAction(_ *glsp.Context, params *protocol.CodeActionParams) (any, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok || result.Index == nil || result.Index.Model == nil {
		return []protocol.CodeAction{}, nil
	}
	return codeActions(protocol.DocumentUri(document.URI), result, params), nil
}

func codeActions(uri protocol.DocumentUri, result coreanalysis.Result, params *protocol.CodeActionParams) []protocol.CodeAction {
	if !wantsQuickFix(params.Context.Only) {
		return nil
	}

	var actions []protocol.CodeAction
	if edit := missingStakeholdersEdit(uri, result); len(edit.Changes) > 0 {
		actions = append(actions, quickFix("Add missing stakeholders", edit))
	}
	actions = append(actions, duplicateValueActions(uri, result)...)
	actions = append(actions, assignmentActions(uri, result, params.Range.Start)...)
	return actions
}

func wantsQuickFix(only []protocol.CodeActionKind) bool {
	return len(only) == 0 || slices.Contains(only, protocol.CodeActionKindQuickFix)
}

func quickFix(title string, edit protocol.WorkspaceEdit) protocol.CodeAction {
	kind := protocol.CodeActionKindQuickFix
	return protocol.CodeAction{Title: title, Kind: &kind, Edit: &edit}
}

func missingStakeholdersEdit(uri protocol.DocumentUri, result coreanalysis.Result) protocol.WorkspaceEdit {
	m := result.Index.Model
	declared, missing := map[string]bool{}, map[string]bool{}
	for _, stakeholder := range m.Stakeholders {
		if stakeholder.Name != "" && !stakeholder.Invalid {
			declared[stakeholder.Name] = true
		}
	}
	add := func(name string) {
		if name != "" && grammar.Compiled.IsValidIdentifier(name) && !declared[name] {
			missing[name] = true
		}
	}
	for _, requirement := range m.Requirements {
		for _, clause := range requirement.Context.Clauses {
			add(clause.Actor)
		}
		add(requirement.Action.Target)
		for _, stakeholder := range requirement.Stakeholders {
			add(stakeholder.Name)
		}
	}
	names := sortedKeys(missing)
	if len(names) == 0 {
		return protocol.WorkspaceEdit{}
	}

	line := 1
	targetURI := uri
	if stakeholders := m.Stakeholders; len(stakeholders) > 0 {
		line = stakeholders[len(stakeholders)-1].Line
		targetURI = sourceLineURI(uri, result, line)
	}

	lines := make([]string, len(names))
	for i, name := range names {
		lines[i] = "stakeholder " + name
	}
	text := strings.Join(lines, "\n")
	if len(m.Stakeholders) == 0 {
		text += "\n\n"
		return editAt(targetURI, protocol.Range{}, text)
	}
	return editAt(targetURI, endOfLineRange(result, line), "\n"+text)
}

func duplicateValueActions(uri protocol.DocumentUri, result coreanalysis.Result) []protocol.CodeAction {
	var actions []protocol.CodeAction
	for _, value := range result.Index.Model.Values {
		if value.Builtin || value.Name == "" {
			continue
		}
		if _, ok := values.LookupByName(value.Name); ok {
			if edit, ok := removeLineEdit(uri, result, value.Line); ok {
				actions = append(actions, quickFix(fmt.Sprintf("Remove duplicate builtin value %s", value.Name), edit))
			}
			continue
		}
		if !value.HasAngle || !value.HasRadius {
			continue
		}
		matches := values.LookupByGeometry(value.Angle, value.Radius)
		if len(matches) == 0 {
			continue
		}
		edit, ok := removeLineEdit(uri, result, value.Line)
		if !ok {
			continue
		}
		for _, ref := range result.Index.ReferencesFor(grammar.DeclarationKindValue, value.Name) {
			location := locationForRange(uri, result, ref.Range)
			edit.Changes[location.URI] = append(edit.Changes[location.URI], protocol.TextEdit{
				Range:   location.Range,
				NewText: matches[0].Name,
			})
		}
		actions = append(actions, quickFix(fmt.Sprintf("Replace duplicate value %s with builtin %s", value.Name, matches[0].Name), edit))
	}
	return actions
}

func assignmentActions(uri protocol.DocumentUri, result coreanalysis.Result, position protocol.Position) []protocol.CodeAction {
	edits := assignmentEdits(uri, result, nil)
	if len(edits.Changes) == 0 {
		return nil
	}

	actions := []protocol.CodeAction{quickFix("Add missing assignment blocks", edits)}
	if requirement, ok := requirementAt(uri, result, position); ok {
		if edit := assignmentEdits(uri, result, &requirement); len(edit.Changes) > 0 {
			actions = append([]protocol.CodeAction{quickFix("Add assignment block for current requirement", edit)}, actions...)
		}
	}
	return actions
}

func assignmentEdits(uri protocol.DocumentUri, result coreanalysis.Result, current *model.Requirement) protocol.WorkspaceEdit {
	byRequirement := map[string][]model.AssignmentEntry{}
	for _, entry := range result.Index.Model.AssignmentEntries {
		byRequirement[entry.RequirementID] = append(byRequirement[entry.RequirementID], entry)
	}

	edit := protocol.WorkspaceEdit{Changes: map[protocol.DocumentUri][]protocol.TextEdit{}}
	for _, requirement := range result.Index.Model.Requirements {
		if current != nil && requirement.ID != current.ID {
			continue
		}
		entries := byRequirement[requirement.ID]
		missing := missingAssignmentStakeholders(requirement, entries)
		if len(missing) == 0 {
			continue
		}
		line := requirementEndLine(requirement)
		text := "\n\nassignment " + requirement.ID + "\n" + assignmentLines(missing)
		if len(entries) > 0 {
			line = assignmentEndLine(entries)
			text = "\n" + assignmentLines(missing)
		}
		targetURI := sourceLineURI(uri, result, line)
		edit.Changes[targetURI] = append(edit.Changes[targetURI], protocol.TextEdit{
			Range:   endOfLineRange(result, line),
			NewText: text,
		})
	}
	if len(edit.Changes) == 0 {
		return protocol.WorkspaceEdit{}
	}
	return edit
}

func missingAssignmentStakeholders(requirement model.Requirement, entries []model.AssignmentEntry) []string {
	if requirement.ID == "" {
		return nil
	}
	assigned, missing := map[string]bool{}, map[string]bool{}
	for _, entry := range entries {
		for _, stakeholder := range entry.Stakeholders {
			assigned[stakeholder.Name] = true
		}
	}
	for _, stakeholder := range requirement.Stakeholders {
		if stakeholder.Name != "" && !assigned[stakeholder.Name] {
			missing[stakeholder.Name] = true
		}
	}
	return sortedKeys(missing)
}

func assignmentLines(stakeholders []string) string {
	lines := make([]string, len(stakeholders))
	for i, stakeholder := range stakeholders {
		lines[i] = stakeholder + " -> "
	}
	return strings.Join(lines, "\n")
}

func requirementAt(uri protocol.DocumentUri, result coreanalysis.Result, position protocol.Position) (model.Requirement, bool) {
	line := int(position.Line) + 1
	if file, ok := sourceFileForURI(uri, result.Index.SourceFiles); ok {
		line += file.StartLine - 1
	}
	for _, requirement := range result.Index.Model.Requirements {
		if line >= requirement.Line && line <= requirementEndLine(requirement) {
			return requirement, true
		}
	}
	return model.Requirement{}, false
}

func requirementEndLine(requirement model.Requirement) int {
	line := requirement.Range.End.Line
	for _, clause := range requirement.Context.Clauses {
		line = max(line, clause.Range.End.Line)
	}
	line = max(line, requirement.Action.Range.End.Line, requirement.StakeholdersRange.End.Line)
	line = max(line, requirement.Metadata.Priority.Range.End.Line, requirement.Metadata.Retention.Range.End.Line, requirement.Metadata.Access.Range.End.Line)
	for _, link := range requirement.Traceability {
		line = max(line, link.Range.End.Line)
	}
	return max(line, requirement.Line)
}

func assignmentEndLine(entries []model.AssignmentEntry) int {
	line := 0
	for _, entry := range entries {
		line = max(line, entry.HeaderLine, entry.Line)
	}
	return line
}

func removeLineEdit(uri protocol.DocumentUri, result coreanalysis.Result, line int) (protocol.WorkspaceEdit, bool) {
	rng, ok := wholeLineRange(result, line)
	if !ok {
		return protocol.WorkspaceEdit{}, false
	}
	return editAt(sourceLineURI(uri, result, line), rng, ""), true
}

func editAt(uri protocol.DocumentUri, rng protocol.Range, text string) protocol.WorkspaceEdit {
	return protocol.WorkspaceEdit{Changes: map[protocol.DocumentUri][]protocol.TextEdit{
		uri: {{Range: rng, NewText: text}},
	}}
}

func sourceLineURI(defaultURI protocol.DocumentUri, result coreanalysis.Result, line int) protocol.DocumentUri {
	if file, ok := result.Index.SourceFileForLine(line); ok {
		return fileURI(file.Path)
	}
	return defaultURI
}

func wholeLineRange(result coreanalysis.Result, line int) (protocol.Range, bool) {
	if line < 1 || line > len(result.Index.Lines) {
		return protocol.Range{}, false
	}
	endLine, endColumn := line, len(result.Index.Lines[line-1].Raw)+1
	if file, ok := result.Index.SourceFileForLine(line); line < len(result.Index.Lines) && (!ok || line < file.EndLine) {
		endLine, endColumn = line+1, 1
	}
	return locationForRange("", result, sourcepos.NewRange(sourcepos.NewPosition(line, 1), sourcepos.NewPosition(endLine, endColumn))).Range, true
}

func endOfLineRange(result coreanalysis.Result, line int) protocol.Range {
	if line < 1 || line > len(result.Index.Lines) {
		return protocol.Range{}
	}
	column := len(result.Index.Lines[line-1].Raw) + 1
	pos := sourcepos.NewPosition(line, column)
	return locationForRange("", result, sourcepos.NewRange(pos, pos)).Range
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Package ioformat converts analyzed DSL documents to and from interchange formats.
package ioformat

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/model"
	"github.com/brogergvhs/value-dsl/internal/semantic"
)

type Document struct {
	Stakeholders []string      `json:"stakeholders,omitempty"`
	Values       []Value       `json:"values,omitempty"`
	Requirements []Requirement `json:"requirements,omitempty"`
	Assignments  []Assignment  `json:"assignments,omitempty"`
}

type Value struct {
	Name   string   `json:"name"`
	Angle  *float64 `json:"angle,omitempty"`
	Radius *float64 `json:"radius,omitempty"`
}

type Requirement struct {
	ID   string   `json:"id"`
	EARS []string `json:"ears"`
}

type Assignment struct {
	Requirement string            `json:"requirement"`
	Entries     []AssignmentEntry `json:"entries"`
}

type AssignmentEntry struct {
	Stakeholders []string `json:"stakeholders"`
	Values       []string `json:"values"`
}

type Codec interface {
	Decode(io.Reader) (Document, error)
	Encode(io.Writer, Document) error
}

type JSONCodec struct{}

var codecs = map[string]Codec{"json": JSONCodec{}}

func Decode(format string, r io.Reader) (Document, error) {
	codec, ok := codecs[strings.ToLower(strings.TrimSpace(format))]
	if !ok {
		return Document{}, fmt.Errorf("unsupported format %q", format)
	}
	return codec.Decode(r)
}

func Encode(format string, w io.Writer, doc Document) error {
	codec, ok := codecs[strings.ToLower(strings.TrimSpace(format))]
	if !ok {
		return fmt.Errorf("unsupported format %q", format)
	}
	return codec.Encode(w, doc)
}

func (JSONCodec) Decode(r io.Reader) (Document, error) {
	var doc Document
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return doc, dec.Decode(&doc)
}

func (JSONCodec) Encode(w io.Writer, doc Document) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func FromModel(m *semantic.SemanticModel) Document {
	if m == nil {
		return Document{}
	}
	doc := Document{
		Stakeholders: make([]string, 0, len(m.Stakeholders)),
		Values:       make([]Value, 0, len(m.Values)),
		Requirements: make([]Requirement, 0, len(m.Requirements)),
	}

	for _, s := range m.Stakeholders {
		if s.Name != "" {
			doc.Stakeholders = append(doc.Stakeholders, s.Name)
		}
	}

	for _, v := range m.Values {
		value := Value{Name: v.Name}
		if v.HasAngle {
			value.Angle = &v.Angle
		}
		if v.HasRadius {
			value.Radius = &v.Radius
		}
		if value.Name != "" {
			doc.Values = append(doc.Values, value)
		}
	}

	for _, r := range m.Requirements {
		if r.ID != "" {
			doc.Requirements = append(doc.Requirements, Requirement{ID: r.ID, EARS: earsLines(r)})
		}
	}

	doc.Assignments = assignments(m)
	return doc
}

func ToDSL(doc Document) string {
	var b strings.Builder
	for _, s := range doc.Stakeholders {
		fmt.Fprintf(&b, "stakeholder %s\n", s)
	}
	if len(doc.Stakeholders) > 0 {
		b.WriteByte('\n')
	}

	for _, v := range doc.Values {
		if v.Angle != nil && v.Radius != nil {
			fmt.Fprintf(&b, "value %s = %g, %g\n", v.Name, *v.Angle, *v.Radius)
		} else {
			fmt.Fprintf(&b, "value %s\n", v.Name)
		}
	}
	if len(doc.Values) > 0 {
		b.WriteByte('\n')
	}

	for _, r := range doc.Requirements {
		fmt.Fprintf(&b, "requirement %s\n", r.ID)
		for _, line := range r.EARS {
			fmt.Fprintln(&b, line)
		}
		b.WriteByte('\n')
	}

	for _, a := range groupAssignments(doc.Assignments) {
		fmt.Fprintf(&b, "assignment %s\n", a.Requirement)
		for _, e := range a.Entries {
			fmt.Fprintf(&b, "%s -> %s\n", strings.Join(e.Stakeholders, ", "), strings.Join(e.Values, ", "))
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func earsLines(r model.Requirement) []string {
	lines := make([]string, 0, len(r.Context.Clauses)+4+len(r.Traceability))
	for _, c := range r.Context.Clauses {
		lines = append(lines, string(c.Type)+" "+c.RawText)
	}
	if r.Action.RawText != "" {
		lines = append(lines, "system shall "+r.Action.RawText)
	}
	if len(r.Stakeholders) > 0 {
		names := make([]string, 0, len(r.Stakeholders))
		for _, s := range r.Stakeholders {
			names = append(names, s.Name)
		}
		lines = append(lines, "stakeholders "+strings.Join(names, ", "))
	}
	if r.Metadata.Priority.Value != "" {
		lines = append(lines, "priority "+r.Metadata.Priority.Value)
	}
	if r.Metadata.Retention.Value != "" {
		lines = append(lines, "retention "+quote(r.Metadata.Retention.Value))
	}
	if r.Metadata.Access.Value != "" {
		lines = append(lines, "access "+quote(r.Metadata.Access.Value))
	}
	for _, t := range r.Traceability {
		lines = append(lines, "linked_to "+quote(t.Value))
	}

	return lines
}

func assignments(m *semantic.SemanticModel) []Assignment {
	out := make([]Assignment, 0)
	index := map[string]int{}

	for _, entry := range m.AssignmentEntries {
		i, ok := index[entry.RequirementID]
		if !ok {
			i = len(out)
			index[entry.RequirementID] = i
			out = append(out, Assignment{Requirement: entry.RequirementID})
		}
		out[i].Entries = append(out[i].Entries, AssignmentEntry{Stakeholders: refs(entry.Stakeholders), Values: refs(entry.Values)})
	}

	return out
}

func groupAssignments(in []Assignment) []Assignment {
	out := make([]Assignment, 0, len(in))
	index := map[string]int{}

	for _, assignment := range in {
		i, ok := index[assignment.Requirement]
		if !ok {
			i = len(out)
			index[assignment.Requirement] = i
			out = append(out, Assignment{Requirement: assignment.Requirement})
		}
		out[i].Entries = append(out[i].Entries, assignment.Entries...)
	}

	return out
}

func refs(in []model.Ref) []string {
	out := make([]string, 0, len(in))
	for _, r := range in {
		out = append(out, r.Name)
	}

	return out
}

func quote(s string) string {
	tokens := grammar.TokenizeRawLine(1, s).Tokens
	if len(tokens) != 1 || tokens[0].Text != s {
		return "'" + strings.ReplaceAll(s, "'", "\\'") + "'"
	}
	return s
}

// Package model defines the core domain types shared across parsing,
// semantic processing, validation, conflict analysis, and import/export layers.
package model

import (
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
)

type Requirement struct {
	Line              int
	Range             sourcepos.Range
	StakeholdersLine  int
	StakeholdersRange sourcepos.Range
	// Invalid marks a declared requirement that has blocking parse or validation errors.
	// It is used to suppress downstream analysis that should only consider valid declarations.
	Invalid      bool
	ID           string
	Context      RequirementContext
	Action       Action
	Stakeholders []Stakeholder
	Metadata     RequirementMetadata
	Traceability []TraceabilityLink
}

type RequirementContext struct {
	Clauses []Clause
}

type Clause struct {
	Line    int
	Range   sourcepos.Range
	Type    grammar.RequirementClauseKind
	Actor   string
	Verb    string
	Target  string
	RawText string
}

type RequirementMetadata struct {
	Priority  MetadataField
	Retention MetadataField
	Access    MetadataField
}

type MetadataField struct {
	Line  int
	Range sourcepos.Range
	Value string
}

type TraceabilityLink struct {
	Line  int
	Range sourcepos.Range
	Value string
}

type Stakeholder struct {
	Line  int
	Range sourcepos.Range
	// Invalid marks a referenced-but-undeclared stakeholder or a declaration with blocking
	// parse/validation errors. It does not mean "unused"; unused is reported as a warning.
	Invalid bool
	Name    string
}

type Value struct {
	Line  int
	Range sourcepos.Range
	// Invalid marks a value declaration with blocking parse/validation errors.
	// Built-in geometry matches are warnings and do not set this flag.
	Invalid       bool
	Name          string
	Category      string
	Angle         float64
	Radius        float64
	Builtin       bool
	HasAngle      bool
	HasRadius     bool
	HasValueComma bool
}

type Action struct {
	Line      int
	Range     sourcepos.Range
	Verb      string
	Object    string
	Target    string
	Mechanism string
	RawText   string
}

type Ref struct {
	Name string
}

type AssignmentEntry struct {
	Line          int
	HeaderLine    int
	RequirementID string
	Stakeholders  []Ref
	Values        []Ref
}

// ConflictResult represents a computed conflict between two stakeholder value assignments.
type ConflictResult struct {
	RequirementID string
	StakeholderA  string
	ValueA        string
	StakeholderB  string
	ValueB        string
	ConflictScore float64
}

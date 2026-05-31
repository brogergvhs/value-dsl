// Package grammar is the declarative source of truth for DSL syntax,
// diagnostics, formatting, LSP metadata, and Tree-sitter export.
package grammar

var whenVerbs = []string{
	"enters", "leaves", "starts", "stops", "requests",
	"crosses", "exceeds", "shows", "reports",
}

type actionFamily struct {
	Kind            ActionKind
	Description     string
	Verbs           []string
	DirectTarget    bool // true: <verb> <target>; false: <verb> <object> of <target>
	AllowsMechanism bool
}

var actionFamilies = []actionFamily{
	{
		Kind:            ActionKindMonitoring,
		Description:     "Observes or captures information about a target.",
		DirectTarget:    false,
		AllowsMechanism: true,
		Verbs: []string{
			"track", "monitor", "observe", "detect",
			"measure", "record", "collect", "capture",
		},
	},
	{
		Kind:            ActionKindNotification,
		Description:     "Informs or warns a target directly.",
		DirectTarget:    true,
		AllowsMechanism: true,
		Verbs:           []string{"notify", "alert", "inform", "warn"},
	},
	{
		Kind:            ActionKindPersistence,
		Description:     "Persists information about a target.",
		DirectTarget:    false,
		AllowsMechanism: true,
		Verbs:           []string{"log", "store", "save", "archive", "retain"},
	},
	{
		Kind:            ActionKindRestriction,
		Description:     "Constrains, removes, or obscures information about a target.",
		DirectTarget:    false,
		AllowsMechanism: false,
		Verbs:           []string{"restrict", "limit", "anonymize", "mask", "delete", "remove"},
	},
}

var Grammar = Spec{
	Identifier:    `[a-zA-Z_][a-zA-Z0-9_-]*`,
	NumberPattern: `-?(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)`,
	StringPattern: `(?:"[^"]*"|'[^']*')`,
	LineComment:   "//",
	Declarations: []Decl{
		{
			Kind:       DeclarationKindStakeholder,
			Keyword:    "stakeholder",
			DenseGroup: true,
			Header: []Matcher{
				Kw{"stakeholder"},
				List{
					Name:      "names",
					Separator: ",",
					Item:      Ident{Name: "name"},
				},
			},
		},
		{
			Kind:       DeclarationKindValue,
			Keyword:    "value",
			DenseGroup: true,
			Header: []Matcher{
				Kw{"value"}, Ident{Name: "name"},
				Opt{Inner: []Matcher{
					Kw{"="},
					Tok{Kind: TokKindNumber, Name: "angle"},
					Kw{","},
					Tok{Kind: TokKindNumber, Name: "radius"},
				}},
			},
		},
		{
			Kind:    DeclarationKindRequirement,
			Keyword: "requirement",
			Header:  []Matcher{Kw{"requirement"}, Ident{Name: "id"}},
			Body:    &requirementBody,
		},
		{
			Kind:    DeclarationKindAssignment,
			Keyword: "assignment",
			Header: []Matcher{
				Kw{"assignment"},
				Ref{Name: "requirement", Kind: DeclarationKindRequirement},
			},
			Body: &assignmentBody,
		},
	},
}

var requirementBody = Block{
	Lines: func() []LineRule {
		lines := []LineRule{
			{
				Name:       "While",
				ClauseKind: RequirementClauseWhile,
				Order:      1,
				Repeatable: true,
				Pattern: []Matcher{
					Kw{"while"},
					Ref{Name: "actor", Kind: DeclarationKindStakeholder},
					Free{Name: "rest"},
				},
			},
			{
				Name:       "When",
				ClauseKind: RequirementClauseWhen,
				Order:      2,
				Pattern: []Matcher{
					Kw{"when"},
					Ref{Name: "actor", Kind: DeclarationKindStakeholder},
					Tok{Kind: TokKindWhenVerb, Name: "verb", Set: whenVerbs},
					Ident{Name: "target"},
				},
			},
			{
				Name:       "If",
				ClauseKind: RequirementClauseIf,
				Order:      3,
				Pattern: []Matcher{
					Kw{"if"},
					Ref{Name: "actor", Kind: DeclarationKindStakeholder},
					Free{Name: "rest"},
				},
			},
			{
				Name:       "Where",
				ClauseKind: RequirementClauseWhere,
				Order:      4,
				Pattern:    []Matcher{Kw{"where"}, Free{Name: "rest"}},
			},
		}

		for _, family := range actionFamilies {
			lines = append(lines, LineRule{
				Name:       "Shall_" + string(family.Kind),
				ClauseKind: RequirementClauseSystemShall,
				Order:      5,
				Required:   true,
				Pattern:    shallPattern(family),
			})
		}

		lines = append(lines,
			LineRule{
				Name:       "Stakeholders",
				ClauseKind: RequirementClauseStakeholders,
				Order:      6,
				Required:   true,
				Pattern: []Matcher{
					Kw{"stakeholders"},
					List{
						Name:      "actors",
						Separator: ",",
						Item:      Ref{Kind: DeclarationKindStakeholder},
					},
				},
			},
			LineRule{
				Name:       "Priority",
				ClauseKind: RequirementClausePriority,
				Order:      7,
				Pattern: []Matcher{
					Kw{"priority"},
					Enum{Name: "level", Values: []string{"low", "medium", "high", "critical"}},
				},
			},
			LineRule{
				Name:       "Retention",
				ClauseKind: RequirementClauseRetention,
				Order:      8,
				Pattern:    []Matcher{Kw{"retention"}, Str{Name: "value"}},
			},
			LineRule{
				Name:       "Access",
				ClauseKind: RequirementClauseAccess,
				Order:      9,
				Pattern:    []Matcher{Kw{"access"}, Str{Name: "value"}},
			},
			LineRule{
				Name:       "LinkedTo",
				ClauseKind: RequirementClauseLinkedTo,
				Order:      10,
				Repeatable: true,
				Pattern:    []Matcher{Kw{"linked_to"}, Str{Name: "target"}},
			},
		)

		return lines
	}(),
}

func shallPattern(family actionFamily) []Matcher {
	verb := Tok{Kind: TokKindActionVerb, Name: "verb", Set: family.Verbs}
	target := Ref{Name: "target", Kind: DeclarationKindStakeholder}

	pattern := []Matcher{Kw{"system"}, Kw{"shall"}, verb}
	if family.DirectTarget {
		pattern = append(pattern, target)
	} else {
		pattern = append(pattern, Str{Name: "object"}, Kw{"of"}, target)
	}
	if family.AllowsMechanism {
		pattern = append(pattern, Opt{Inner: []Matcher{Kw{"using"}, Str{Name: "mechanism"}}})
	}
	return pattern
}

var assignmentBody = Block{
	Lines: []LineRule{
		{
			Name:       "AssignmentEntry",
			Order:      1,
			Required:   true,
			Repeatable: true,
			Expansion:  ExpansionCartesian,
			Semantics: []SemanticBinding{
				{Role: SemanticRoleAssignmentSource, Field: "stakeholders"},
				{Role: SemanticRoleAssignmentTarget, Field: "values"},
			},
			Pattern: []Matcher{
				List{
					Name:      "stakeholders",
					Separator: ",",
					Item:      Ref{Kind: DeclarationKindStakeholder},
				},
				Kw{"->"},
				List{
					Name:      "values",
					Separator: ",",
					Item:      Ref{Kind: DeclarationKindValue},
				},
			},
		},
	},
}

package completion

import (
	"slices"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/grammar"
)

// completionPlan describes what to suggest at the cursor position inside a clause.
// Invariants:
//   - inside=false means all other fields are ignored.
//   - at most one expect* field may be true.
//   - keywords may be combined with one expect* field when both keyword and symbol
//     completions are valid at the same cursor position.
type completionPlan struct {
	inside            bool
	keywords          []string
	expectStakeholder bool // stakeholder ref (actor, target, list member)
	expectVerb        bool // when-verb
	expectIdentifier  bool // free identifier (when target, metadata value)
	expectRequirement bool // requirement ref (assignment header)
	expectValue       bool // value ref (assignment entry)
}

func (p completionPlan) valid() bool {
	count := 0
	for _, active := range []bool{
		p.expectStakeholder,
		p.expectVerb,
		p.expectIdentifier,
		p.expectRequirement,
		p.expectValue,
	} {
		if active {
			count++
		}
	}
	return count <= 1
}

func completeActionPrefix(prefix string) completionPlan {
	trimmed := strings.TrimSpace(prefix)
	if !strings.HasPrefix(trimmed, "system shall") {
		return completionPlan{}
	}

	actionText := strings.TrimSpace(strings.TrimPrefix(trimmed, "system shall"))
	fields := strings.Fields(actionText)
	if len(fields) == 0 {
		return completionPlan{inside: true, keywords: grammar.Compiled.ActionVerbs}
	}

	spec, ok := grammar.Compiled.ActionSpecByVerb[fields[0]]
	if !ok {
		return completionPlan{inside: true, keywords: grammar.Compiled.ActionVerbs}
	}

	return completeActionFields(spec, fields[1:], strings.HasSuffix(prefix, " "))
}

func completeActionFields(spec grammar.ActionSpec, operands []string, trailingSpace bool) completionPlan {
	plan := completionPlan{inside: true}

	switch spec.Form {
	case grammar.ActionFormDirectTarget:
		usingIndex := slices.Index(operands, "using")
		if usingIndex >= 0 {
			return plan
		}
		switch {
		case len(operands) == 0 && trailingSpace:
			plan.expectStakeholder = true
		case len(operands) == 1 && !trailingSpace:
			plan.expectStakeholder = true
		case len(operands) >= 1 && trailingSpace && spec.AllowsMechanism:
			plan.keywords = []string{"using"}
		}
		return plan

	case grammar.ActionFormObjectOfTarget:
		ofIndex := slices.Index(operands, "of")
		usingIndex := slices.Index(operands, "using")
		targetCount := 0
		if ofIndex >= 0 {
			targetCount = len(operands) - ofIndex - 1
		}
		switch {
		case ofIndex == -1:
			if trailingSpace && len(operands) >= 1 {
				plan.keywords = []string{"of"}
			}
		case ofIndex == len(operands)-1 && trailingSpace:
			plan.expectStakeholder = true
		case usingIndex == -1 && ofIndex == len(operands)-2 && !trailingSpace:
			plan.expectStakeholder = true
		case usingIndex == -1 && ofIndex == len(operands)-2 && trailingSpace:
			if spec.AllowsMechanism {
				plan.keywords = []string{"using"}
			}
		case usingIndex == -1 && trailingSpace && targetCount >= 1 && spec.AllowsMechanism:
			plan.keywords = []string{"using"}
		case usingIndex == -1 && ofIndex >= 0 && targetCount <= 1:
			plan.expectStakeholder = true
		}
		return plan
	}

	return plan
}

func completeAssignmentPrefix(prefix string, isHeader bool) completionPlan {
	if isHeader {
		return completionPlan{inside: true, expectRequirement: true}
	}
	if strings.Contains(prefix, "->") {
		return completionPlan{inside: true, expectValue: true}
	}

	return completionPlan{inside: true, expectStakeholder: true}
}

func completeWhenPrefix(prefix string) completionPlan {
	trimmed := strings.TrimSpace(prefix)
	if firstCompletionToken(trimmed) != "when" {
		return completionPlan{}
	}

	fields := strings.Fields(strings.TrimSpace(strings.TrimPrefix(trimmed, "when")))
	trailing := strings.HasSuffix(prefix, " ")
	switch {
	case len(fields) == 0:
		return completionPlan{inside: true, expectStakeholder: true}
	case len(fields) == 1 && trailing:
		return completionPlan{inside: true, expectVerb: true}
	case len(fields) == 1:
		return completionPlan{inside: true, expectStakeholder: true}
	case len(fields) == 2 && !trailing:
		return completionPlan{inside: true, expectVerb: true}
	case len(fields) == 2 && trailing:
		return completionPlan{inside: true, expectIdentifier: true}
	default:
		return completionPlan{inside: true}
	}
}

func completeActorClausePrefix(prefix, keyword string) completionPlan {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" || firstCompletionToken(trimmed) != keyword {
		return completionPlan{}
	}

	rest := strings.Fields(strings.TrimSpace(strings.TrimPrefix(trimmed, keyword)))
	actor := len(rest) == 0 || (len(rest) == 1 && !strings.HasSuffix(prefix, " "))

	return completionPlan{inside: true, expectStakeholder: actor}
}

func completeStakeholdersPrefix(prefix string) completionPlan {
	trimmed := strings.TrimSpace(prefix)
	if firstCompletionToken(trimmed) != "stakeholders" {
		return completionPlan{}
	}

	fields := strings.Fields(trimmed)
	var kws []string
	if len(fields) <= 1 && !strings.Contains(prefix, ",") {
		kws = []string{"stakeholders"}
	}

	return completionPlan{inside: true, keywords: kws, expectStakeholder: true}
}

func completeMetadataPrefix(prefix string) completionPlan {
	tok := firstCompletionToken(strings.TrimSpace(prefix))
	for _, kw := range grammar.Compiled.MetadataKeywords {
		if kw != "stakeholders" && kw == tok {
			return completionPlan{inside: true, expectIdentifier: true}
		}
	}
	return completionPlan{}
}

func isTopLevelBoundaryLine(line string) bool {
	token := firstCompletionToken(strings.TrimSpace(line))

	if token == "" {
		return strings.Contains(line, "->")
	}
	if decl, ok := grammar.Compiled.DeclarationByKeyword[token]; ok && decl.Body == nil {
		return true
	}
	if slices.Contains(grammar.Compiled.MetadataKeywords, token) {
		return true
	}

	return strings.Contains(line, "->")
}

func isTopLevelKeywordPrefix(prefix string) bool {
	return slices.ContainsFunc(grammar.Compiled.TopLevelKeywords, func(kw string) bool { return strings.HasPrefix(kw, prefix) })
}

func firstCompletionToken(v string) string {
	if f := strings.Fields(v); len(f) > 0 {
		return f[0]
	}
	return ""
}

package semantictokens

import (
	"strings"
	"testing"
)

func TestGraphMatchesFollowSharedDeclarationReferenceGraph(t *testing.T) {
	text := strings.Join([]string{
		"stakeholder Worker",
		"value privacy_pref = 1.58, 0.91",
		"requirement R1",
		"while Worker is in DangerousArea",
		"system shall track location of Worker using Camera",
		"stakeholders Worker",
		"assignment R1",
		"Worker -> privacy_pref",
		"assignment R2",
		"Worker -> privacy",
	}, "\n") + "\n"

	result := mustAnalyzeText(t, text)
	matches := indexMatches(result.Index, result.Index.Lines)

	assertSemanticRole(t, text, matches, "Worker", 1, RoleStakeholderDeclaration)
	assertSemanticRole(t, text, matches, "privacy_pref", 2, RoleValueDeclaration)
	assertSemanticRole(t, text, matches, "R1", 3, RoleRequirementDeclaration)
	assertSemanticRole(t, text, matches, "Worker", 4, RoleStakeholderReference)
	assertSemanticRole(t, text, matches, "Worker", 5, RoleStakeholderReference)
	assertSemanticRole(t, text, matches, "Worker", 6, RoleStakeholderReference)
	assertSemanticRole(t, text, matches, "R1", 7, RoleRequirementReference)
	assertSemanticRole(t, text, matches, "privacy_pref", 8, RoleValueReference)
	assertSemanticRole(t, text, matches, "privacy", 10, RoleBuiltinValue)
	assertNoSemanticRole(t, text, matches, "Camera", 5)
	assertNoSemanticRole(t, text, matches, "DangerousArea", 4)
	assertNoSemanticRole(t, text, matches, "track", 5)
	assertNoSemanticRole(t, text, matches, "location", 5)
}

func TestGraphMatchesRemainAvailableInPartiallyBrokenFile(t *testing.T) {
	text := strings.Join([]string{
		"stakeholder Worker",
		"value privacy_pref = 10.8,",
		"unknown top level",
		"",
		"requirement",
		"system shall",
		"stakeholders",
		"",
		"requirement R2",
		"when Worker enters DangerZone",
		"system shall notify Worker using Alarm",
		"stakeholders Worker",
		"",
		"assignment R2",
		"Worker -> privacy_pref",
	}, "\n") + "\n"

	result := mustAnalyzeText(t, text)
	matches := indexMatches(result.Index, result.Index.Lines)

	assertSemanticRole(t, text, matches, "Worker", 1, RoleStakeholderDeclaration)
	assertSemanticRole(t, text, matches, "privacy_pref", 2, RoleValueDeclaration)
	assertSemanticRole(t, text, matches, "R2", 9, RoleRequirementDeclaration)
	assertSemanticRole(t, text, matches, "Worker", 10, RoleStakeholderReference)
	assertSemanticRole(t, text, matches, "R2", 14, RoleRequirementReference)
	assertSemanticRole(t, text, matches, "privacy_pref", 15, RoleValueReference)
	assertNoSemanticRole(t, text, matches, "unknown", 3)
}

func TestResolveReturnsEmptyDataArrayWhenNoTokens(t *testing.T) {
	tokens := Resolve("", mustAnalyzeText(t, ""))
	if tokens == nil || tokens.Data == nil || len(tokens.Data) != 0 {
		t.Fatalf("expected empty semantic token data, got %+v", tokens)
	}
}

func assertSemanticRole(t *testing.T, text string, matches []Match, token string, line int, want Role) {
	t.Helper()

	for _, match := range matches {
		if match.Line != line {
			continue
		}
		if matchedText(text, match) != token {
			continue
		}
		if match.Role != want {
			t.Fatalf("token %q on line %d role = %q, want %q", token, line, match.Role, want)
		}
		return
	}

	t.Fatalf("missing token %q on line %d", token, line)
}

func assertNoSemanticRole(t *testing.T, text string, matches []Match, token string, line int) {
	t.Helper()

	for _, match := range matches {
		if match.Line == line && matchedText(text, match) == token {
			t.Fatalf("unexpected semantic token %q on line %d with role %q", token, line, match.Role)
		}
	}
}

func matchedText(text string, match Match) string {
	lines := strings.Split(text, "\n")
	line := lines[match.Line-1]
	return line[match.StartColumn-1 : match.EndColumn-1]
}

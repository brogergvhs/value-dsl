package grammar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/values"
)

func TestReservedKeywordsIncludeCoreLanguageTokens(t *testing.T) {
	tests := []string{
		"stakeholder",
		"value",
		"requirement",
		"assignment",
		"when",
		"while",
		"if",
		"where",
		"system",
		"shall",
		"priority",
		"linked_to",
		"track",
		"enters",
	}

	for _, token := range tests {
		if !slices.Contains(Compiled.ReservedKeywords, token) {
			t.Fatalf("expected %q to be reserved", token)
		}
	}
}

func TestTokenInfoCarriesDocumentation(t *testing.T) {
	info, ok := TokenInfoFor("stakeholder")
	if !ok {
		t.Fatal("expected token info for stakeholder")
	}
	if info.Documentation == "" {
		t.Fatalf("expected stakeholder token documentation, got %+v", info)
	}
}

func TestClauseOrderMatchesEARSCanonicalOrder(t *testing.T) {
	rules := Compiled.requirementClauseRules
	got := make([]RequirementClauseKind, 0, 4)
	for _, rule := range rules {
		switch rule.Kind {
		case RequirementClauseWhile, RequirementClauseWhen, RequirementClauseIf, RequirementClauseWhere:
			got = append(got, rule.Kind)
		}
	}
	want := []RequirementClauseKind{
		RequirementClauseWhile,
		RequirementClauseWhen,
		RequirementClauseIf,
		RequirementClauseWhere,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d clause types, got %d", len(want), len(got))
	}

	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("unexpected clause order at %d: got %q want %q", idx, got[idx], want[idx])
		}
	}
}

func TestValueCategoriesExposeDescriptions(t *testing.T) {
	categories := values.Categories()
	if len(categories) == 0 {
		t.Fatal("expected value categories")
	}
	if categories[0].Name != "power" {
		t.Fatalf("unexpected first category: %+v", categories[0])
	}
	if categories[4].Name != "self_direction" || categories[4].Description == "" {
		t.Fatalf("expected self_direction category description, got %+v", categories[4])
	}
}

func TestCommentHelpersStripAndClampVisibleText(t *testing.T) {
	line := "system shall notify Worker // comment Worker"
	if got := TokenizeRawLine(1, line).Code; got != "system shall notify Worker " {
		t.Fatalf("unexpected stripped line: %q", got)
	}
	if got := TokenizeRawLine(1, `retention "https://example.test/a//b" // link`).Code; got != `retention "https://example.test/a//b" ` {
		t.Fatalf("unexpected double-quoted stripped line: %q", got)
	}
	if got := TokenizeRawLine(1, `linked_to 'Safety // Case' // trace`).Code; got != `linked_to 'Safety // Case' ` {
		t.Fatalf("unexpected single-quoted stripped line: %q", got)
	}

	tokenized := TokenizeRawLine(1, line)
	character, inComment := tokenized.VisibleOffset(len(line))
	visible := tokenized.Code
	if !inComment {
		t.Fatal("expected cursor at end of line to be inside comment")
	}
	if visible != "system shall notify Worker " {
		t.Fatalf("unexpected visible line: %q", visible)
	}
	if character != len(visible) {
		t.Fatalf("unexpected clamped character: %d", character)
	}
}

func TestRequirementClauseKindsAreDetectedFromLines(t *testing.T) {
	tests := []struct {
		line string
		kind RequirementClauseKind
	}{
		{"while Worker is in DangerousArea", RequirementClauseWhile},
		{"when Worker enters DangerousArea", RequirementClauseWhen},
		{"if Worker has_no_consent", RequirementClauseIf},
		{"where SafetyMode enabled", RequirementClauseWhere},
		{"system shall notify Worker", RequirementClauseSystemShall},
		{"stakeholders Worker", RequirementClauseStakeholders},
		{"priority high", RequirementClausePriority},
		{"retention short_term", RequirementClauseRetention},
		{"access supervisor_only", RequirementClauseAccess},
		{"linked_to HazardAnalysis", RequirementClauseLinkedTo},
	}

	for _, tt := range tests {
		kind, ok := Compiled.RequirementClauseKindForLine(tt.line)
		if !ok {
			t.Fatalf("expected clause kind for %q", tt.line)
		}
		if kind != tt.kind {
			t.Fatalf("unexpected clause kind for %q: got %q want %q", tt.line, kind, tt.kind)
		}
	}
}

func TestNextRequirementClauseKeywordsFollowLanguageBlockOrder(t *testing.T) {
	tests := []struct {
		name string
		seen []RequirementClauseKind
		want []string
	}{
		{
			name: "start of requirement",
			want: []string{"while", "when", "if", "where", "system shall"},
		},
		{
			name: "after when",
			seen: []RequirementClauseKind{RequirementClauseWhen},
			want: []string{"if", "where", "system shall"},
		},
		{
			name: "after action",
			seen: []RequirementClauseKind{RequirementClauseSystemShall},
			want: []string{"stakeholders"},
		},
		{
			name: "after stakeholders",
			seen: []RequirementClauseKind{RequirementClauseSystemShall, RequirementClauseStakeholders},
			want: []string{"priority", "retention", "access", "linked_to"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compiled.NextRequirementClauseKeywords(tt.seen)
			if len(got) != len(tt.want) {
				t.Fatalf("unexpected keyword count: got %v want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("unexpected keyword order: got %v want %v", got, tt.want)
				}
			}
		})
	}
}

func TestExportProvidesStableMachineReadableSpec(t *testing.T) {
	exported := Export()

	if exported.Version != ExportVersion {
		t.Fatalf("unexpected export version: %+v", exported)
	}
	if exported.IdentifierPattern != Compiled.Spec.Identifier {
		t.Fatalf("unexpected identifier pattern: %+v", exported)
	}
	if exported.LineCommentPrefix != Compiled.Spec.LineComment {
		t.Fatalf("unexpected line comment prefix: %+v", exported)
	}
	if len(exported.Declarations) != len(Compiled.Spec.Declarations) {
		t.Fatalf("expected declaration export, got %+v", exported.Declarations)
	}
	if len(exported.Actions) != len(Compiled.ActionSpecs) {
		t.Fatalf("expected action export, got %+v", exported.Actions)
	}
	if !slices.Contains(exported.ReservedKeywords, "requirement") {
		t.Fatalf("expected reserved keyword export, got %+v", exported.ReservedKeywords)
	}
	if len(exported.Tokens) == 0 || exported.Tokens[0].Text == "" {
		t.Fatalf("expected exported token info, got %+v", exported.Tokens)
	}
	if len(exported.Declarations[2].Body.Lines) == 0 || exported.Declarations[2].Body.Lines[0].ClauseKind == "" {
		t.Fatalf("expected exported line rules, got %+v", exported.Declarations[2].Body)
	}

	data, err := ExportJSON()
	if err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}

	var decoded ExportedSpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("export JSON should unmarshal: %v", err)
	}
	if decoded.IdentifierPattern != Compiled.Spec.Identifier {
		t.Fatalf("unexpected decoded export: %+v", decoded)
	}
	if decoded.LineCommentPrefix != Compiled.Spec.LineComment {
		t.Fatalf("unexpected decoded comment prefix: %+v", decoded)
	}
}

func TestWriteJSONFileCreatesExportOnDisk(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "generated", "grammar.json")
	if err := WriteJSONFile(outputPath); err != nil {
		t.Fatalf("WriteJSONFile() error = %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected export file to be written: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty export file")
	}

	var decoded ExportedSpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("written export should unmarshal: %v", err)
	}
	if decoded.Version != ExportVersion {
		t.Fatalf("unexpected exported file contents: %+v", decoded)
	}
}

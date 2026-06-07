package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
)

func TestCLIParseCommandSmoke(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "valid.dsl")
	dsl := `stakeholder Worker

value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
`

	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"parse", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, `"Declarations"`) || !strings.Contains(got, `"Kind": "requirement"`) {
		t.Fatalf("expected AST JSON output, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func runTestAnalysis(t *testing.T, dsl string, strict bool) coreanalysis.Result {
	t.Helper()
	result, err := coreanalysis.Run(dsl, coreanalysis.Options{
		Strict:         strict,
		BuildConflicts: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	return result
}

func TestCLIFormatCommandCanonicalizesOutput(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "format.dsl")
	dsl := `stakeholder Worker
value privacy_pref = 1.58, 0.91
requirement R1
system shall notify Worker
stakeholders Worker
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"format", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	expected := `stakeholder Worker

value privacy_pref = 1.58, 0.91

requirement R1
system shall notify Worker
stakeholders Worker
`
	if stdout.String() != expected {
		t.Fatalf("expected formatted output:\n%s\ngot:\n%s", expected, stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLIExportCommandWritesJSON(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "valid.dsl")
	outPath := filepath.Join(tmpDir, "valid.json")
	dsl := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Manager

assignment R1
Worker -> privacy_pref
Manager -> privacy_pref
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"export", "--format", "json", "--output", outPath, dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(content)
	for _, want := range []string{`"stakeholders"`, `"Worker"`, `"ears"`, `"when Worker enters DangerousArea"`, `"assignment`} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected exported JSON to contain %q, got %s", want, got)
		}
	}
	if strings.Count(got, `"requirement": "R1"`) != 1 {
		t.Fatalf("expected exported JSON to group assignment entries under one R1 assignment, got %s", got)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIImportCommandWritesDSL(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "valid.json")
	outPath := filepath.Join(tmpDir, "valid.dsl")
	input := `{
  "stakeholders": ["Worker", "Manager"],
  "values": [{"name": "privacy_pref", "angle": 1.58, "radius": 0.91}],
  "requirements": [{
    "id": "R1",
    "ears": [
      "when Worker enters DangerousArea",
      "system shall track location of Worker using Camera",
      "stakeholders Worker, Manager",
      "linked_to \"google.com/search?q=lsp\""
    ]
  }],
  "assignments": [{
    "requirement": "R1",
    "entries": [{"stakeholders": ["Worker"], "values": ["privacy_pref"]}]
  }, {
    "requirement": "R1",
    "entries": [{"stakeholders": ["Manager"], "values": ["privacy_pref"]}]
  }]
}`
	if err := os.WriteFile(jsonPath, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"import", "--format", "json", "--output", outPath, jsonPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(content)
	for _, want := range []string{"stakeholder Worker", "value privacy_pref = 1.58, 0.91", "when Worker enters DangerousArea", "linked_to 'google.com/search?q=lsp'", "assignment R1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected imported DSL to contain %q, got %s", want, got)
		}
	}
	if strings.Count(got, "assignment R1\n") != 1 || !strings.Contains(got, "assignment R1\nWorker -> privacy_pref\nManager -> privacy_pref\n") {
		t.Fatalf("expected imported assignment entries to be grouped, got %s", got)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIParseCommandReportsDiagnosticsWithCodes(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "invalid.dsl")
	dsl := `stakeholder Worker

requirement R1
system shall inspect location of Worker using Camera
stakeholders Worker
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"parse", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected parse to succeed on semantically invalid input, got %v", err)
	}
	if !strings.Contains(stdout.String(), `"Kind": "system_shall"`) || !strings.Contains(stdout.String(), `"Text": "inspect"`) || !strings.Contains(stdout.String(), `"Declarations"`) {
		t.Fatalf("expected AST JSON output for semantically invalid input, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLIFormatCommandFormatsWithoutSemanticValidation(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "invalid-format.dsl")
	dsl := `stakeholder Worker

requirement R1
system shall track location of Worker using Camera
stakeholders
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"format", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected format to succeed without semantic validation, got %v", err)
	}
	if !strings.Contains(stdout.String(), "requirement R1\nsystem shall track location of Worker using Camera\nstakeholders\n") {
		t.Fatalf("unexpected format output: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLIFormatWriteRefusesMalformedInput(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "unsafe-format.dsl")
	dsl := `stakeholder Worker
requirement R1
unknown clause
system shall notify Worker
stakeholders Worker
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"format", "--write", dslPath})
	defer func() { _ = formatCmd.Flags().Set("write", "false") }()

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected format --write to reject malformed input")
	}
	if !strings.Contains(err.Error(), "format failed: format --write refused") {
		t.Fatalf("unexpected format --write error: %v", err)
	}
	content, err := os.ReadFile(dslPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != dsl {
		t.Fatalf("format --write modified unsafe input:\n%s", string(content))
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output on Execute error, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIValidateCommandPrintsStandardDiagnosticBlock(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "invalid-validate.dsl")
	dsl := `stakeholder Worker

requirement R1
system shall track location of Worker using Camera
stakeholders
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"validate", dslPath})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected validate to fail on diagnostics")
	}
	if !strings.Contains(err.Error(), "validate failed:\n5:13 - error - requirement stakeholders actors missing - [requirement.stakeholders.actors.missing]") {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output on Execute error, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIValidateCommandRejectsUnsupportedWhenVerb(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "invalid-when-verb.dsl")
	dsl := `stakeholder Worker

requirement R1
when Worker dances DangerZone
system shall notify Worker
stakeholders Worker
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"validate", dslPath})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected validate to fail on diagnostics")
	}
	if strings.Contains(err.Error(), "No validation issues found.") {
		t.Fatalf("expected unsupported when-verb diagnostic, got %v", err)
	}
	if !strings.Contains(err.Error(), "4:13 - error - requirement when verb unsupported - [requirement.when.verb.unsupported]") {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output on Execute error, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIValidateCommandDoesNotFailOnWarningsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "warning-validate.dsl")
	dsl := `stakeholder Worker
stakeholder Unused

requirement R1
system shall notify Worker
stakeholders Worker
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"validate", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected validate warning-only diagnostics to succeed, got %v", err)
	}
	if !strings.Contains(stdout.String(), "validate warnings:\n2:13 - warning - stakeholder unused - [stakeholder.unused]\n") {
		t.Fatalf("unexpected validate warning output: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLIValidateCommandLoadsWorkspaceFiles(t *testing.T) {
	tmpDir := t.TempDir()
	files := map[string]string{
		"main.dsl": `requirement R1
system shall notify Worker
stakeholders Worker
`,
		"stakeholders.dsl": "stakeholder Worker\n",
		"assignment_lone.dsl": `assignment Missing
Ghost -> privacy_pref
`,
		"_lone/ignored.dsl": "stakeholder Ignored\n",
	}
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"validate", filepath.Join(tmpDir, "main.dsl")})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected workspace validation to succeed, got %v", err)
	}
	if stdout.String() != "No validation issues found.\n" {
		t.Fatalf("unexpected validate output: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLIAnalyzeCommandReportsWorkspaceDiagnosticsAtSourceFileLine(t *testing.T) {
	tmpDir := t.TempDir()
	files := map[string]string{
		"main.dsl": `requirement R1
system shall notify Worker
stakeholders Worker
`,
		"stakeholders.dsl": "stakeholder Worker\n",
		"features/equipment.dsl": `requirement R2
system shall notify Worker
stakeholders Missing
`,
	}
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"analyze", filepath.Join(tmpDir, "main.dsl")})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected analyze to fail on workspace diagnostic")
	}
	want := filepath.Join("features", "equipment.dsl") + ":3:1 - error - requirement stakeholder unknown - [requirement.stakeholder.unknown]"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected source-file diagnostic %q, got %v", want, err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output on Execute error, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIValidateCommandTreatsLoneFilesAsStandalone(t *testing.T) {
	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.dsl")
	lonePath := filepath.Join(tmpDir, "case_lone.dsl")
	if err := os.WriteFile(mainPath, []byte("stakeholder Worker\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(lonePath, []byte(`requirement R1
system shall notify Worker
stakeholders Worker
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"validate", lonePath})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected standalone lone validation to fail")
	}
	if !strings.Contains(err.Error(), "requirement.stakeholder.unknown") {
		t.Fatalf("expected lone file to ignore workspace declarations, got %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output on Execute error, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIAnalyzeCommandPrintsStandardDiagnosticBlock(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "invalid-analyze.dsl")
	dsl := `stakeholder Worker

requirement R1
system shall track location of Worker using Camera
stakeholders
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"analyze", dslPath})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected analyze to fail on diagnostics")
	}
	if !strings.Contains(err.Error(), "analyze failed:\n5:13 - error - requirement stakeholders actors missing - [requirement.stakeholders.actors.missing]") {
		t.Fatalf("unexpected analyze error: %v", err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("expected no command output on Execute error, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestCLIAnalyzeCommandDoesNotFailOnWarningsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "warning-analyze.dsl")
	dsl := `stakeholder Worker
stakeholder Unused

requirement R1
system shall notify Worker
stakeholders Worker
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"analyze", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected analyze warning-only diagnostics to succeed, got %v", err)
	}
	if !strings.Contains(stdout.String(), "analyze warnings:\n2:13 - warning - stakeholder unused - [stakeholder.unused]\n") {
		t.Fatalf("unexpected analyze warning output: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "No value conflicts detected.\n") {
		t.Fatalf("expected analyze to continue after warnings, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLIAnalyzeCommandPrintsConflictValuePairs(t *testing.T) {
	tmpDir := t.TempDir()
	dslPath := filepath.Join(tmpDir, "conflicts.dsl")
	dsl := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91
value authority_pref = 0.02, 0.88

requirement R1
system shall track location of Worker using Camera
stakeholders Worker, Manager

assignment R1
Worker -> privacy_pref
Manager -> authority_pref
`
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"analyze", dslPath})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Manager (authority_pref) vs Worker (privacy_pref) → conflict score:") {
		t.Fatalf("expected conflict value pairs in output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunPipelineNonStrictKeepsMultipleDiagnosticsVisible(t *testing.T) {
	dsl := `stakeholder Worker

value privacy_pref = 10.8,

requirement R1
where
system shall notify
stakeholders

assignment R1
Worker -> privacy_pref
`

	result := runTestAnalysis(t, dsl, false)

	if diagnostics := result.AllDiagnostics(); len(diagnostics) < 4 {
		t.Fatalf("expected multiple diagnostics from non-strict pipeline, got %+v", diagnostics)
	}
}

func TestRunPipelineStrictStopsOnDiagnostics(t *testing.T) {
	dsl := `stakeholder Worker

value privacy_pref = 10.8,

requirement R1
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
`

	_, err := coreanalysis.Run(dsl, coreanalysis.Options{
		Strict:         true,
		BuildConflicts: true,
	})
	if err == nil {
		t.Fatal("expected strict pipeline error, got nil")
	}
	if !strings.Contains(err.Error(), "strict policy rejected result") {
		t.Fatalf("unexpected strict pipeline error: %v", err)
	}
}

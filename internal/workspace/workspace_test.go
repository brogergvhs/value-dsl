package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTextLoadsWorkspaceAndSkipsLoneFiles(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	write("main.dsl", "requirement R1\nsystem shall notify Worker\nstakeholders Worker\n")
	write("stakeholders.dsl", "stakeholder Worker\n")
	write("nested/values.dsl", "value privacy_pref = 1.58, 0.91\n")
	write("case_lone.dsl", "stakeholder LoneFile\n")
	write("_lone/isolated.dsl", "stakeholder LoneDir\n")

	text, err := LoadText(filepath.Join(root, "main.dsl"))
	if err != nil {
		t.Fatalf("LoadText() error = %v", err)
	}
	for _, want := range []string{"requirement R1", "stakeholder Worker", "value privacy_pref"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected workspace text to contain %q, got %q", want, text)
		}
	}
	for _, notWant := range []string{"LoneFile", "LoneDir"} {
		if strings.Contains(text, notWant) {
			t.Fatalf("expected workspace text to skip %q, got %q", notWant, text)
		}
	}
}

func TestLoadDocumentForOpenFilesUsesContainingWorkspace(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	write("main.dsl", "requirement R1\nsystem shall notify Worker\nstakeholders Worker\n")
	write("stakeholders.dsl", "stakeholder Worker\n")
	write("features/equipment.dsl", "requirement Old\nsystem shall notify Missing\nstakeholders Missing\n")
	write("_lone/ignored.dsl", "stakeholder Ignored\n")

	equipmentPath := filepath.Join(root, "features", "equipment.dsl")
	document, err := LoadDocumentForOpenFiles(equipmentPath, map[string]string{
		equipmentPath: "requirement R2\nsystem shall notify Worker\nstakeholders Worker\n",
	})
	if err != nil {
		t.Fatalf("LoadDocumentForOpenFiles() error = %v", err)
	}
	if !strings.HasPrefix(document.Text, "requirement R2\n") {
		t.Fatalf("expected open file text first, got %q", document.Text)
	}
	for _, want := range []string{"requirement R1", "stakeholder Worker"} {
		if !strings.Contains(document.Text, want) {
			t.Fatalf("expected workspace text to contain %q, got %q", want, document.Text)
		}
	}
	for _, notWant := range []string{"requirement Old", "Ignored"} {
		if strings.Contains(document.Text, notWant) {
			t.Fatalf("expected workspace text to skip %q, got %q", notWant, document.Text)
		}
	}
	if len(document.Files) == 0 || document.Files[0].Path != filepath.Join(root, "features", "equipment.dsl") || document.Files[0].StartLine != 1 {
		t.Fatalf("unexpected open file mapping: %+v", document.Files)
	}
}

func TestLoadDocumentUsesOutermostMainDSLAsWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	write("main.dsl", "stakeholder Outer\n")
	write("sub/main.dsl", "stakeholder Inner\n")
	write("sub/feature.dsl", "requirement R1\nsystem shall notify Outer\nstakeholders Outer\n")

	document, err := LoadDocument(filepath.Join(root, "sub", "main.dsl"))
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}
	if !strings.Contains(document.Text, "stakeholder Outer") || !strings.Contains(document.Text, "stakeholder Inner") {
		t.Fatalf("expected outer workspace to include outer and nested files, got %q", document.Text)
	}
	if len(document.Files) == 0 || document.Files[0].Path != filepath.Join(root, "main.dsl") {
		t.Fatalf("expected outer main.dsl first, got %+v", document.Files)
	}

	featurePath := filepath.Join(root, "sub", "feature.dsl")
	open, err := LoadDocumentForOpenFiles(featurePath, map[string]string{
		featurePath: "requirement R2\nsystem shall notify Outer\nstakeholders Outer\n",
	})
	if err != nil {
		t.Fatalf("LoadDocumentForOpenFiles() error = %v", err)
	}
	if !strings.Contains(open.Text, "stakeholder Outer") || !strings.Contains(open.Text, "stakeholder Inner") {
		t.Fatalf("expected open file to use outer workspace root, got %q", open.Text)
	}
	if len(open.Files) == 0 || open.Files[0].Path != filepath.Join(root, "sub", "feature.dsl") {
		t.Fatalf("expected open file first for editor diagnostics, got %+v", open.Files)
	}
}

func TestLoadTextTreatsLoneInputsAsStandalone(t *testing.T) {
	root := t.TempDir()
	mainPath := filepath.Join(root, "main.dsl")
	lonePath := filepath.Join(root, "case_lone.dsl")
	inLoneDir := filepath.Join(root, "_lone", "case.dsl")
	for path, text := range map[string]string{
		mainPath:  "stakeholder Shared\n",
		lonePath:  "stakeholder LoneFile\n",
		inLoneDir: "stakeholder LoneDir\n",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	text, err := LoadText(lonePath)
	if err != nil {
		t.Fatalf("LoadText(lone file) error = %v", err)
	}
	if strings.Contains(text, "Shared") || !strings.Contains(text, "LoneFile") {
		t.Fatalf("expected lone file to load standalone, got %q", text)
	}

	text, err = LoadText(inLoneDir)
	if err != nil {
		t.Fatalf("LoadText(lone dir) error = %v", err)
	}
	if strings.Contains(text, "Shared") || !strings.Contains(text, "LoneDir") {
		t.Fatalf("expected _lone file to load standalone, got %q", text)
	}
}

func TestLoadTextTreatsNonMainDSLAsStandalone(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "feature.dsl")
	if err := os.WriteFile(entry, []byte("requirement R1\nsystem shall notify Worker\nstakeholders Worker\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "stakeholders.dsl"), []byte("stakeholder Worker\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	text, err := LoadText(entry)
	if err != nil {
		t.Fatalf("LoadText() error = %v", err)
	}
	if strings.Contains(text, "stakeholder Worker") {
		t.Fatalf("expected non-main DSL file to load standalone, got %q", text)
	}
}

func TestLoadTextUsesContainingWorkspaceForNonMainDSL(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	write("main.dsl", "requirement R1\nsystem shall notify Worker\nstakeholders Worker\n")
	write("stakeholders.dsl", "stakeholder Worker\n")
	write("features/equipment.dsl", "value safety_pref = 0.5, 0.5\n")

	document, err := LoadDocument(filepath.Join(root, "features", "equipment.dsl"))
	if err != nil {
		t.Fatalf("LoadDocument() error = %v", err)
	}
	for _, want := range []string{"requirement R1", "stakeholder Worker", "value safety_pref"} {
		if !strings.Contains(document.Text, want) {
			t.Fatalf("expected workspace text to contain %q, got %q", want, document.Text)
		}
	}
	if len(document.Files) == 0 || document.Files[0].Path != filepath.Join(root, "main.dsl") {
		t.Fatalf("expected outer main.dsl first, got %+v", document.Files)
	}
}

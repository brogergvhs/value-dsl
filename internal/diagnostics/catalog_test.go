package diagnostics

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestCatalogCoversValidationCodes(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "validation", "validator.go"))
	if err != nil {
		t.Fatalf("read validator.go: %v", err)
	}

	re := regexp.MustCompile(`Code\w+\s+DiagnosticCode\s*=\s*"([^"]+)"`)
	matches := re.FindAllSubmatch(source, -1)
	if len(matches) == 0 {
		t.Fatal("no validation diagnostic codes found")
	}
	for _, match := range matches {
		code := string(match[1])
		if _, ok := semanticCodes[code]; !ok {
			t.Fatalf("diagnostic code %q missing from catalog", code)
		}
	}
}

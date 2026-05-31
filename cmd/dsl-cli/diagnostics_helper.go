package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/validation"
)

func renderDiagnostics(diagnostics []validation.Diagnostic, files []docindex.SourceFile) string {
	lines := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		lines = append(lines, formatDiagnostic(diagnostic, files))
	}
	return strings.Join(lines, "\n")
}

func commandFailure(command string, err error) error {
	var strictErr *analysis.StrictPolicyError
	if errors.As(err, &strictErr) {
		return fmt.Errorf("%s failed:\n%s", command, renderDiagnostics(strictErr.Diagnostics, nil))
	}
	return fmt.Errorf("%s failed: %w", command, err)
}

func reportDiagnostics(out io.Writer, command string, diagnostics []validation.Diagnostic, files []docindex.SourceFile) error {
	if len(diagnostics) == 0 {
		return nil
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == validation.SeverityError {
			return fmt.Errorf("%s failed:\n%s", command, renderDiagnostics(diagnostics, files))
		}
	}

	fmt.Fprintf(out, "%s warnings:\n", command)
	for _, diagnostic := range diagnostics {
		fmt.Fprintln(out, formatDiagnostic(diagnostic, files))
	}
	return nil
}

func formatDiagnostic(diagnostic validation.Diagnostic, files []docindex.SourceFile) string {
	if len(files) <= 1 {
		return validation.FormatDiagnostic(diagnostic)
	}
	root := workspaceRoot(files)
	for _, file := range files {
		if diagnostic.Line < file.StartLine || diagnostic.Line > file.EndLine {
			continue
		}
		diagnostic.Line -= file.StartLine - 1
		return fmt.Sprintf("%s:%s", displayPath(root, file.Path), validation.FormatDiagnostic(diagnostic))
	}
	return validation.FormatDiagnostic(diagnostic)
}

func workspaceRoot(files []docindex.SourceFile) string {
	for _, file := range files {
		if filepath.Base(file.Path) == "main.dsl" {
			return filepath.Dir(file.Path)
		}
	}
	return ""
}

func displayPath(root, path string) string {
	if root == "" {
		return path
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return path
	}
	return rel
}

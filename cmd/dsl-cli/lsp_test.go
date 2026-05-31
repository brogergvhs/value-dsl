//go:build dsl_lsp

package main

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/lsp"
)

func TestCLILSPVersionCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"lsp", "--version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got := stdout.String(); !strings.HasPrefix(got, "value-dsl-lsp ") {
		t.Fatalf("expected LSP version output, got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestCLILSPCommandPrintsConsoleMessage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	lspDebug = false
	lspVersion = false

	originalRunLSPStdio := runLSPStdio
	runLSPStdio = func(_ *slog.Logger, stderr io.Writer) error {
		_, err := io.WriteString(stderr, lsp.ConsoleMessage)
		return err
	}
	t.Cleanup(func() {
		runLSPStdio = originalRunLSPStdio
	})

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"lsp"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
	if got := stderr.String(); got != lsp.ConsoleMessage {
		t.Fatalf("expected LSP console message:\n%s\ngot:\n%s", lsp.ConsoleMessage, got)
	}
}

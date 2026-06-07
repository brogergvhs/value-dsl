package main

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/brogergvhs/value-dsl/internal/lsp"
)

func TestRunVersion(t *testing.T) {
	for _, args := range [][]string{
		{"--version"},
		{"-v"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := run(args, &stdout, &stderr)

			if code != 0 {
				t.Fatalf("expected exit code 0, got %d", code)
			}
			if got := stdout.String(); !strings.HasPrefix(got, "value-dsl-lsp ") {
				t.Fatalf("expected version output, got %q", got)
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected no stderr output, got %q", stderr.String())
			}
		})
	}
}

func TestRunRejectsSingleDashVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"-version"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
	if got := stderr.String(); !strings.Contains(got, "unknown flag: -version") {
		t.Fatalf("expected single-dash version rejection, got %q", got)
	}
}

func TestRunRejectsPositionalArguments(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"lsp"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
	got := stderr.String()
	for _, want := range []string{
		`unexpected argument "lsp"`,
		"dsl-lsp is already the language server command.",
		"Run `dsl-lsp` directly, or use `dsl lsp` with the combined CLI.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected stderr to contain %q, got %q", want, got)
		}
	}
}

func TestRunStartsLSPWithConsoleMessage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	originalRunLSPStdio := runLSPStdio
	runLSPStdio = func(_ *slog.Logger, stderr io.Writer) error {
		_, err := io.WriteString(stderr, lsp.ConsoleMessage)
		return err
	}
	t.Cleanup(func() {
		runLSPStdio = originalRunLSPStdio
	})

	code := run(nil, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
	if got := stderr.String(); got != lsp.ConsoleMessage {
		t.Fatalf("expected LSP console message:\n%s\ngot:\n%s", lsp.ConsoleMessage, got)
	}
}

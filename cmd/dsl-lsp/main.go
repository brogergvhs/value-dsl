// Package main starts the GLSP-based language server for the DSL.
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/brogergvhs/value-dsl/internal/lsp"
)

var runLSPStdio = lsp.RunStdio

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	debug := false

	for _, arg := range args {
		switch arg {
		case "--debug":
			debug = true
		case "--version", "-v":
			fmt.Fprintf(stdout, "value-dsl-lsp %s\n", lsp.Version())
			return 0
		case "--help", "-h":
			fmt.Fprintln(stdout, "Usage: dsl-lsp [--debug] [--version]")
			return 0
		default:
			if len(arg) > 0 && arg[0] == '-' {
				fmt.Fprintf(stderr, "unknown flag: %s\n", arg)
				return 2
			}
			fmt.Fprintf(stderr, "unexpected argument %q\n\ndsl-lsp is already the language server command.\nRun `dsl-lsp` directly, or use `dsl lsp` with the combined CLI.\n", arg)
			return 2
		}
	}

	var logger *slog.Logger
	if debug {
		logger = slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	if err := runLSPStdio(logger, stderr); err != nil {
		if logger != nil {
			logger.Error("lsp server stopped", "error", err)
		}
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}

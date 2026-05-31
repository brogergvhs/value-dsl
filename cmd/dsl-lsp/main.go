// Package main starts the GLSP-based language server for the DSL.
package main

import (
	"flag"
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
	flags := flag.NewFlagSet("dsl-lsp", flag.ContinueOnError)
	flags.SetOutput(stderr)

	debug := flags.Bool("debug", false, "write debug logs to stderr")
	version := flags.Bool("version", false, "print version and exit")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "unexpected argument %q\n\n", flags.Arg(0))
		fmt.Fprintln(stderr, "dsl-lsp is already the language server command.")
		fmt.Fprintln(stderr, "Run `dsl-lsp` directly, or use `dsl lsp` with the combined CLI.")
		return 2
	}

	if *version {
		fmt.Fprintf(stdout, "value-dsl-lsp %s\n", lsp.Version())
		return 0
	}

	var logger *slog.Logger
	if *debug {
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

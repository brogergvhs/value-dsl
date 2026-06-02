// Package main starts the GLSP-based language server for the DSL.
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/brogergvhs/value-dsl/internal/lsp"
	"github.com/spf13/cobra"
)

var runLSPStdio = lsp.RunStdio

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	debug := false
	exitCode := 0

	cmd := &cobra.Command{
		Use:           "dsl-lsp",
		Version:       lsp.Version(),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			return fmt.Errorf("unexpected argument %q\n\ndsl-lsp is already the language server command.\nRun `dsl-lsp` directly, or use `dsl lsp` with the combined CLI.", args[0])
		},
		Run: func(_ *cobra.Command, _ []string) {
			var logger *slog.Logger
			if debug {
				logger = slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
			}
			if err := runLSPStdio(logger, stderr); err != nil {
				if logger != nil {
					logger.Error("lsp server stopped", "error", err)
				}
				fmt.Fprintln(stderr, err.Error())
				exitCode = 1
			}
		},
	}
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetVersionTemplate("value-dsl-lsp {{.Version}}\n")
	cmd.Flags().BoolVar(&debug, "debug", false, "write debug logs to stderr")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 2
	}
	return exitCode
}

//go:build dsl_lsp

package main

import (
	"fmt"
	"log/slog"

	"github.com/brogergvhs/value-dsl/internal/lsp"

	"github.com/spf13/cobra"
)

var (
	lspDebug   bool
	lspVersion bool
)

var runLSPStdio = lsp.RunStdio

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Run the DSL language server over stdio",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if lspVersion {
			fmt.Fprintf(cmd.OutOrStdout(), "value-dsl-lsp %s\n", lsp.Version())
			return nil
		}

		var logger *slog.Logger
		if lspDebug {
			logger = slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: slog.LevelDebug}))
		}
		return runLSPStdio(logger, cmd.ErrOrStderr())
	},
}

func init() {
	lspCmd.Flags().BoolVar(&lspDebug, "debug", false, "write debug logs to stderr")
	lspCmd.Flags().BoolVar(&lspVersion, "version", false, "print LSP server version and exit")
	rootCmd.AddCommand(lspCmd)
}

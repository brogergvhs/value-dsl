package main

import (
	"fmt"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file|->",
	Short: "Validate DSL semantic model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runAnalysis(args[0], coreanalysis.Options{})
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}
		diagnostics := result.AllDiagnostics()
		if len(diagnostics) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No validation issues found.")
			return nil
		}
		return reportDiagnostics(cmd.OutOrStdout(), cmd.Name(), diagnostics, result.Index.SourceFiles)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

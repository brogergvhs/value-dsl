package main

import (
	"fmt"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/model"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <file|->",
	Short: "Run analysis pipeline and report conflicts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runAnalysis(args[0], coreanalysis.Options{BuildConflicts: true})
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}

		if diagnostics := result.AllDiagnostics(); len(diagnostics) > 0 {
			if err := reportDiagnostics(cmd.OutOrStdout(), cmd.Name(), diagnostics, result.Index.SourceFiles); err != nil {
				return err
			}
		}

		conflicts := []model.ConflictResult(nil)
		if result.Index != nil {
			conflicts = result.Index.Conflicts
		}
		if len(conflicts) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No value conflicts detected.")
			return nil
		}

		for _, c := range conflicts {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Requirement %s:\n  %s (%s) vs %s (%s) → conflict score: %.2f\n",
				c.RequirementID,
				c.StakeholderA,
				c.ValueA,
				c.StakeholderB,
				c.ValueB,
				c.ConflictScore,
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

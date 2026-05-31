package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/formatting"

	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format <file|->",
	Short: "Format DSL source canonically",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]
		dsl, err := readDSLFile(inputPath)
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}

		write, _ := cmd.Flags().GetBool("write")
		if write {
			if strings.TrimSpace(inputPath) == "-" {
				return commandFailure(cmd.Name(), fmt.Errorf("cannot use --write when reading DSL from stdin"))
			}
			formatted, err := formatting.FormatForWrite(dsl)
			if err != nil {
				return commandFailure(cmd.Name(), err)
			}
			return os.WriteFile(inputPath, []byte(formatted), 0o644)
		}

		formatted, err := formatting.Format(dsl)
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}

		fmt.Fprint(cmd.OutOrStdout(), formatted)
		return nil
	},
}

func init() {
	formatCmd.Flags().BoolP("write", "w", false, "write result back to file")
	rootCmd.AddCommand(formatCmd)
}

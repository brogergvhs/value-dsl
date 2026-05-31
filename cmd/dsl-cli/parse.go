package main

import (
	"encoding/json"

	"github.com/brogergvhs/value-dsl/internal/parser"
	"github.com/brogergvhs/value-dsl/internal/workspace"

	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse <file|->",
	Short: "Parse DSL file and print the canonical AST as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dsl, err := workspace.LoadText(args[0])
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}

		out, err := json.MarshalIndent(parser.Parse(dsl).AST, "", "  ")
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}

		_, err = cmd.OutOrStdout().Write(append(out, '\n'))
		return err
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}

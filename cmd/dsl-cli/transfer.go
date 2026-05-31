package main

import (
	"bytes"
	"os"
	"strings"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/formatting"
	"github.com/brogergvhs/value-dsl/internal/ioformat"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <file|->",
	Short: "Export DSL to an interchange format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		result, err := runAnalysis(args[0], coreanalysis.Options{})
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}
		if err := reportDiagnostics(cmd.ErrOrStderr(), cmd.Name(), result.AllDiagnostics(), result.Index.SourceFiles); err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := ioformat.Encode(format, &buf, ioformat.FromModel(result.Index.Model)); err != nil {
			return commandFailure(cmd.Name(), err)
		}

		return writeOutput(cmd, buf.Bytes())
	},
}

var importCmd = &cobra.Command{
	Use:   "import <file|->",
	Short: "Import DSL from an interchange format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		content, err := readDSLFile(args[0])
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}
		doc, err := ioformat.Decode(format, strings.NewReader(content))
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}
		dsl, err := formatting.Format(ioformat.ToDSL(doc))
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}
		result, err := coreanalysis.Run(dsl, coreanalysis.Options{})
		if err != nil {
			return commandFailure(cmd.Name(), err)
		}
		if err := reportDiagnostics(cmd.ErrOrStderr(), cmd.Name(), result.AllDiagnostics(), nil); err != nil {
			return err
		}

		return writeOutput(cmd, []byte(dsl))
	},
}

func init() {
	for _, c := range []*cobra.Command{exportCmd, importCmd} {
		c.Flags().StringP("format", "f", "json", "interchange format")
		c.Flags().StringP("output", "o", "", "output file (stdout when empty)")
		rootCmd.AddCommand(c)
	}
}

func writeOutput(cmd *cobra.Command, content []byte) error {
	output, _ := cmd.Flags().GetString("output")
	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "-" {
		_, err := cmd.OutOrStdout().Write(content)
		return err
	}

	return os.WriteFile(output, content, 0o644)
}

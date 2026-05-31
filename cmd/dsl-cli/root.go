package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/workspace"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "dsl",
	Short:         "Value-oriented monitoring DSL tool",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	return nil
}

func readDSLFile(path string) (string, error) {
	if strings.TrimSpace(path) == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read DSL from stdin: %w", err)
		}

		return string(content), nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read DSL file %q: %w", path, err)
	}

	return string(content), nil
}

func runAnalysis(dslPath string, opts coreanalysis.Options) (coreanalysis.Result, error) {
	document, err := workspace.LoadDocument(dslPath)
	if err != nil {
		return coreanalysis.Result{}, err
	}
	result, err := coreanalysis.Run(document.Text, opts)
	if result.Index != nil {
		result.Index.SourceFiles = document.Files
	}
	return result, err
}

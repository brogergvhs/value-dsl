// Package main writes the current DSL language specification to JSON.
package main

import (
	"fmt"
	"os"

	"github.com/brogergvhs/value-dsl/internal/grammar"
)

const defaultOutputPath = "generated/grammar.json"

func main() {
	outputPath := defaultOutputPath
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [output-path]\n", os.Args[0])
		os.Exit(2)
	}
	if len(os.Args) == 2 {
		outputPath = os.Args[1]
	}

	if err := grammar.WriteJSONFile(outputPath); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

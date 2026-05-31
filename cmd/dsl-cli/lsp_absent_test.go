//go:build !dsl_lsp

package main

import "testing"

func TestCLIStandaloneBuildOmitsLSPCommand(t *testing.T) {
	for _, command := range rootCmd.Commands() {
		if command.Name() == "lsp" {
			t.Fatal("standalone CLI build should not include the LSP command")
		}
	}
}

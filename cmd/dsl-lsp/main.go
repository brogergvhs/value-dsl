// Package main starts the GLSP-based language server for the DSL.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/brogergvhs/value-dsl/internal/lsp"

	glspserver "github.com/tliron/glsp/server"
)

func main() {
	debug := flag.Bool("debug", false, "write debug logs to stderr")
	version := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("value-dsl-lsp %s\n", lsp.Version())
		return
	}

	var logger *slog.Logger
	if *debug {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	server := lsp.NewServer(logger)
	if logger != nil {
		logger.Debug("starting value-dsl-lsp")
	}
	if err := glspserver.NewServer(server.Handler(), "value-dsl-lsp", false).RunStdio(); err != nil {
		if logger != nil {
			logger.Error("lsp server stopped", "error", err)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

package lsp

import (
	"fmt"
	"io"
	"log/slog"

	glspserver "github.com/tliron/glsp/server"
)

// ConsoleMessage is written to stderr when the language server starts.
const ConsoleMessage = `Hello human!

This command is intended to be run by language server clients such
as a text editor rather than being run directly in the console.

If you are seeing this in the logs of your editor you can safely
ignore this message.

If you have run ` + "`dsl lsp`" + ` yourself in your terminal then exit
this program by pressing ctrl+c.

Shamelessly stole this from the amazing Gleam lang :)
`

// RunStdio starts the language server over stdin/stdout.
func RunStdio(logger *slog.Logger, stderr io.Writer) error {
	if stderr != nil {
		if _, err := io.WriteString(stderr, ConsoleMessage); err != nil {
			return fmt.Errorf("write LSP console message: %w", err)
		}
	}

	server := NewServer(logger)
	if logger != nil {
		logger.Debug("starting " + serverName)
	}
	return glspserver.NewServer(server.Handler(), serverName, false).RunStdio()
}

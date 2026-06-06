// Package lsp provides the GLSP server and shared editor-facing infrastructure.
package lsp

import (
	"log/slog"
	"sync"
	"time"

	"github.com/brogergvhs/value-dsl/internal/config"
	"github.com/brogergvhs/value-dsl/internal/version"

	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	serverName = "value-dsl-lsp"
)

type Server struct {
	documents     *Store
	analysisCache *Cache
	handler       protocol.Handler
	debounce      time.Duration
	logger        *slog.Logger
	watchFiles    bool

	lifecycleMu sync.Mutex
	scheduleMu  sync.Mutex
	scheduled   map[string]scheduledAnalysis
	serials     map[string]uint64
}

type scheduledAnalysis struct {
	timer  *time.Timer
	serial uint64
}

func NewServer(loggers ...*slog.Logger) *Server {
	var logger *slog.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	s := &Server{
		documents:     NewStore(),
		analysisCache: NewCache(),
		debounce:      config.Load().LSPDebounce,
		logger:        logger,
		scheduled:     make(map[string]scheduledAnalysis),
		serials:       make(map[string]uint64),
	}
	s.handler = s.newHandler()
	return s
}

func (s *Server) Handler() *protocol.Handler { return &s.handler }

func Version() string { return version.String() }

func (s *Server) newHandler() protocol.Handler {
	return protocol.Handler{
		Initialize:                     s.initialize,
		Initialized:                    s.initialized,
		Shutdown:                       s.shutdown,
		Exit:                           s.exit,
		WorkspaceSymbol:                s.workspaceSymbol,
		WorkspaceDidChangeWatchedFiles: s.didChangeWatchedFiles,
		TextDocumentDidOpen:            s.didOpen,
		TextDocumentDidChange:          s.didChange,
		TextDocumentDidClose:           s.didClose,
		TextDocumentCompletion:         s.complete,
		TextDocumentHover:              s.hover,
		TextDocumentDeclaration:        s.declaration,
		TextDocumentDefinition:         s.definition,
		TextDocumentReferences:         s.references,
		TextDocumentRename:             s.rename,
		TextDocumentDocumentSymbol:     s.documentSymbol,
		TextDocumentDocumentLink:       s.documentLink,
		TextDocumentDocumentHighlight:  s.documentHighlight,
		TextDocumentFoldingRange:       s.foldingRange,
		TextDocumentSelectionRange:     s.selectionRange,
		TextDocumentFormatting:         s.format,
		TextDocumentRangeFormatting:    s.rangeFormat,
		TextDocumentSemanticTokensFull: s.semanticTokens,
	}
}

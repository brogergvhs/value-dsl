// Package lsp provides the GLSP server and shared editor-facing infrastructure.
package lsp

import (
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/config"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/lsp/semantictokens"
	"github.com/brogergvhs/value-dsl/internal/validation"
	"github.com/brogergvhs/value-dsl/internal/version"
	"github.com/brogergvhs/value-dsl/internal/workspace"

	"github.com/tliron/glsp"
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
		TextDocumentDocumentHighlight:  s.documentHighlight,
		TextDocumentFoldingRange:       s.foldingRange,
		TextDocumentSelectionRange:     s.selectionRange,
		TextDocumentFormatting:         s.format,
		TextDocumentSemanticTokensFull: s.semanticTokens,
	}
}

// initialize handles the initialize request.
func (s *Server) initialize(_ *glsp.Context, _ *protocol.InitializeParams) (any, error) {
	capabilities := s.handler.CreateServerCapabilities()
	openClose := true
	change := protocol.TextDocumentSyncKindFull
	capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{OpenClose: &openClose, Change: &change}
	capabilities.WorkspaceSymbolProvider = true
	capabilities.CompletionProvider = &protocol.CompletionOptions{}
	capabilities.HoverProvider = true
	capabilities.DeclarationProvider = true
	capabilities.DefinitionProvider = true
	capabilities.ReferencesProvider = true
	capabilities.RenameProvider = true
	capabilities.DocumentSymbolProvider = true
	capabilities.DocumentHighlightProvider = true
	capabilities.FoldingRangeProvider = true
	capabilities.SelectionRangeProvider = true
	capabilities.DocumentFormattingProvider = true
	capabilities.SemanticTokensProvider = protocol.SemanticTokensOptions{
		Legend: protocol.SemanticTokensLegend{
			TokenTypes:     semantictokens.LegendTypes(),
			TokenModifiers: semantictokens.LegendModifiers(),
		},
		Full: true,
	}
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo:   &protocol.InitializeResultServerInfo{Name: serverName, Version: &[]string{version.String()}[0]},
	}, nil
}

// initialized handles the initialized request.
func (s *Server) initialized(_ *glsp.Context, _ *protocol.InitializedParams) error { return nil }

// shutdown handles the shutdown request.
func (s *Server) shutdown(_ *glsp.Context) error { s.cancelAllScheduled(); return nil }

// exit handles the exit request.
func (s *Server) exit(_ *glsp.Context) error { return nil }

// Text document lifecycle

// didOpen handles the textDocument/didOpen request.
func (s *Server) didOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	s.cancelScheduled(params.TextDocument.URI)
	document := s.documents.Set(params.TextDocument.URI, int32(params.TextDocument.Version), params.TextDocument.Text)
	if s.logger != nil {
		s.logger.Debug("didOpen", "uri", document.URI, "version", document.Version)
	}
	publishDiagnostics(context.Notify, params.TextDocument.URI, int32(params.TextDocument.Version), s.analyzeDocument(document))
	return nil
}

// didChange handles the textDocument/didChange request.
func (s *Server) didChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	text, err := fullTextFromChanges(params.ContentChanges)
	if err != nil {
		return err
	}
	document, ok := s.documents.Update(params.TextDocument.URI, int32(params.TextDocument.Version), text)
	if !ok {
		return nil
	}
	if s.logger != nil {
		s.logger.Debug("didChange", "uri", document.URI, "version", document.Version)
	}
	s.scheduleAnalysis(context.Notify, document)
	return nil
}

// didClose handles the textDocument/didClose request.
func (s *Server) didClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	uri := params.TextDocument.URI
	s.cancelScheduled(uri)
	s.documents.Delete(uri)
	s.analysisCache.Delete(uri)
	if s.logger != nil {
		s.logger.Debug("didClose", "uri", uri)
	}
	publishDiagnostics(context.Notify, uri, 0, coreanalysis.Result{})
	return nil
}

// Analysis helpers

func (s *Server) currentDocumentAnalysis(uri protocol.DocumentUri) (Document, coreanalysis.Result, bool) {
	document, ok := s.documents.Get(uri)
	if !ok {
		return Document{}, coreanalysis.Result{}, false
	}
	return document, s.analyzeDocument(document), true
}

func (s *Server) analyzeDocument(document Document) coreanalysis.Result {
	if cached, ok := s.analysisCache.GetDocument(document); ok {
		if s.logger != nil {
			s.logger.Debug("analysis cache hit", "uri", document.URI, "version", document.Version)
		}
		return cached
	}
	if path, ok := fileURIPath(document.URI); ok {
		if workspaceDocument, err := workspace.LoadDocumentForOpenFile(path, document.Text); err == nil {
			return s.analyzeDocumentWithWorkspace(document, workspaceDocument)
		}
	}
	result, _ := coreanalysis.Run(document.Text, coreanalysis.Options{})
	s.analysisCache.PutDocument(document, result)
	if s.logger != nil {
		s.logger.Debug("analysis complete", "uri", document.URI, "version", document.Version, "diagnostics", len(result.AllDiagnostics()))
	}
	return result
}

func (s *Server) analyzeDocumentWithWorkspace(document Document, workspaceDocument workspace.Document) coreanalysis.Result {
	result, _ := coreanalysis.Run(workspaceDocument.Text, coreanalysis.Options{})
	result = withOpenFileSource(result, workspaceDocument)
	s.analysisCache.PutDocument(document, result)
	if s.logger != nil {
		s.logger.Debug("analysis complete", "uri", document.URI, "version", document.Version, "diagnostics", len(result.AllDiagnostics()))
	}
	return result
}

func withOpenFileSource(result coreanalysis.Result, workspaceDocument workspace.Document) coreanalysis.Result {
	if result.Index == nil {
		return result
	}

	index := *result.Index
	index.SourceFiles = workspaceDocument.Files
	if len(workspaceDocument.Files) > 0 {
		index.ParseDiagnostics = diagnosticsInSourceFile(index.ParseDiagnostics, workspaceDocument.Files[0])
		index.Diagnostics = diagnosticsInSourceFile(index.Diagnostics, workspaceDocument.Files[0])
	}
	result.Index = &index

	return result
}

func diagnosticsInSourceFile(diagnostics []validation.Diagnostic, file docindex.SourceFile) []validation.Diagnostic {
	if len(diagnostics) == 0 {
		return diagnostics
	}

	out := diagnostics[:0]
	for _, diagnostic := range diagnostics {
		if diagnostic.Line >= file.StartLine && diagnostic.Line <= file.EndLine {
			diagnostic.Line -= file.StartLine - 1
			out = append(out, diagnostic)
		}
	}

	return out
}

func fileURIPath(uri string) (string, bool) {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Scheme != "file" {
		return "", false
	}
	if parsed.Host != "" && parsed.Host != "localhost" {
		return "", false
	}
	if parsed.Path == "" {
		return "", false
	}

	return parsed.Path, true
}

func (s *Server) navigationAnalysis(uri string, current coreanalysis.Result) coreanalysis.Result {
	if current.BuildErr == nil && current.Index != nil && current.Index.Model != nil && len(current.Index.ParseDiagnostics) == 0 {
		return current
	}

	fallback, ok := s.analysisCache.GetNavigationFallback(uri)
	if !ok {
		return current
	}
	navigation := current
	if navigation.Index == nil {
		navigation.Index = fallback
		return navigation
	}

	navigation.Index = navigation.Index.WithNavigationDataFromFallback(fallback)
	return navigation
}

// Scheduling

func (s *Server) scheduleAnalysis(notify glsp.NotifyFunc, document Document) {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	if scheduled, ok := s.scheduled[document.URI]; ok {
		scheduled.timer.Stop()
	}
	uri := document.URI
	serial := s.nextScheduleSerialLocked(uri)
	s.scheduled[uri] = scheduledAnalysis{serial: serial}
	timer := time.AfterFunc(s.debounce, func() {
		if !s.isScheduled(uri, serial) {
			return
		}
		if current, ok := s.documents.Get(uri); !ok || current.Version != document.Version || current.Hash != document.Hash {
			return
		}
		result := s.analyzeDocument(document)
		if !s.isScheduled(uri, serial) {
			return
		}
		if current, ok := s.documents.Get(uri); !ok || current.Version != document.Version || current.Hash != document.Hash {
			return
		}
		publishDiagnostics(notify, protocol.DocumentUri(uri), document.Version, result)
		s.clearScheduled(uri, serial)
	})
	s.scheduled[uri] = scheduledAnalysis{timer: timer, serial: serial}
}

func (s *Server) cancelScheduled(uri string) {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	s.cancelScheduledLocked(uri)
}

func (s *Server) cancelScheduledLocked(uri string) {
	if scheduled, ok := s.scheduled[uri]; ok {
		if scheduled.timer != nil {
			scheduled.timer.Stop()
		}
		delete(s.scheduled, uri)
	}
	s.serials[uri]++
}

func (s *Server) cancelAllScheduled() {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	for uri := range s.scheduled {
		s.cancelScheduledLocked(uri)
	}
}

func (s *Server) nextScheduleSerialLocked(uri string) uint64 {
	s.serials[uri]++
	return s.serials[uri]
}

func (s *Server) isScheduled(uri string, serial uint64) bool {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	scheduled, ok := s.scheduled[uri]
	return ok && scheduled.serial == serial && s.serials[uri] == serial
}

func (s *Server) clearScheduled(uri string, serial uint64) {
	s.scheduleMu.Lock()
	defer s.scheduleMu.Unlock()
	if scheduled, ok := s.scheduled[uri]; ok && scheduled.serial == serial {
		delete(s.scheduled, uri)
	}
}

func fullTextFromChanges(changes []any) (string, error) {
	if len(changes) == 0 {
		return "", fmt.Errorf("didChange received no content changes")
	}

	switch change := changes[len(changes)-1].(type) {
	case protocol.TextDocumentContentChangeEventWhole:
		return change.Text, nil
	case protocol.TextDocumentContentChangeEvent:
		if change.Range == nil {
			return change.Text, nil
		}
		return "", fmt.Errorf("didChange received incremental content update but the server currently requires full text sync")
	default:
		return "", fmt.Errorf("didChange received unsupported content change payload %T", changes[len(changes)-1])
	}
}

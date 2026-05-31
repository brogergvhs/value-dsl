// Package lsp provides the GLSP server and shared editor-facing infrastructure.
package lsp

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/config"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/formatting"
	"github.com/brogergvhs/value-dsl/internal/lsp/completion"
	"github.com/brogergvhs/value-dsl/internal/lsp/semantictokens"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/validation"
	"github.com/brogergvhs/value-dsl/internal/workspace"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const (
	serverName    = "value-dsl-lsp"
	serverVersion = "0.1.0"
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

func Version() string { return serverVersion }

func (s *Server) newHandler() protocol.Handler {
	return protocol.Handler{
		Initialize:                     s.initialize,
		Initialized:                    s.initialized,
		Shutdown:                       s.shutdown,
		Exit:                           s.exit,
		TextDocumentDidOpen:            s.didOpen,
		TextDocumentDidChange:          s.didChange,
		TextDocumentDidClose:           s.didClose,
		TextDocumentCompletion:         s.complete,
		TextDocumentHover:              s.hover,
		TextDocumentDefinition:         s.definition,
		TextDocumentReferences:         s.references,
		TextDocumentRename:             s.rename,
		TextDocumentDocumentSymbol:     s.documentSymbol,
		TextDocumentFormatting:         s.format,
		TextDocumentSemanticTokensFull: s.semanticTokens,
	}
}

func (s *Server) initialize(_ *glsp.Context, _ *protocol.InitializeParams) (any, error) {
	capabilities := s.handler.CreateServerCapabilities()
	openClose := true
	change := protocol.TextDocumentSyncKindFull
	capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{OpenClose: &openClose, Change: &change}
	capabilities.CompletionProvider = &protocol.CompletionOptions{}
	capabilities.HoverProvider = true
	capabilities.DefinitionProvider = true
	capabilities.ReferencesProvider = true
	capabilities.RenameProvider = true
	capabilities.DocumentSymbolProvider = true
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
		ServerInfo:   &protocol.InitializeResultServerInfo{Name: serverName, Version: &[]string{serverVersion}[0]},
	}, nil
}

func (s *Server) initialized(_ *glsp.Context, _ *protocol.InitializedParams) error { return nil }
func (s *Server) shutdown(_ *glsp.Context) error                                   { s.cancelAllScheduled(); return nil }
func (s *Server) exit(_ *glsp.Context) error                                       { return nil }

// Text document lifecycle

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

// Feature handlers

func (s *Server) complete(_ *glsp.Context, params *protocol.CompletionParams) (any, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return protocol.CompletionList{IsIncomplete: false}, nil
	}
	return protocol.CompletionList{IsIncomplete: false, Items: completion.Items(document.Text, params.Position, s.navigationAnalysis(document.URI, result))}, nil
}

func (s *Server) hover(_ *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	return resolveHover(document.Text, params.Position, s.navigationAnalysis(document.URI, result)), nil
}

func (s *Server) definition(_ *glsp.Context, params *protocol.DefinitionParams) (any, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.Location{}, nil
	}
	return resolveDefinition(params.TextDocument.URI, params.Position, s.navigationAnalysis(document.URI, result)), nil
}

func (s *Server) references(_ *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.Location{}, nil
	}
	return resolveReferences(params.TextDocument.URI, params.Position, params.Context.IncludeDeclaration, s.navigationAnalysis(document.URI, result)), nil
}

func (s *Server) rename(_ *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return &protocol.WorkspaceEdit{}, nil
	}
	return resolveRename(params.TextDocument.URI, params.Position, params.NewName, s.navigationAnalysis(document.URI, result))
}

func (s *Server) documentSymbol(_ *glsp.Context, params *protocol.DocumentSymbolParams) (any, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.DocumentSymbol{}, nil
	}
	return buildDocumentSymbols(s.navigationAnalysis(document.URI, result)), nil
}

func (s *Server) format(_ *glsp.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	document, ok := s.documents.Get(params.TextDocument.URI)
	if !ok {
		return []protocol.TextEdit{}, nil
	}
	formatted, err := formatting.Format(document.Text)
	if err != nil {
		return nil, fmt.Errorf("format document: %w", err)
	}
	if formatted == document.Text {
		return []protocol.TextEdit{}, nil
	}
	return []protocol.TextEdit{{Range: fullDocumentRange(document.Text), NewText: formatted}}, nil
}

func (s *Server) semanticTokens(_ *glsp.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return &protocol.SemanticTokens{}, nil
	}
	return semantictokens.Resolve(document.Text, result), nil
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

// Utility

func publishDiagnostics(notify glsp.NotifyFunc, uri protocol.DocumentUri, version int32, result coreanalysis.Result) {
	if notify == nil {
		return
	}
	versionUint := protocol.UInteger(version)
	notify(string(protocol.ServerTextDocumentPublishDiagnostics), protocol.PublishDiagnosticsParams{
		URI:         uri,
		Version:     &versionUint,
		Diagnostics: diagnosticsFromResult(result),
	})
}

func fullDocumentRange(text string) protocol.Range {
	lastLine := strings.Count(text, "\n")
	lastLineText := text
	if i := strings.LastIndex(text, "\n"); i >= 0 {
		lastLineText = text[i+1:]
	}
	lastCharacter, _ := sourcepos.UTF16OffsetFromByteOK(lastLineText, len(lastLineText))
	return protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: uint32(lastLine), Character: uint32(lastCharacter)},
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

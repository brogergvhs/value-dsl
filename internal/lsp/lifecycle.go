package lsp

import (
	"github.com/brogergvhs/value-dsl/internal/lsp/semantictokens"
	"github.com/brogergvhs/value-dsl/internal/version"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// initialize handles the initialize request.
func (s *Server) initialize(_ *glsp.Context, params *protocol.InitializeParams) (any, error) {
	if workspace := params.Capabilities.Workspace; workspace != nil && workspace.DidChangeWatchedFiles != nil && workspace.DidChangeWatchedFiles.DynamicRegistration != nil {
		s.watchFiles = *workspace.DidChangeWatchedFiles.DynamicRegistration
	}
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
	capabilities.CodeActionProvider = protocol.CodeActionOptions{CodeActionKinds: []protocol.CodeActionKind{protocol.CodeActionKindQuickFix}}
	capabilities.DocumentSymbolProvider = true
	capabilities.DocumentHighlightProvider = true
	capabilities.FoldingRangeProvider = true
	capabilities.SelectionRangeProvider = true
	capabilities.DocumentFormattingProvider = true
	capabilities.DocumentRangeFormattingProvider = true
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
func (s *Server) initialized(context *glsp.Context, _ *protocol.InitializedParams) error {
	if s.watchFiles && context.Call != nil {
		go context.Call(string(protocol.ServerClientRegisterCapability), protocol.RegistrationParams{Registrations: []protocol.Registration{{
			ID:     "value-dsl-watch-dsl-files",
			Method: string(protocol.MethodWorkspaceDidChangeWatchedFiles),
			RegisterOptions: protocol.DidChangeWatchedFilesRegistrationOptions{Watchers: []protocol.FileSystemWatcher{{
				GlobPattern: "**/*.dsl",
			}}},
		}}}, &struct{}{})
	}
	return nil
}

// shutdown handles the shutdown request.
func (s *Server) shutdown(_ *glsp.Context) error { s.cancelAllScheduled(); return nil }

// exit handles the exit request.
func (s *Server) exit(_ *glsp.Context) error { return nil }

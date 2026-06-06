package lsp

import (
	"fmt"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// didOpen handles the textDocument/didOpen request.
func (s *Server) didOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	s.cancelScheduled(params.TextDocument.URI)
	document := s.documents.Set(params.TextDocument.URI, int32(params.TextDocument.Version), params.TextDocument.Text)
	affected := s.workspaceOpenDocuments(document)
	for _, affectedDocument := range affected {
		s.analysisCache.Delete(affectedDocument.URI)
	}
	if s.logger != nil {
		s.logger.Debug("didOpen", "uri", document.URI, "version", document.Version)
	}
	if !s.publishWorkspaceDiagnostics(context.Notify, document) {
		publishDiagnostics(context.Notify, params.TextDocument.URI, int32(params.TextDocument.Version), s.analyzeDocument(document))
	}
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
	affected := s.workspaceOpenDocuments(document)
	for _, affectedDocument := range affected {
		if affectedDocument.URI != document.URI {
			s.analysisCache.Delete(affectedDocument.URI)
		}
	}
	s.scheduleWorkspaceAnalysis(context.Notify, document)
	return nil
}

// didClose handles the textDocument/didClose request.
func (s *Server) didClose(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	uri := params.TextDocument.URI
	document, hadDocument := s.documents.Get(uri)
	var affected []Document
	if hadDocument {
		affected = s.workspaceOpenDocuments(document)
	}
	s.cancelScheduled(uri)
	s.documents.Delete(uri)
	s.analysisCache.Delete(uri)
	if s.logger != nil {
		s.logger.Debug("didClose", "uri", uri)
	}
	publishDiagnostics(context.Notify, uri, 0, coreanalysis.Result{})
	for _, affectedDocument := range affected {
		if affectedDocument.URI == uri {
			continue
		}
		s.analysisCache.Delete(affectedDocument.URI)
		s.scheduleWorkspaceAnalysis(context.Notify, affectedDocument)
		break
	}
	return nil
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

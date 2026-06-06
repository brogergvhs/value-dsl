package lsp

import (
	"path/filepath"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// didChangeWatchedFiles handles the workspace/didChangeWatchedFiles request.
func (s *Server) didChangeWatchedFiles(context *glsp.Context, params *protocol.DidChangeWatchedFilesParams) error {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	for _, change := range params.Changes {
		path, ok := fileURIPath(string(change.URI))
		if !ok || filepath.Ext(path) != ".dsl" {
			continue
		}
		uri := string(change.URI)
		if _, open := s.documents.Get(uri); !open {
			s.analysisCache.Delete(uri)
		}
		if change.Type == protocol.FileChangeTypeDeleted {
			publishDiagnostics(context.Notify, change.URI, -1, coreanalysis.Result{})
		}
		s.scheduleWorkspaceAnalysis(context.Notify, Document{URI: uri})
	}
	return nil
}

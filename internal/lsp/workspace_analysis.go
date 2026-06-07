package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/validation"
	"github.com/brogergvhs/value-dsl/internal/workspace"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

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
	if workspaceDocument, result, ok := s.analyzeWorkspace(document); ok {
		if file, ok := sourceFileForURI(protocol.DocumentUri(document.URI), workspaceDocument.Files); ok {
			result = withSourceFileSource(result, file)
		}
		s.analysisCache.PutDocument(document, result)
		if s.logger != nil {
			s.logger.Debug("analysis complete", "uri", document.URI, "version", document.Version, "diagnostics", len(result.AllDiagnostics()))
		}
		return result
	}
	result, _ := coreanalysis.Run(document.Text, coreanalysis.Options{})
	s.analysisCache.PutDocument(document, result)
	return result
}

func (s *Server) analyzeWorkspace(document Document) (workspace.Document, coreanalysis.Result, bool) {
	workspaceDocument, ok := s.loadWorkspace(document)
	if !ok {
		return workspace.Document{}, coreanalysis.Result{}, false
	}
	result, _ := coreanalysis.Run(workspaceDocument.Text, coreanalysis.Options{})
	if result.Index != nil {
		index := *result.Index
		index.SourceFiles = workspaceDocument.Files
		result.Index = &index
	}
	return workspaceDocument, result, true
}

func (s *Server) loadWorkspace(document Document) (workspace.Document, bool) {
	path, ok := fileURIPath(document.URI)
	if !ok {
		return workspace.Document{}, false
	}
	workspaceDocument, err := workspace.LoadDocumentForOpenFiles(path, s.openFileTextsByPath())
	if err != nil || len(workspaceDocument.Files) == 0 {
		return workspace.Document{}, false
	}
	return workspaceDocument, true
}

func (s *Server) publishWorkspaceDiagnostics(notify glsp.NotifyFunc, document Document) bool {
	workspaceDocument, result, ok := s.analyzeWorkspace(document)
	if !ok || result.Index == nil {
		return false
	}
	open := s.openDocumentsByPath()
	for _, file := range workspaceDocument.Files {
		uri := fileURI(file.Path)
		version := int32(-1)
		fileResult := withSourceFileSource(result, file)
		if openDocument, ok := open[filepath.Clean(file.Path)]; ok {
			version = openDocument.Version
			s.analysisCache.PutDocument(openDocument, fileResult)
		}
		publishDiagnostics(notify, uri, version, fileResult)
	}
	return true
}

func withSourceFileSource(result coreanalysis.Result, file docindex.SourceFile) coreanalysis.Result {
	if result.Index == nil {
		return result
	}
	index := *result.Index
	index.ParseDiagnostics = diagnosticsInSourceFile(index.ParseDiagnostics, file)
	index.Diagnostics = diagnosticsInSourceFile(index.Diagnostics, file)
	result.Index = &index
	if result.BuildErr != nil {
		line, _ := extractPosition(result.BuildErr.Error())
		if line < file.StartLine || line > file.EndLine {
			result.BuildErr = nil
		} else {
			result.BuildErr = fmt.Errorf("line %d: %s", line-file.StartLine+1, result.BuildErr.Error())
		}
	}
	return result
}

func diagnosticsInSourceFile(diagnostics []validation.Diagnostic, file docindex.SourceFile) []validation.Diagnostic {
	if len(diagnostics) == 0 {
		return diagnostics
	}

	out := make([]validation.Diagnostic, 0, len(diagnostics))
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

func (s *Server) openFileTextsByPath() map[string]string {
	documents := s.openDocumentsByPath()
	texts := make(map[string]string, len(documents))
	for path, document := range documents {
		texts[path] = document.Text
	}
	return texts
}

func (s *Server) openDocumentsByPath() map[string]Document {
	documents := s.documents.All()
	byPath := make(map[string]Document, len(documents))
	for _, document := range documents {
		path, ok := fileURIPath(document.URI)
		if ok {
			byPath[filepath.Clean(path)] = document
		}
	}
	return byPath
}

func (s *Server) workspaceOpenDocuments(document Document) []Document {
	workspaceDocument, ok := s.loadWorkspace(document)
	if !ok {
		return nil
	}
	open := s.openDocumentsByPath()
	affected := make([]Document, 0, len(open))
	for _, file := range workspaceDocument.Files {
		if document, ok := open[filepath.Clean(file.Path)]; ok {
			affected = append(affected, document)
		}
	}
	return affected
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

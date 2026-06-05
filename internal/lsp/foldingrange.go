package lsp

import (
	"sort"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/ast"
	"github.com/brogergvhs/value-dsl/internal/docindex"
	"github.com/brogergvhs/value-dsl/internal/grammar"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type sourceFold struct {
	startLine int
	endLine   int
}

// foldingRange handles the textDocument/foldingRange request.
func (s *Server) foldingRange(_ *glsp.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	document, result, ok := s.currentDocumentAnalysis(params.TextDocument.URI)
	if !ok {
		return []protocol.FoldingRange{}, nil
	}
	return buildFoldingRanges(protocol.DocumentUri(document.URI), result), nil
}

func buildFoldingRanges(uri protocol.DocumentUri, result coreanalysis.Result) []protocol.FoldingRange {
	if result.Index == nil || result.Index.AST == nil {
		return nil
	}

	file, hasFile := sourceFileForURI(uri, result.Index.SourceFiles)
	folds := sourceFolds(result.Index.AST.Declarations)
	ranges := make([]protocol.FoldingRange, 0, len(folds))
	for _, fold := range folds {
		if r, ok := protocolFoldingRange(fold, file, hasFile); ok {
			ranges = append(ranges, r)
		}
	}
	return ranges
}

func sourceFolds(declarations []ast.DeclarationNode) []sourceFold {
	folds := make([]sourceFold, 0, len(declarations))
	folds = append(folds, groupedDeclarationFolds(declarations, grammar.DeclarationKindStakeholder)...)
	folds = append(folds, groupedDeclarationFolds(declarations, grammar.DeclarationKindValue)...)

	for _, decl := range declarations {
		switch decl.Kind {
		case grammar.DeclarationKindRequirement, grammar.DeclarationKindAssignment:
			if fold, ok := declarationFold(decl); ok {
				folds = append(folds, fold)
			}
		}
	}

	sort.Slice(folds, func(i, j int) bool {
		if folds[i].startLine == folds[j].startLine {
			return folds[i].endLine < folds[j].endLine
		}
		return folds[i].startLine < folds[j].startLine
	})
	return folds
}

func groupedDeclarationFolds(declarations []ast.DeclarationNode, kind grammar.DeclarationKind) []sourceFold {
	var folds []sourceFold
	var groupStart, groupEnd, count int

	flush := func() {
		if count > 1 && groupEnd > groupStart {
			folds = append(folds, sourceFold{startLine: groupStart, endLine: groupEnd})
		}
		groupStart, groupEnd, count = 0, 0, 0
	}

	for _, decl := range declarations {
		if decl.Kind != kind {
			flush()
			continue
		}
		endLine := declarationEndLine(decl)
		if count == 0 {
			groupStart, groupEnd, count = decl.Line, endLine, 1
			continue
		}
		if decl.Line == groupEnd+1 {
			groupEnd = endLine
			count++
			continue
		}
		flush()
		groupStart, groupEnd, count = decl.Line, endLine, 1
	}
	flush()
	return folds
}

func declarationFold(decl ast.DeclarationNode) (sourceFold, bool) {
	endLine := declarationEndLine(decl)
	if endLine <= decl.Line {
		return sourceFold{}, false
	}
	return sourceFold{startLine: decl.Line, endLine: endLine}, true
}

func declarationEndLine(decl ast.DeclarationNode) int {
	endLine := decl.Range.End.Line
	for _, node := range decl.Body {
		endLine = max(endLine, node.Range.End.Line)
	}
	for _, node := range decl.Malformed {
		endLine = max(endLine, node.Range.End.Line)
	}
	return endLine
}

func sourceFileForURI(uri protocol.DocumentUri, files []docindex.SourceFile) (docindex.SourceFile, bool) {
	path, ok := fileURIPath(string(uri))
	if !ok {
		return docindex.SourceFile{}, false
	}
	for _, file := range files {
		if file.Path == path {
			return file, true
		}
	}
	return docindex.SourceFile{}, false
}

func protocolFoldingRange(fold sourceFold, file docindex.SourceFile, hasFile bool) (protocol.FoldingRange, bool) {
	startLine, endLine := fold.startLine, fold.endLine
	if hasFile {
		if startLine < file.StartLine || endLine > file.EndLine {
			return protocol.FoldingRange{}, false
		}
		startLine -= file.StartLine - 1
		endLine -= file.StartLine - 1
	}
	if endLine <= startLine {
		return protocol.FoldingRange{}, false
	}
	return protocol.FoldingRange{
		StartLine: toProtocolUint(startLine),
		EndLine:   toProtocolUint(endLine),
	}, true
}

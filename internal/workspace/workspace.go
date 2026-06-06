// Package workspace loads a DSL entry file together with sibling DSL files.
package workspace

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/brogergvhs/value-dsl/internal/docindex"
)

const rootFile = "main.dsl"

type Document struct {
	Text  string
	Files []docindex.SourceFile
}

// LoadText returns the source text visible from entry. Files below a main.dsl
// see all .dsl files in that workspace; *_lone.dsl files, files under _lone
// directories, and files outside a workspace are loaded by themselves.
func LoadText(entry string) (string, error) {
	document, err := LoadDocument(entry)
	return document.Text, err
}

func LoadDocument(entry string) (Document, error) {
	if strings.TrimSpace(entry) == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return Document{}, fmt.Errorf("read DSL from stdin: %w", err)
		}
		return Document{Text: string(content)}, nil
	}
	entry = filepath.Clean(entry)
	return loadDocument(entry, readFile)
}

// LoadDocumentForOpenFile returns the workspace text and source mapping visible
// to an open editor buffer.
func LoadDocumentForOpenFile(path, openText string) (Document, error) {
	return LoadDocumentForOpenFiles(path, openText, nil)
}

// LoadDocumentForOpenFiles returns the workspace text and source mapping visible
// to an open editor buffer, substituting any other open buffers for their
// on-disk contents.
func LoadDocumentForOpenFiles(path, openText string, openFiles map[string]string) (Document, error) {
	if strings.TrimSpace(path) == "-" {
		return Document{Text: openText}, nil
	}
	path = filepath.Clean(path)
	if standalone(path) || filepath.Ext(path) != ".dsl" {
		return Document{Text: openText}, nil
	}

	root, ok := containingRoot(path)
	if !ok {
		return Document{Text: openText}, nil
	}

	openByPath := make(map[string]string, len(openFiles)+1)
	for candidate, text := range openFiles {
		openByPath[filepath.Clean(candidate)] = text
	}
	openByPath[path] = openText

	return loadWorkspaceDocument(root, path, func(candidate string) (string, error) {
		if text, ok := openByPath[filepath.Clean(candidate)]; ok {
			return text, nil
		}
		return readFile(candidate)
	})
}

func loadDocument(entry string, read func(string) (string, error)) (Document, error) {
	if standalone(entry) || filepath.Ext(entry) != ".dsl" {
		text, err := read(entry)
		if err != nil {
			return Document{}, err
		}
		return Document{Text: text}, nil
	}
	if _, err := os.Stat(entry); err != nil {
		return Document{}, fmt.Errorf("read DSL file %q: %w", entry, err)
	}

	root, ok := containingRoot(entry)
	if !ok {
		text, err := read(entry)
		if err != nil {
			return Document{}, err
		}
		return Document{Text: text}, nil
	}
	return loadWorkspaceDocument(root, filepath.Join(root, rootFile), read)
}

func loadWorkspaceDocument(root, first string, read func(string) (string, error)) (Document, error) {
	paths := []string(nil)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "_lone" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".dsl" && !strings.HasSuffix(filepath.Base(path), "_lone.dsl") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return Document{}, fmt.Errorf("load DSL workspace %q: %w", root, err)
	}
	sort.Strings(paths)
	paths = entryFirst(paths, first)

	var b strings.Builder
	files := make([]docindex.SourceFile, 0, len(paths))
	currentLine := 1
	for i, path := range paths {
		content, err := read(path)
		if err != nil {
			return Document{}, err
		}
		if i > 0 {
			b.WriteByte('\n')
			currentLine++
		}
		startLine := currentLine
		b.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			b.WriteByte('\n')
			content += "\n"
		}
		endLine := startLine + strings.Count(content, "\n")
		files = append(files, docindex.SourceFile{
			Path:      filepath.Clean(path),
			StartLine: startLine,
			EndLine:   endLine,
		})
		currentLine = endLine
	}

	return Document{Text: b.String(), Files: files}, nil
}

func containingRoot(path string) (string, bool) {
	dir := filepath.Dir(path)
	var root string
	for {
		if _, err := os.Stat(filepath.Join(dir, rootFile)); err == nil {
			root = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return root, root != ""
		}
		dir = parent
	}
}

func entryFirst(paths []string, entry string) []string {
	for i, path := range paths {
		if filepath.Clean(path) != entry {
			continue
		}
		copy(paths[1:i+1], paths[:i])
		paths[0] = path
		return paths
	}

	return paths
}

func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read DSL file %q: %w", path, err)
	}
	return string(content), nil
}

func standalone(path string) bool {
	if strings.HasSuffix(filepath.Base(path), "_lone.dsl") {
		return true
	}
	for part := range strings.SplitSeq(filepath.Clean(path), string(filepath.Separator)) {
		if part == "_lone" {
			return true
		}
	}

	return false
}

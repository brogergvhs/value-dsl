package lsp

import (
	"hash/fnv"
	"sort"
	"sync"
)

type Document struct {
	URI     string
	Version int32
	Text    string
	Hash    uint64
}

type Store struct {
	mu        sync.RWMutex
	documents map[string]Document
}

func NewStore() *Store {
	return &Store{
		documents: make(map[string]Document),
	}
}

func (s *Store) Set(uri string, version int32, text string) Document {
	s.mu.Lock()
	defer s.mu.Unlock()

	document := newDocument(uri, version, text)
	s.documents[uri] = document
	return document
}

func (s *Store) Update(uri string, version int32, text string) (Document, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.documents[uri]; !ok {
		return Document{}, false
	}
	document := newDocument(uri, version, text)
	s.documents[uri] = document
	return document, true
}

func (s *Store) Get(uri string) (Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	document, ok := s.documents[uri]
	return document, ok
}

func (s *Store) Delete(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.documents, uri)
}

func (s *Store) All() []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	documents := make([]Document, 0, len(s.documents))
	for _, document := range s.documents {
		documents = append(documents, document)
	}
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].URI < documents[j].URI
	})
	return documents
}

func newDocument(uri string, version int32, text string) Document {
	return Document{
		URI:     uri,
		Version: version,
		Text:    text,
		Hash:    hashText(text),
	}
}

func hashText(text string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	return h.Sum64()
}

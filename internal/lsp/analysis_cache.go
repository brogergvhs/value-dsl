package lsp

import (
	"sync"
	"time"

	"github.com/brogergvhs/value-dsl/internal/docindex"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
)

const maxCachedDocuments = 256

type cachedResult struct {
	Version  int32
	Hash     uint64
	Result   coreanalysis.Result
	Fallback *docindex.Document // last good index; may be from an older version
	Accessed time.Time
}

type Cache struct {
	mu      sync.RWMutex
	results map[string]cachedResult
}

func NewCache() *Cache {
	return &Cache{results: make(map[string]cachedResult)}
}

func (c *Cache) GetDocument(document Document) (coreanalysis.Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cached, ok := c.results[document.URI]
	if !ok || cached.Version != document.Version || cached.Hash != document.Hash {
		return coreanalysis.Result{}, false
	}
	cached.Accessed = time.Now()
	c.results[document.URI] = cached
	return cached.Result, true
}

func (c *Cache) PutDocument(document Document, result coreanalysis.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if cached, ok := c.results[document.URI]; ok && cached.Version > document.Version {
		return
	}
	fallback := c.results[document.URI].Fallback
	if result.BuildErr == nil && result.Index != nil && result.Index.Model != nil && len(result.Index.ParseDiagnostics) == 0 {
		fallback = result.Index
	}
	c.results[document.URI] = cachedResult{Version: document.Version, Hash: document.Hash, Result: result, Fallback: fallback, Accessed: now}
	c.pruneLocked()
}

// GetNavigationFallback returns the last successfully indexed semantic view
// that navigation features may reuse while the current document is syntactically broken.
func (c *Cache) GetNavigationFallback(uri string) (*docindex.Document, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	index := c.results[uri].Fallback
	return index, index != nil
}

func (c *Cache) Delete(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.results, uri)
}

func (c *Cache) pruneLocked() {
	for len(c.results) > maxCachedDocuments {
		var oldestURI string
		var oldest time.Time
		for uri, cached := range c.results {
			if oldestURI == "" || cached.Accessed.Before(oldest) {
				oldestURI = uri
				oldest = cached.Accessed
			}
		}
		delete(c.results, oldestURI)
	}
}

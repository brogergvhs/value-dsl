package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func run(parent context.Context, cfg config, opts options) (runResult, error) {
	if opts.runs < 1 {
		return runResult{}, errors.New("-runs must be at least 1")
	}
	if opts.count < 1 {
		return runResult{}, errors.New("-count must be at least 1")
	}
	if opts.warmup < 0 {
		return runResult{}, errors.New("-warmup must be at least 0")
	}
	rootURI := fileURI(cfg.Root)
	fileURI := fileURI(cfg.File)
	data, err := os.ReadFile(cfg.File)
	if err != nil {
		return runResult{}, fmt.Errorf("read benchmark file: %w", err)
	}
	text := string(data)

	for warmupIndex := range opts.warmup {
		if _, err := runOnce(parent, cfg, rootURI, fileURI, text, 1); err != nil {
			return runResult{}, fmt.Errorf("warmup %d: %w", warmupIndex+1, err)
		}
	}

	samples := make(map[string][]float64)
	for runIndex := range opts.runs {
		runSamples, err := runOnce(parent, cfg, rootURI, fileURI, text, opts.count)
		if err != nil {
			return runResult{}, fmt.Errorf("run %d: %w", runIndex+1, err)
		}
		for name, values := range runSamples {
			samples[name] = append(samples[name], values...)
		}
	}

	metrics := make([]metric, 0, len(samples)*2)
	for _, name := range sortedSampleNames(samples) {
		if strings.HasSuffix(name, "_error_count") {
			metrics = append(metrics, metric{Name: name, N: sumInts(samples[name])})
			continue
		}
		metrics = append(metrics, metric{Name: name + "_median_ms", MS: percentile(samples[name], 0.50)})
		metrics = append(metrics, metric{Name: name + "_p95_ms", MS: percentile(samples[name], 0.95)})
	}
	return runResult{Name: cfg.Name, Count: opts.count, Runs: opts.runs, Warmup: opts.warmup, Metrics: metrics}, nil
}

func runOnce(parent context.Context, cfg config, rootURI, fileURI, text string, count int) (map[string][]float64, error) {
	ctx, cancel := context.WithTimeout(parent, time.Duration(cfg.TimeoutMS)*time.Millisecond)
	defer cancel()

	start := time.Now()
	c, err := startClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer c.close()

	initializeStart := time.Now()
	if _, err := c.request(ctx, "initialize", map[string]any{
		"processId": os.Getpid(),
		"rootPath":  cfg.Root,
		"rootUri":   rootURI,
		"workspaceFolders": []map[string]any{
			{"uri": rootURI, "name": filepath.Base(cfg.Root)},
		},
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"completion": map[string]any{"completionItem": map[string]any{}},
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}
	initializeMS := elapsedMS(initializeStart)
	spawnToInitializeMS := elapsedMS(start)
	if err := c.notify("initialized", map[string]any{}); err != nil {
		return nil, fmt.Errorf("initialized: %w", err)
	}

	openStart := time.Now()
	if err := c.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri":        fileURI,
			"languageId": cfg.LanguageID,
			"version":    1,
			"text":       text,
		},
	}); err != nil {
		return nil, fmt.Errorf("didOpen: %w", err)
	}

	samples := map[string][]float64{
		"spawn_to_initialize": {spawnToInitializeMS},
		"initialize_response": {initializeMS},
	}
	if cfg.waitDiagnostics() {
		if err := c.waitDiagnostics(ctx); err != nil {
			return nil, err
		}
		samples["did_open_to_diagnostics"] = []float64{elapsedMS(openStart)}
		samples["spawn_to_diagnostics"] = []float64{elapsedMS(start)}
	} else {
		samples["did_open_notification"] = []float64{elapsedMS(openStart)}
		samples["spawn_to_open_notification"] = []float64{elapsedMS(start)}
	}

	for _, name := range sortedRequestNames(cfg.Positions) {
		ms, errors, err := benchmarkRequest(ctx, c, fileURI, name, cfg.Positions[name], count)
		if err != nil {
			return nil, err
		}
		if len(ms) > 0 {
			samples[name] = append(samples[name], ms...)
		}
		if errors > 0 {
			samples[name+"_error_count"] = append(samples[name+"_error_count"], float64(errors))
		}
	}

	_, _ = c.request(ctx, "shutdown", nil)
	_ = c.notify("exit", nil)
	return samples, nil
}

func benchmarkRequest(ctx context.Context, c *client, uri, name string, pos position, count int) ([]float64, int, error) {
	params := map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     pos,
	}
	if name == "references" {
		params["context"] = map[string]any{"includeDeclaration": true}
	}

	samples := make([]float64, 0, count)
	errorCount := 0
	method := "textDocument/" + name
	for range count {
		start := time.Now()
		if _, err := c.request(ctx, method, params); err != nil {
			if errors.Is(err, errRPC) {
				errorCount++
				continue
			}
			return nil, errorCount, fmt.Errorf("%s: %w", method, err)
		}
		samples = append(samples, elapsedMS(start))
	}
	return samples, errorCount, nil
}

func sortedRequestNames(positions map[string]position) []string {
	names := make([]string, 0, len(positions))
	for name := range positions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedSampleNames(samples map[string][]float64) []string {
	names := make([]string, 0, len(samples))
	for name := range samples {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	values = append([]float64(nil), values...)
	sort.Float64s(values)
	idx := int(math.Ceil(p*float64(len(values)))) - 1
	idx = max(0, min(idx, len(values)-1))
	return values[idx]
}

func sumInts(values []float64) int {
	total := 0
	for _, value := range values {
		total += int(value)
	}
	return total
}

func elapsedMS(start time.Time) float64 {
	return float64(time.Since(start)) / float64(time.Millisecond)
}

func fileURI(path string) string {
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}

package main

import (
	"context"
	"flag"
)

func main() {
	configPath := flag.String("config", "", "path to lspbench JSON config or directory")
	count := flag.Int("count", 30, "number of warm requests per method")
	runs := flag.Int("runs", 1, "number of full server lifecycle runs")
	warmup := flag.Int("warmup", 1, "number of unmeasured full server lifecycle runs before measurement")
	format := flag.String("format", "text", "output format: text, json, or csv")
	graph := flag.String("graph", "", "write a PNG graph to this path")
	flag.Parse()

	if *format != "text" && *format != "json" && *format != "csv" {
		failf("unsupported -format %q", *format)
	}
	if *configPath == "" {
		failf("missing -config")
	}
	validateGraphPath(*graph)
	opts := options{count: *count, runs: *runs, warmup: *warmup}
	paths, err := configPaths(*configPath)
	if err != nil {
		failf("%v", err)
	}
	if len(paths) == 1 {
		result, failure, ok := runConfig(context.Background(), paths[0], opts)
		if !ok {
			failf("%s", failure.Error)
		}
		batch := batchResult{Results: []runResult{result}}
		printResults(*format, batch, false)
		writeGraph(*graph, batch)
		return
	}

	if *format != "json" {
		writeGraph(*graph, streamResults(*format, paths, opts))
		return
	}

	batch := collectResults(context.Background(), paths, opts)
	printResults(*format, batch, true)
	writeGraph(*graph, batch)
}

func runConfig(ctx context.Context, path string, opts options) (runResult, runFailure, bool) {
	cfg, err := loadConfig(path)
	if err != nil {
		return runResult{}, runFailure{Config: path, Error: err.Error()}, false
	}
	result, err := run(ctx, cfg, opts)
	if err != nil {
		return runResult{}, runFailure{Config: path, Error: err.Error()}, false
	}
	return result, runFailure{}, true
}

func collectResults(ctx context.Context, paths []string, opts options) batchResult {
	batch := batchResult{}
	for _, path := range paths {
		result, failure, ok := runConfig(ctx, path, opts)
		if ok {
			batch.Results = append(batch.Results, result)
		} else {
			batch.Errors = append(batch.Errors, failure)
		}
	}
	return batch
}

func streamResults(format string, paths []string, opts options) batchResult {
	if format == "csv" {
		printCSVHeader()
	}
	batch := batchResult{}
	printed := false
	for _, path := range paths {
		result, failure, ok := runConfig(context.Background(), path, opts)
		printOne(format, result, failure, ok, printed)
		if ok {
			batch.Results = append(batch.Results, result)
		} else {
			batch.Errors = append(batch.Errors, failure)
		}
		printed = true
	}
	return batch
}

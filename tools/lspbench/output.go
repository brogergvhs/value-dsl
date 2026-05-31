package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func printResults(format string, batch batchResult, multi bool) {
	switch format {
	case "json":
		value := any(batch)
		if !multi && len(batch.Results) == 1 && len(batch.Errors) == 0 {
			value = batch.Results[0]
		}
		out, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			failf("marshal result: %v", err)
		}
		fmt.Println(string(out))
	case "csv":
		printCSV(batch)
	case "text":
		printText(batch)
	default:
		failf("unsupported -format %q", format)
	}
}

func printText(batch batchResult) {
	for i, result := range batch.Results {
		if i > 0 {
			fmt.Println()
		}
		printTextResult(result)
	}
	for _, failure := range batch.Errors {
		if len(batch.Results) > 0 {
			fmt.Println()
		}
		printTextFailure(failure)
	}
}

func printCSV(batch batchResult) {
	printCSVHeader()
	for _, result := range batch.Results {
		printCSVResult(result)
	}
	for _, failure := range batch.Errors {
		printCSVFailure(failure)
	}
}

func printOne(format string, result runResult, failure runFailure, ok bool, leadingBlank bool) {
	if format == "text" && leadingBlank {
		fmt.Println()
	}
	switch format {
	case "csv":
		if ok {
			printCSVResult(result)
		} else {
			printCSVFailure(failure)
		}
	case "text":
		if ok {
			printTextResult(result)
		} else {
			printTextFailure(failure)
		}
	default:
		failf("unsupported -format %q", format)
	}
}

func printTextResult(result runResult) {
	fmt.Printf("%s runs=%d count=%d warmup=%d\n", result.Name, result.Runs, result.Count, result.Warmup)
	for _, metric := range result.Metrics {
		if metric.N != 0 || strings.HasSuffix(metric.Name, "_count") {
			fmt.Printf("%-34s %d\n", metric.Name, metric.N)
			continue
		}
		fmt.Printf("%-34s %.3f ms\n", metric.Name, metric.MS)
	}
}

func printTextFailure(failure runFailure) {
	fmt.Printf("%s failed: %s\n", failure.Config, failure.Error)
}

func printCSVHeader() {
	fmt.Println("name,runs,count,warmup,metric,ms,n,error")
}

func printCSVResult(result runResult) {
	for _, metric := range result.Metrics {
		fmt.Printf("%s,%d,%d,%d,%s,%.6f,%d,\n", result.Name, result.Runs, result.Count, result.Warmup, metric.Name, metric.MS, metric.N)
	}
}

func printCSVFailure(failure runFailure) {
	fmt.Printf("%s,0,0,0,error,0,0,%q\n", failure.Config, failure.Error)
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

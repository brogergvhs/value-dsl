package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type graphLayout struct {
	left, top, plotW, barH, barGap, groupH, width, height int
	maxValue                                              float64
}

func writeGraph(path string, batch batchResult) {
	if path == "" {
		return
	}
	if len(batch.Results) == 0 {
		failf("no successful results to graph")
	}
	file, err := os.Create(path)
	if err != nil {
		failf("write graph: %v", err)
	}
	if err := png.Encode(file, graphImage(batch.Results)); err != nil {
		_ = file.Close()
		failf("write graph: %v", err)
	}
	if err := file.Close(); err != nil {
		failf("write graph: %v", err)
	}
}

func validateGraphPath(path string) {
	if path != "" && strings.ToLower(filepath.Ext(path)) != ".png" {
		failf("unsupported graph extension %q (use .png)", filepath.Ext(path))
	}
}

func graphImage(results []runResult) image.Image {
	names := graphMetricNames(results)
	layout := newGraphLayout(results, names)
	img := image.NewRGBA(image.Rect(0, 0, layout.width, layout.height))
	fillRect(img, img.Bounds(), color.White)
	drawString(img, 24, 34, "LSP benchmark", textColor)
	drawString(img, 24, 56, "bar width uses log10(1 + ms); labels show measured milliseconds", mutedColor)
	drawString(img, 24, 74, "open-only = didOpen notification timing; diagnostics were not waited for", mutedColor)
	drawLegend(img, results, layout.left, 24)
	for i, name := range names {
		drawMetricGroup(img, layout, results, name, i)
	}
	return img
}

func drawMetricGroup(img *image.RGBA, layout graphLayout, results []runResult, name string, i int) {
	y := layout.top + i*layout.groupH
	drawString(img, 24, y+len(results)*(layout.barH+layout.barGap)/2, name, textColor)
	fillRect(img, image.Rect(layout.left, y-8, layout.left+layout.plotW, y-7), gridColor)
	for j, result := range results {
		value, note, ok := graphMetricValue(result, name)
		if !ok {
			continue
		}
		barY := y + j*(layout.barH+layout.barGap)
		barW := int(math.Round(graphScale(value) / layout.maxValue * float64(layout.plotW)))
		fillRect(img, image.Rect(layout.left, barY, layout.left+max(1, barW), barY+layout.barH), graphColors[j%len(graphColors)])
		drawString(img, layout.left+barW+6, barY+layout.barH-3, graphValueLabel(value, note), textColor)
	}
}

func newGraphLayout(results []runResult, names []string) graphLayout {
	const left, top, plotW, right, bottom, barH, barGap, groupGap = 270, 104, 760, 190, 44, 14, 4, 18
	layout := graphLayout{
		left:     left,
		top:      top,
		plotW:    plotW,
		barH:     barH,
		barGap:   barGap,
		groupH:   len(results)*(barH+barGap) + groupGap,
		width:    left + plotW + right,
		maxValue: 1,
	}
	layout.height = top + bottom + max(1, len(names))*layout.groupH
	for _, result := range results {
		for _, name := range names {
			if value, _, ok := graphMetricValue(result, name); ok {
				layout.maxValue = math.Max(layout.maxValue, graphScale(value))
			}
		}
	}
	return layout
}

func drawLegend(img *image.RGBA, results []runResult, x, y int) {
	for i, result := range results {
		itemX := x + (i%4)*170
		itemY := y + (i/4)*20
		fillRect(img, image.Rect(itemX, itemY-10, itemX+12, itemY+2), graphColors[i%len(graphColors)])
		drawString(img, itemX+18, itemY, result.Name, mutedColor)
	}
}

func graphMetricNames(results []runResult) []string {
	seen := map[string]bool{}
	for _, result := range results {
		for _, metric := range result.Metrics {
			if metric.N == 0 && !strings.HasSuffix(metric.Name, "_count") {
				seen[graphMetricName(metric.Name)] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func graphMetricName(name string) string {
	switch name {
	case "did_open_to_diagnostics_median_ms", "did_open_notification_median_ms":
		return "did_open_ready_median_ms"
	case "did_open_to_diagnostics_p95_ms", "did_open_notification_p95_ms":
		return "did_open_ready_p95_ms"
	case "spawn_to_diagnostics_median_ms", "spawn_to_open_notification_median_ms":
		return "spawn_to_ready_median_ms"
	case "spawn_to_diagnostics_p95_ms", "spawn_to_open_notification_p95_ms":
		return "spawn_to_ready_p95_ms"
	default:
		return name
	}
}

func graphMetricValue(result runResult, name string) (float64, string, bool) {
	switch name {
	case "did_open_ready_median_ms":
		return comparableMetricValue(result, "did_open_to_diagnostics_median_ms", "did_open_notification_median_ms")
	case "did_open_ready_p95_ms":
		return comparableMetricValue(result, "did_open_to_diagnostics_p95_ms", "did_open_notification_p95_ms")
	case "spawn_to_ready_median_ms":
		return comparableMetricValue(result, "spawn_to_diagnostics_median_ms", "spawn_to_open_notification_median_ms")
	case "spawn_to_ready_p95_ms":
		return comparableMetricValue(result, "spawn_to_diagnostics_p95_ms", "spawn_to_open_notification_p95_ms")
	default:
		value, ok := metricValue(result, name)
		return value, "", ok
	}
}

func comparableMetricValue(result runResult, diagnosticsName, notificationName string) (float64, string, bool) {
	if value, ok := metricValue(result, diagnosticsName); ok {
		return value, "", true
	}
	if value, ok := metricValue(result, notificationName); ok {
		return value, "open-only", true
	}
	return 0, "", false
}

func metricValue(result runResult, name string) (float64, bool) {
	for _, metric := range result.Metrics {
		if metric.Name == name {
			return metric.MS, true
		}
	}
	return 0, false
}

func graphValueLabel(value float64, note string) string {
	if note == "" {
		return fmt.Sprintf("%.3f", value)
	}
	return fmt.Sprintf("%.3f %s", value, note)
}

func graphScale(ms float64) float64 {
	return math.Log10(1 + math.Max(0, ms))
}

func fillRect(img draw.Image, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect, &image.Uniform{C: c}, image.Point{}, draw.Src)
}

func drawString(img draw.Image, x, y int, text string, c color.Color) {
	(&font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{C: c},
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}).DrawString(text)
}

var (
	textColor   = color.RGBA{31, 41, 55, 255}
	mutedColor  = color.RGBA{75, 85, 99, 255}
	gridColor   = color.RGBA{229, 231, 235, 255}
	graphColors = []color.RGBA{
		{37, 99, 235, 255},
		{220, 38, 38, 255},
		{5, 150, 105, 255},
		{217, 119, 6, 255},
		{124, 58, 237, 255},
		{8, 145, 178, 255},
		{190, 18, 60, 255},
		{77, 124, 15, 255},
	}
)

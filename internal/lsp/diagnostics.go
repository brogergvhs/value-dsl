package lsp

import (
	"regexp"
	"strconv"

	coreanalysis "github.com/brogergvhs/value-dsl/internal/analysis"
	"github.com/brogergvhs/value-dsl/internal/grammar"
	"github.com/brogergvhs/value-dsl/internal/sourcepos"
	"github.com/brogergvhs/value-dsl/internal/validation"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var parseLocationPattern = regexp.MustCompile(`:\s*([0-9]+):([0-9]+):`)
var linePattern = regexp.MustCompile(`line ([0-9]+)`)

// publishDiagnostics sends the textDocument/publishDiagnostics notification.
func publishDiagnostics(notify glsp.NotifyFunc, uri protocol.DocumentUri, version int32, result coreanalysis.Result) {
	if notify == nil {
		return
	}
	var versionPtr *protocol.UInteger
	if version >= 0 {
		versionUint := protocol.UInteger(version)
		versionPtr = &versionUint
	}
	notify(string(protocol.ServerTextDocumentPublishDiagnostics), protocol.PublishDiagnosticsParams{
		URI:         uri,
		Version:     versionPtr,
		Diagnostics: diagnosticsFromResult(result),
	})
}

func diagnosticsFromResult(result coreanalysis.Result) []protocol.Diagnostic {
	if result.Index == nil {
		if result.BuildErr != nil {
			return []protocol.Diagnostic{buildErrorDiagnostic(nil, result.BuildErr.Error())}
		}
		return []protocol.Diagnostic{}
	}

	diagnostics := make([]protocol.Diagnostic, 0, len(result.Index.ParseDiagnostics)+len(result.Index.Diagnostics)+1)
	for _, diagnostic := range result.Index.ParseDiagnostics {
		diagnostics = append(diagnostics, validationDiagnostic(result.Index.Lines, diagnostic))
	}

	if result.BuildErr != nil {
		diagnostics = append(diagnostics, buildErrorDiagnostic(result.Index.Lines, result.BuildErr.Error()))
		return diagnostics
	}

	for _, diagnostic := range result.Index.Diagnostics {
		diagnostics = append(diagnostics, validationDiagnostic(result.Index.Lines, diagnostic))
	}

	return diagnostics
}

func buildErrorDiagnostic(lines []grammar.TokenLine, message string) protocol.Diagnostic {
	line, column := extractPosition(message)
	severity := protocol.DiagnosticSeverityError
	source := "value-dsl"
	return protocol.Diagnostic{
		Range:    toProtocolRangeInLines(lines, sourcepos.NewRange(sourcepos.NewPosition(line, column), sourcepos.NewPosition(line, column))),
		Message:  message,
		Severity: &severity,
		Source:   &source,
	}
}

func validationDiagnostic(lines []grammar.TokenLine, diagnostic validation.Diagnostic) protocol.Diagnostic {
	diagnostic = validation.NormalizeDiagnostic(diagnostic)
	line, column := max(diagnostic.Line, 1), max(diagnostic.Column, 1)
	d := protocol.Diagnostic{
		Range:   toProtocolRangeInLines(lines, sourcepos.NewRange(sourcepos.NewPosition(line, column), sourcepos.NewPosition(line, column))),
		Message: diagnostic.Message,
	}

	if diagnostic.Code != "" {
		lspCode := protocol.IntegerOrString{Value: string(diagnostic.Code)}
		d.Code = &lspCode
	}

	severity := protocol.DiagnosticSeverityError
	if diagnostic.Severity == validation.SeverityWarning {
		severity = protocol.DiagnosticSeverityWarning
	}

	d.Severity = &severity
	source := string(diagnostic.Source)
	if source == "" {
		source = "value-dsl"
	}

	d.Source = &source
	return d
}

func extractPosition(message string) (int, int) {
	if matches := parseLocationPattern.FindStringSubmatch(message); len(matches) == 3 {
		line, lineErr := strconv.Atoi(matches[1])
		column, columnErr := strconv.Atoi(matches[2])
		if lineErr == nil && columnErr == nil {
			return line, column
		}
	}

	if matches := linePattern.FindStringSubmatch(message); len(matches) == 2 {
		line, err := strconv.Atoi(matches[1])
		if err == nil {
			return line, 1
		}
	}

	return 1, 1
}

package lsp

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brogergvhs/value-dsl/internal/validation"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TestInitializeAdvertisesFullTextSync(t *testing.T) {
	server := NewServer()

	resultValue, err := server.initialize(&glsp.Context{}, &protocol.InitializeParams{})
	if err != nil {
		t.Fatalf("initialize() error = %v", err)
	}

	result, ok := resultValue.(protocol.InitializeResult)
	if !ok {
		t.Fatalf("unexpected initialize result type: %T", resultValue)
	}

	syncOptions, ok := result.Capabilities.TextDocumentSync.(protocol.TextDocumentSyncOptions)
	if !ok {
		t.Fatalf("unexpected textDocumentSync type: %T", result.Capabilities.TextDocumentSync)
	}
	if syncOptions.Change == nil || *syncOptions.Change != protocol.TextDocumentSyncKindFull {
		t.Fatalf("unexpected sync mode: %+v", syncOptions)
	}
	if syncOptions.OpenClose == nil || !*syncOptions.OpenClose {
		t.Fatalf("expected openClose sync support, got %+v", syncOptions)
	}
	workspaceSymbolProvider, ok := result.Capabilities.WorkspaceSymbolProvider.(bool)
	if !ok || !workspaceSymbolProvider {
		t.Fatalf("expected workspace symbol support, got %+v", result.Capabilities.WorkspaceSymbolProvider)
	}
	hoverProvider, ok := result.Capabilities.HoverProvider.(bool)
	if !ok || !hoverProvider {
		t.Fatalf("expected hover support, got %+v", result.Capabilities.HoverProvider)
	}
	definitionProvider, ok := result.Capabilities.DefinitionProvider.(bool)
	if !ok || !definitionProvider {
		t.Fatalf("expected definition support, got %+v", result.Capabilities.DefinitionProvider)
	}
	declarationProvider, ok := result.Capabilities.DeclarationProvider.(bool)
	if !ok || !declarationProvider {
		t.Fatalf("expected declaration support, got %+v", result.Capabilities.DeclarationProvider)
	}
	referencesProvider, ok := result.Capabilities.ReferencesProvider.(bool)
	if !ok || !referencesProvider {
		t.Fatalf("expected references support, got %+v", result.Capabilities.ReferencesProvider)
	}
	renameProvider, ok := result.Capabilities.RenameProvider.(bool)
	if !ok || !renameProvider {
		t.Fatalf("expected rename support, got %+v", result.Capabilities.RenameProvider)
	}
	documentSymbolProvider, ok := result.Capabilities.DocumentSymbolProvider.(bool)
	if !ok || !documentSymbolProvider {
		t.Fatalf("expected document symbol support, got %+v", result.Capabilities.DocumentSymbolProvider)
	}
	if result.Capabilities.DocumentLinkProvider == nil {
		t.Fatalf("expected document link support")
	}
	documentHighlightProvider, ok := result.Capabilities.DocumentHighlightProvider.(bool)
	if !ok || !documentHighlightProvider {
		t.Fatalf("expected document highlight support, got %+v", result.Capabilities.DocumentHighlightProvider)
	}
	foldingRangeProvider, ok := result.Capabilities.FoldingRangeProvider.(bool)
	if !ok || !foldingRangeProvider {
		t.Fatalf("expected folding range support, got %+v", result.Capabilities.FoldingRangeProvider)
	}
	selectionRangeProvider, ok := result.Capabilities.SelectionRangeProvider.(bool)
	if !ok || !selectionRangeProvider {
		t.Fatalf("expected selection range support, got %+v", result.Capabilities.SelectionRangeProvider)
	}
	formattingProvider, ok := result.Capabilities.DocumentFormattingProvider.(bool)
	if !ok || !formattingProvider {
		t.Fatalf("expected formatting support, got %+v", result.Capabilities.DocumentFormattingProvider)
	}
	rangeFormattingProvider, ok := result.Capabilities.DocumentRangeFormattingProvider.(bool)
	if !ok || !rangeFormattingProvider {
		t.Fatalf("expected range formatting support, got %+v", result.Capabilities.DocumentRangeFormattingProvider)
	}
	semanticProvider, ok := result.Capabilities.SemanticTokensProvider.(protocol.SemanticTokensOptions)
	if !ok {
		t.Fatalf("expected semantic token options, got %+v", result.Capabilities.SemanticTokensProvider)
	}
	if semanticProvider.Full != true {
		t.Fatalf("expected full semantic token support, got %+v", semanticProvider.Full)
	}
	if len(semanticProvider.Legend.TokenTypes) == 0 || len(semanticProvider.Legend.TokenModifiers) == 0 {
		t.Fatalf("expected semantic token legend, got %+v", semanticProvider.Legend)
	}
}

func TestDidOpenPublishesValidationDiagnostics(t *testing.T) {
	server := NewServer()
	var published protocol.PublishDiagnosticsParams
	var method string

	context := &glsp.Context{
		Notify: func(m string, params any) {
			method = m
			published = params.(protocol.PublishDiagnosticsParams)
		},
	}

	err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///spec.dsl",
			Version: 1,
			Text: `stakeholder Worker

requirement R1
when Worker enters Danger Zone
system shall notify Worker
stakeholders Worker
`,
		},
	})
	if err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	if method != string(protocol.ServerTextDocumentPublishDiagnostics) {
		t.Fatalf("unexpected notification method: %q", method)
	}
	if published.URI != "file:///spec.dsl" {
		t.Fatalf("unexpected uri: %q", published.URI)
	}
	if len(published.Diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(published.Diagnostics))
	}
}

func TestDidOpenUsesWorkspaceFilesForMainDSLDiagnostics(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	mainText := `requirement R1
system shall notify Worker
stakeholders Worker, Supervisor

requirement R2
system shall log location of Worker using Database
stakeholders Worker, SafetyOfficer
`
	write("main.dsl", mainText)
	write("stakeholders.dsl", "stakeholder Worker\nstakeholder Supervisor\nstakeholder SafetyOfficer\n")
	write("values/preferences.dsl", "value privacy_pref = 1.58, 0.91\n")
	write("case_lone.dsl", "stakeholder IgnoredLone\n")
	write("_lone/ignored.dsl", "stakeholder IgnoredDir\n")

	server := NewServer()
	var published protocol.PublishDiagnosticsParams
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			published = params.(protocol.PublishDiagnosticsParams)
		},
	}
	uri := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "main.dsl")}).String())

	if err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     uri,
			Version: 1,
			Text:    mainText,
		},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	if len(published.Diagnostics) != 0 {
		t.Fatalf("expected workspace-backed main.dsl diagnostics to be clean, got %+v", published.Diagnostics)
	}
}

func TestDidOpenUsesWorkspaceFilesForSiblingDiagnostics(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	assignmentsText := `assignment R1
Worker -> safety_pref
Supervisor -> authority

assignment R2
Worker -> privacy_pref
SafetyOfficer -> safety_pref
`
	write("main.dsl", `requirement R1
system shall notify Worker
stakeholders Worker, Supervisor

requirement R2
system shall log location of Worker using Database
stakeholders Worker, SafetyOfficer
`)
	write("stakeholders.dsl", "stakeholder Worker\nstakeholder Supervisor\nstakeholder SafetyOfficer\n")
	write("assignments.dsl", assignmentsText)
	write("values/preferences.dsl", "value privacy_pref = 1.58, 0.91\nvalue safety_pref = 0.42, 0.88\n")

	server := NewServer()
	var published protocol.PublishDiagnosticsParams
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			published = params.(protocol.PublishDiagnosticsParams)
		},
	}
	uri := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "assignments.dsl")}).String())

	if err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     uri,
			Version: 1,
			Text:    assignmentsText,
		},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	if len(published.Diagnostics) != 0 {
		t.Fatalf("expected workspace-backed sibling diagnostics to be clean, got %+v", published.Diagnostics)
	}
}

func TestWorkspaceSiblingNavigationUsesWorkspaceSymbolsAndFileLocations(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	equipmentText := `requirement R3
system shall monitor protective_equipment_usage of Worker using Sensor
stakeholders Worker, SafetyOfficer
`
	write("main.dsl", "requirement R1\nsystem shall notify Worker\nstakeholders Worker\n")
	write("stakeholders.dsl", "stakeholder Worker\nstakeholder SafetyOfficer\n")
	write("features/equipment.dsl", equipmentText)

	server := NewServer()
	equipmentURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "features", "equipment.dsl")}).String())
	stakeholdersURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "stakeholders.dsl")}).String())
	if err := server.didOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     equipmentURI,
			Version: 1,
			Text:    equipmentText,
		},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	position := protocol.Position{Line: 1, Character: 53}
	definitionResult, err := server.definition(&glsp.Context{}, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: equipmentURI},
			Position:     position,
		},
	})
	if err != nil {
		t.Fatalf("definition() error = %v", err)
	}
	definitions := definitionResult.([]protocol.Location)
	if len(definitions) != 1 || definitions[0].URI != stakeholdersURI || definitions[0].Range.Start.Line != 0 {
		t.Fatalf("expected Worker definition in stakeholders.dsl, got %+v", definitions)
	}

	hover, err := server.hover(&glsp.Context{}, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: equipmentURI},
			Position:     position,
		},
	})
	if err != nil {
		t.Fatalf("hover() error = %v", err)
	}
	if hover == nil || !strings.Contains(hover.Contents.(protocol.MarkupContent).Value, "**stakeholder** `Worker`") {
		t.Fatalf("expected workspace-backed Worker hover, got %+v", hover)
	}

	tokens, err := server.semanticTokens(&glsp.Context{}, &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: equipmentURI},
	})
	if err != nil {
		t.Fatalf("semanticTokens() error = %v", err)
	}
	if tokens == nil || len(tokens.Data) == 0 {
		t.Fatalf("expected workspace-backed semantic tokens")
	}

	folds, err := server.foldingRange(&glsp.Context{}, &protocol.FoldingRangeParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: equipmentURI},
	})
	if err != nil {
		t.Fatalf("foldingRange() error = %v", err)
	}
	if len(folds) != 1 || folds[0].StartLine != 0 || folds[0].EndLine != 2 {
		t.Fatalf("expected only equipment requirement fold with local line numbers, got %+v", folds)
	}

	symbols, err := server.workspaceSymbol(&glsp.Context{}, &protocol.WorkspaceSymbolParams{Query: "Worker"})
	if err != nil {
		t.Fatalf("workspaceSymbol() error = %v", err)
	}
	if len(symbols) != 1 || symbols[0].Name != "Worker" || symbols[0].Location.URI != stakeholdersURI {
		t.Fatalf("expected Worker workspace symbol in stakeholders.dsl, got %+v", symbols)
	}
}

func TestWorkspaceSiblingAnalysisInvalidatesWhenOpenFileChanges(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	mainText := `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
system shall notify Worker
stakeholders Worker
`
	assignmentsText := `assignment R1
Worker -> privacy_pref
`
	write("main.dsl", mainText)
	write("assignments.dsl", assignmentsText)

	server := NewServer()
	server.debounce = time.Hour
	defer server.cancelAllScheduled()

	mainURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "main.dsl")}).String())
	assignmentsURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "assignments.dsl")}).String())

	if err := server.didOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     mainURI,
			Version: 1,
			Text:    mainText,
		},
	}); err != nil {
		t.Fatalf("didOpen(main) error = %v", err)
	}
	if err := server.didOpen(&glsp.Context{}, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     assignmentsURI,
			Version: 1,
			Text:    assignmentsText,
		},
	}); err != nil {
		t.Fatalf("didOpen(assignments) error = %v", err)
	}

	_, initialResult, ok := server.currentDocumentAnalysis(assignmentsURI)
	if !ok {
		t.Fatal("expected open assignments document")
	}
	if diagnostics := initialResult.AllDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("expected clean initial assignments diagnostics, got %+v", diagnostics)
	}

	changedMainText := strings.Replace(mainText, "requirement R1", "requirement R7", 1)
	if err := server.didChange(&glsp.Context{}, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: mainURI},
			Version:                2,
		},
		ContentChanges: []any{protocol.TextDocumentContentChangeEventWhole{Text: changedMainText}},
	}); err != nil {
		t.Fatalf("didChange(main) error = %v", err)
	}

	definitionResult, err := server.definition(&glsp.Context{}, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: assignmentsURI},
			Position:     protocol.Position{Line: 0, Character: 11},
		},
	})
	if err != nil {
		t.Fatalf("definition() error = %v", err)
	}
	if definitions := definitionResult.([]protocol.Location); len(definitions) != 0 {
		t.Fatalf("expected stale R1 definition to disappear after main.dsl edit, got %+v", definitions)
	}

	_, changedResult, ok := server.currentDocumentAnalysis(assignmentsURI)
	if !ok {
		t.Fatal("expected open assignments document after main change")
	}
	foundUnknownRequirement := false
	for _, diagnostic := range changedResult.AllDiagnostics() {
		if diagnostic.Code == validation.CodeAssignmentRequirementUnknown {
			foundUnknownRequirement = true
			break
		}
	}
	if !foundUnknownRequirement {
		t.Fatalf("expected assignment unknown requirement diagnostic after main.dsl edit, got %+v", changedResult.AllDiagnostics())
	}
}

func TestWorkspaceDiagnosticsPublishClosedSiblingErrors(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	mainText := `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
system shall notify Worker
stakeholders Worker
`
	write("main.dsl", mainText)
	write("assignments.dsl", "assignment R1\nWorker -> privacy_pref\n")

	server := NewServer()
	server.debounce = 10 * time.Millisecond
	defer server.cancelAllScheduled()

	mainURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "main.dsl")}).String())
	assignmentsURI := protocol.DocumentUri((&url.URL{Scheme: "file", Path: filepath.Join(root, "assignments.dsl")}).String())
	var mu sync.Mutex
	published := map[protocol.DocumentUri]protocol.PublishDiagnosticsParams{}
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			diagnostics := params.(protocol.PublishDiagnosticsParams)
			mu.Lock()
			published[diagnostics.URI] = diagnostics
			mu.Unlock()
		},
	}

	if err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: mainURI, Version: 1, Text: mainText},
	}); err != nil {
		t.Fatalf("didOpen(main) error = %v", err)
	}
	changedMainText := strings.Replace(mainText, "requirement R1", "requirement R7", 1)
	if err := server.didChange(context, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: mainURI},
			Version:                2,
		},
		ContentChanges: []any{protocol.TextDocumentContentChangeEventWhole{Text: changedMainText}},
	}); err != nil {
		t.Fatalf("didChange(main) error = %v", err)
	}
	time.Sleep(40 * time.Millisecond)

	mu.Lock()
	assignmentsDiagnostics, ok := published[assignmentsURI]
	mu.Unlock()
	if !ok {
		t.Fatalf("expected diagnostics publish for closed assignments.dsl, got %+v", published)
	}
	if assignmentsDiagnostics.Version != nil {
		t.Fatalf("expected closed-file diagnostics without version, got %+v", assignmentsDiagnostics.Version)
	}
	if len(assignmentsDiagnostics.Diagnostics) != 1 || assignmentsDiagnostics.Diagnostics[0].Code == nil ||
		assignmentsDiagnostics.Diagnostics[0].Code.Value != string(validation.CodeAssignmentRequirementUnknown) {
		t.Fatalf("expected closed assignments unknown requirement diagnostic, got %+v", assignmentsDiagnostics.Diagnostics)
	}
}

func TestWorkspaceSymbolReturnsOpenDocumentDeclarations(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
value privacy_pref = 1.58, 0.91

requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	symbols, err := server.workspaceSymbol(&glsp.Context{}, &protocol.WorkspaceSymbolParams{Query: ""})
	if err != nil {
		t.Fatalf("workspaceSymbol() error = %v", err)
	}
	if len(symbols) != 3 {
		t.Fatalf("expected stakeholder, value, and requirement symbols, got %+v", symbols)
	}

	filtered, err := server.workspaceSymbol(&glsp.Context{}, &protocol.WorkspaceSymbolParams{Query: "privacy"})
	if err != nil {
		t.Fatalf("workspaceSymbol(query) error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "privacy_pref" || filtered[0].Kind != protocol.SymbolKindConstant {
		t.Fatalf("expected filtered value symbol, got %+v", filtered)
	}
}

func TestFoldingRangeReturnsDeclarationGroupsAndBlocks(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91
value safety_pref = 0.42, 0.88

requirement R1
when Worker enters Zone
system shall notify Worker
stakeholders Worker

assignment R1
Worker -> privacy_pref
Manager -> safety_pref
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	folds, err := server.foldingRange(&glsp.Context{}, &protocol.FoldingRangeParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
	})
	if err != nil {
		t.Fatalf("foldingRange() error = %v", err)
	}

	want := [][2]protocol.UInteger{
		{0, 1},
		{3, 4},
		{6, 9},
		{11, 13},
	}
	if len(folds) != len(want) {
		t.Fatalf("expected folds %v, got %+v", want, folds)
	}
	for i, fold := range folds {
		if fold.StartLine != want[i][0] || fold.EndLine != want[i][1] {
			t.Fatalf("fold %d = %+v, want start=%d end=%d", i, fold, want[i][0], want[i][1])
		}
	}
}

func TestSelectionRangeExpandsTokenToFoldBlock(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91
value safety_pref = 0.42, 0.88

requirement R1
when Worker enters Zone
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	ranges, err := server.selectionRange(&glsp.Context{}, &protocol.SelectionRangeParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
		Positions: []protocol.Position{
			{Line: 8, Character: 20},
			{Line: 0, Character: 14},
		},
	})
	if err != nil {
		t.Fatalf("selectionRange() error = %v", err)
	}
	if len(ranges) != 2 {
		t.Fatalf("expected one selection range per position, got %+v", ranges)
	}
	if ranges[0].Range.Start.Line != 8 || ranges[0].Parent == nil || ranges[0].Parent.Range.Start.Line != 6 || ranges[0].Parent.Range.End.Line != 9 {
		t.Fatalf("expected token selection with requirement block parent, got %+v", ranges[0])
	}
	if ranges[1].Range.Start.Line != 0 || ranges[1].Parent == nil || ranges[1].Parent.Range.Start.Line != 0 || ranges[1].Parent.Range.End.Line != 1 {
		t.Fatalf("expected stakeholder token selection with grouped declaration parent, got %+v", ranges[1])
	}
}

func TestCompletionUsesContextAndSymbols(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
stakeholder Manager

value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall notify Worker
stakeholders Worker, Manager
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	resultValue, err := server.complete(&glsp.Context{}, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 8, Character: 13},
		},
	})
	if err != nil {
		t.Fatalf("complete() error = %v", err)
	}

	result, ok := resultValue.(protocol.CompletionList)
	if !ok {
		t.Fatalf("unexpected completion result type: %T", resultValue)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected completion items")
	}

	foundWorker := false
	for _, item := range result.Items {
		if item.Label == "Worker" {
			foundWorker = true
			break
		}
	}
	if !foundWorker {
		t.Fatalf("expected Worker completion, got %+v", result.Items)
	}
}

func TestHoverReturnsKeywordDocumentation(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	resultValue, err := server.hover(&glsp.Context{}, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 3, Character: 1},
		},
	})
	if err != nil {
		t.Fatalf("hover() error = %v", err)
	}

	result := resultValue
	if result == nil {
		t.Fatalf("unexpected hover result: %#v", resultValue)
	}

	contents, ok := result.Contents.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("unexpected hover contents: %#v", result.Contents)
	}
	if contents.Kind != protocol.MarkupKindMarkdown || contents.Value == "" {
		t.Fatalf("unexpected hover content: %#v", contents)
	}
	if result.Range == nil || result.Range.Start.Line != 3 {
		t.Fatalf("expected hover range on when token, got %+v", result.Range)
	}
}

func TestHoverReturnsDeclaredValueDocumentation(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

value privacy_pref = 1.58, 0.91

requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	resultValue, err := server.hover(&glsp.Context{}, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 2, Character: 7},
		},
	})
	if err != nil {
		t.Fatalf("hover() error = %v", err)
	}

	result := resultValue
	if result == nil {
		t.Fatalf("unexpected hover result: %#v", resultValue)
	}

	contents, ok := result.Contents.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("unexpected hover contents: %#v", result.Contents)
	}
	if contents.Kind != protocol.MarkupKindMarkdown || contents.Value == "" {
		t.Fatalf("unexpected hover content: %#v", contents)
	}
	if result.Range == nil || result.Range.Start.Line != 2 {
		t.Fatalf("expected hover range on value token, got %+v", result.Range)
	}
}

func TestFormattingReturnsSingleFullDocumentEdit(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
value privacy_pref = 1.58, 0.91
requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)

	edits, err := server.format(&glsp.Context{}, &protocol.DocumentFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
	})
	if err != nil {
		t.Fatalf("format() error = %v", err)
	}
	if len(edits) != 1 {
		t.Fatalf("expected one formatting edit, got %d", len(edits))
	}
	if edits[0].Range.Start.Line != 0 || edits[0].Range.Start.Character != 0 {
		t.Fatalf("unexpected formatting start range: %+v", edits[0].Range)
	}
	if edits[0].NewText == documentText {
		t.Fatal("expected formatting to change document text")
	}
	if !strings.Contains(edits[0].NewText, "value privacy_pref = 1.58, 0.91") {
		t.Fatalf("unexpected formatted text: %q", edits[0].NewText)
	}
}

func TestRangeFormattingReturnsLineBoundedEdit(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
value privacy_pref = 1.5800, 0.9100
value safety_pref = 0.4200, 0.8800
requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)

	edits, err := server.rangeFormat(&glsp.Context{}, &protocol.DocumentRangeFormattingParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
		Range: protocol.Range{
			Start: protocol.Position{Line: 1, Character: 0},
			End:   protocol.Position{Line: 3, Character: 0},
		},
	})
	if err != nil {
		t.Fatalf("rangeFormat() error = %v", err)
	}
	if len(edits) != 1 {
		t.Fatalf("expected one range formatting edit, got %d", len(edits))
	}
	if edits[0].Range.Start.Line != 1 || edits[0].Range.End.Line != 3 {
		t.Fatalf("expected edit to stay within selected lines, got %+v", edits[0].Range)
	}
	if edits[0].NewText != "value privacy_pref = 1.58, 0.91\nvalue safety_pref = 0.42, 0.88\n" {
		t.Fatalf("unexpected range formatted text: %q", edits[0].NewText)
	}
}

func TestSemanticTokensReturnMeaningAwareHighlightData(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker
value privacy_pref = 1.58, 0.91
requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker
assignment R1
Worker -> privacy_pref // inline note
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	result, err := server.semanticTokens(&glsp.Context{}, &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
	})
	if err != nil {
		t.Fatalf("semanticTokens() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected semantic token result")
	}
	if len(result.Data) == 0 {
		t.Fatal("expected semantic token data")
	}
	if len(result.Data)%5 != 0 {
		t.Fatalf("semantic token data length must be divisible by 5, got %d", len(result.Data))
	}
}

func TestDocumentSymbolReturnsSemanticOutline(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

value privacy_pref = 1.58, 0.91

requirement R1
when Worker enters DangerousArea
system shall notify Worker
stakeholders Worker
priority high
linked_to HazardAnalysis
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	resultValue, err := server.documentSymbol(&glsp.Context{}, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
	})
	if err != nil {
		t.Fatalf("documentSymbol() error = %v", err)
	}

	result, ok := resultValue.([]protocol.DocumentSymbol)
	if !ok {
		t.Fatalf("unexpected documentSymbol result type: %T", resultValue)
	}
	if len(result) != 3 {
		t.Fatalf("expected top-level symbols, got %+v", result)
	}
	if result[2].Name != "R1" || len(result[2].Children) == 0 {
		t.Fatalf("expected requirement outline with children, got %+v", result[2])
	}
}

func TestDocumentLinkReturnsTraceabilityURLs(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
linked_to "google.com/search?q=lsp"
linked_to SAFETY-001
linked_to "https://example.com/spec"
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	links, err := server.documentLink(&glsp.Context{}, &protocol.DocumentLinkParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
	})
	if err != nil {
		t.Fatalf("documentLink() error = %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected two document links, got %+v", links)
	}
	if links[0].Target == nil || string(*links[0].Target) != "https://google.com/search?q=lsp" || links[0].Range.Start.Line != 5 || links[0].Range.Start.Character == 0 {
		t.Fatalf("unexpected first document link: %+v", links[0])
	}
	if links[1].Target == nil || string(*links[1].Target) != "https://example.com/spec" || links[1].Range.Start.Line != 7 || links[1].Range.Start.Character == 0 {
		t.Fatalf("unexpected second document link: %+v", links[1])
	}
}

func TestDefinitionReturnsStakeholderDeclaration(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	resultValue, err := server.definition(&glsp.Context{}, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 3, Character: 20},
		},
	})
	if err != nil {
		t.Fatalf("definition() error = %v", err)
	}

	result, ok := resultValue.([]protocol.Location)
	if !ok {
		t.Fatalf("unexpected definition result type: %T", resultValue)
	}
	if len(result) != 1 {
		t.Fatalf("expected one definition, got %+v", result)
	}
	if result[0].Range.Start.Line != 0 || result[0].Range.Start.Character != 12 {
		t.Fatalf("unexpected definition range: %+v", result[0])
	}
}

func TestDeclarationReusesDefinition(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
system shall notify Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	resultValue, err := server.declaration(&glsp.Context{}, &protocol.DeclarationParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 3, Character: 20},
		},
	})
	if err != nil {
		t.Fatalf("declaration() error = %v", err)
	}

	result, ok := resultValue.([]protocol.Location)
	if !ok {
		t.Fatalf("unexpected declaration result type: %T", resultValue)
	}
	if len(result) != 1 || result[0].Range.Start.Line != 0 || result[0].Range.Start.Character != 12 {
		t.Fatalf("expected declaration to reuse definition location, got %+v", result)
	}
}

func TestReferencesReturnsAllMatchingLocations(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	result, err := server.references(&glsp.Context{}, &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 3, Character: 20},
		},
		Context: protocol.ReferenceContext{IncludeDeclaration: true},
	})
	if err != nil {
		t.Fatalf("references() error = %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected declaration plus two references, got %+v", result)
	}
}

func TestDocumentHighlightReturnsDeclarationAndReferences(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	highlights, err := server.documentHighlight(&glsp.Context{}, &protocol.DocumentHighlightParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 3, Character: 20},
		},
	})
	if err != nil {
		t.Fatalf("documentHighlight() error = %v", err)
	}
	if len(highlights) != 3 {
		t.Fatalf("expected declaration plus two references, got %+v", highlights)
	}
	if highlights[0].Range.Start.Line != 0 || highlights[0].Kind == nil || *highlights[0].Kind != protocol.DocumentHighlightKindWrite {
		t.Fatalf("expected write highlight on declaration, got %+v", highlights[0])
	}
	for _, highlight := range highlights[1:] {
		if highlight.Kind == nil || *highlight.Kind != protocol.DocumentHighlightKindRead {
			t.Fatalf("expected read highlight on reference, got %+v", highlight)
		}
	}
}

func TestRenameReturnsWorkspaceEdit(t *testing.T) {
	server := NewServer()
	documentText := `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`

	server.documents.Set("file:///spec.dsl", 1, documentText)
	server.analyzeDocument(server.documents.Set("file:///spec.dsl", 1, documentText))

	edit, err := server.rename(&glsp.Context{}, &protocol.RenameParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 3, Character: 20},
		},
		NewName: "Employee",
	})
	if err != nil {
		t.Fatalf("rename() error = %v", err)
	}
	if edit == nil || len(edit.Changes["file:///spec.dsl"]) != 3 {
		t.Fatalf("expected rename edits, got %+v", edit)
	}
}

func TestDidChangeDebouncesAndPublishesLatestVersionOnly(t *testing.T) {
	server := NewServer()
	server.debounce = 20 * time.Millisecond

	var (
		mu        sync.Mutex
		published []protocol.PublishDiagnosticsParams
	)
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			mu.Lock()
			defer mu.Unlock()
			published = append(published, params.(protocol.PublishDiagnosticsParams))
		},
	}

	if err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///spec.dsl",
			Version: 1,
			Text: `stakeholder Worker
stakeholder Manager

requirement R1
system shall notify Worker
stakeholders Worker
`,
		},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	if err := server.didChange(context, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Version:                2,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEventWhole{Text: `stakeholder Worker

requirement R1
when Worker enters Danger Zone
system shall notify Worker
stakeholders Worker
`},
		},
	}); err != nil {
		t.Fatalf("didChange(version=2) error = %v", err)
	}

	if err := server.didChange(context, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Version:                3,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEventWhole{Text: `stakeholder Worker

requirement R1
when Worker enters DangerousArea
system shall notify Worker
stakeholders Worker
`},
		},
	}); err != nil {
		t.Fatalf("didChange(version=3) error = %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(published) < 2 {
		t.Fatalf("expected open diagnostics and one debounced change publish, got %+v", published)
	}
	last := published[len(published)-1]
	if last.Version == nil || *last.Version != 3 {
		t.Fatalf("expected latest published diagnostics for version 3, got %+v", last)
	}
	if len(last.Diagnostics) != 0 {
		t.Fatalf("expected latest version to be valid, got %+v", last.Diagnostics)
	}
}

func TestDidCloseCancelsPendingScheduledAnalysis(t *testing.T) {
	server := NewServer()
	server.debounce = 40 * time.Millisecond

	var (
		mu        sync.Mutex
		published []protocol.PublishDiagnosticsParams
	)
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			mu.Lock()
			defer mu.Unlock()
			published = append(published, params.(protocol.PublishDiagnosticsParams))
		},
	}

	if err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///spec.dsl",
			Version: 1,
			Text: `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`,
		},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	if err := server.didChange(context, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Version:                2,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEventWhole{Text: `stakeholder Worker

requirement R1
when Worker enters Danger Zone
system shall notify Worker
stakeholders Worker
`},
		},
	}); err != nil {
		t.Fatalf("didChange() error = %v", err)
	}

	if err := server.didClose(context, &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
	}); err != nil {
		t.Fatalf("didClose() error = %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	last := published[len(published)-1]
	if last.Version == nil || *last.Version != 0 {
		t.Fatalf("expected close diagnostics publish to be last, got %+v", last)
	}
}

func TestNavigationFallbackIsParseOnlyAndDiagnosticsStayCurrent(t *testing.T) {
	server := NewServer()
	server.debounce = 20 * time.Millisecond

	var (
		mu        sync.Mutex
		published []protocol.PublishDiagnosticsParams
	)
	context := &glsp.Context{
		Notify: func(_ string, params any) {
			mu.Lock()
			defer mu.Unlock()
			published = append(published, params.(protocol.PublishDiagnosticsParams))
		},
	}

	if err := server.didOpen(context, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///spec.dsl",
			Version: 1,
			Text: `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders Worker
`,
		},
	}); err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	if err := server.didChange(context, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Version:                2,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEventWhole{Text: `stakeholder Worker

requirement R1
system shall notify Worker
stakeholders U
`},
		},
	}); err != nil {
		t.Fatalf("didChange(version=2) error = %v", err)
	}
	time.Sleep(40 * time.Millisecond)

	resultValue, err := server.complete(context, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 4, Character: 14},
		},
	})
	if err != nil {
		t.Fatalf("complete(version=2) error = %v", err)
	}
	items := resultValue.(protocol.CompletionList).Items
	for _, item := range items {
		if item.Label == "Manager" {
			t.Fatalf("did not expect semantic-error fallback completion, got %+v", items)
		}
	}

	if err := server.didChange(context, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Version:                3,
		},
		ContentChanges: []any{
			protocol.TextDocumentContentChangeEventWhole{Text: "requirement R1\nsystem shall notify Worker\nstakeholders \n"},
		},
	}); err != nil {
		t.Fatalf("didChange(version=3) error = %v", err)
	}
	time.Sleep(40 * time.Millisecond)

	resultValue, err = server.hover(context, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///spec.dsl"},
			Position:     protocol.Position{Line: 1, Character: 20},
		},
	})
	if err != nil {
		t.Fatalf("hover(version=3) error = %v", err)
	}
	hover, ok := resultValue.(*protocol.Hover)
	if !ok {
		t.Fatalf("unexpected hover result type: %T", resultValue)
	}
	if hover == nil {
		t.Fatal("expected parse-fallback hover")
	}
	contents := hover.Contents.(protocol.MarkupContent)
	if !strings.Contains(contents.Value, "**stakeholder** `Worker`") {
		t.Fatalf("expected parse-fallback stakeholder hover, got %+v", hover)
	}

	mu.Lock()
	defer mu.Unlock()
	last := published[len(published)-1]
	if last.Version == nil || *last.Version != 3 {
		t.Fatalf("expected latest diagnostics for version 3, got %+v", last)
	}
	if len(last.Diagnostics) == 0 {
		t.Fatalf("expected current parse diagnostics for version 3, got %+v", last)
	}
}

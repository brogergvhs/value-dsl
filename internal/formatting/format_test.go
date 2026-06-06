package formatting

import (
	"errors"
	"strings"
	"testing"
)

func TestFormatCanonicalizesDocumentLayout(t *testing.T) {
	input := `stakeholder Worker
stakeholder Supervisor
value privacy_pref = 1.5, .91
value safety_pref = 4.42, 0.93
requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker,Supervisor
priority high
linked_to HazardAnalysis
assignment R1
Worker -> privacy_pref
Supervisor -> safety_pref
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	expected := `stakeholder Worker
stakeholder Supervisor

value privacy_pref = 1.5, 0.91
value safety_pref = 4.42, 0.93

requirement R1
when Worker enters DangerousArea
system shall track location of Worker using Camera
stakeholders Worker, Supervisor
priority high
linked_to HazardAnalysis

assignment R1
Worker -> privacy_pref
Supervisor -> safety_pref
`

	if formatted != expected {
		t.Fatalf("Format() mismatch:\nexpected:\n%s\ngot:\n%s", expected, formatted)
	}
}

func TestFormatSeparatesStakeholderAndValueBlocks(t *testing.T) {
	input := `stakeholder Supervisor
stakeholder Worker
stakeholder SafetyOfficer
stakeholder TestStake
value privacy_pref = 1.17, 0.91
value authority_pref = 1.00, 0.20
value test = 1.00, 0.60
value safety_pref = 4.42,
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	expected := `stakeholder Supervisor
stakeholder Worker
stakeholder SafetyOfficer
stakeholder TestStake

value privacy_pref = 1.17, 0.91
value authority_pref = 1, 0.2
value test = 1, 0.6
value safety_pref = 4.42,
`

	if formatted != expected {
		t.Fatalf("Format() mismatch:\nexpected:\n%s\ngot:\n%s", expected, formatted)
	}
}

func TestFormatNormalizesStringQuotes(t *testing.T) {
	input := `stakeholder Worker
requirement R1
system shall notify Worker using "camera feed"
stakeholders Worker
retention "30 days"
linked_to 'Safety Case'
linked_to "google.com/search?q=lsp"
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	expected := `stakeholder Worker

requirement R1
system shall notify Worker using 'camera feed'
stakeholders Worker
retention '30 days'
linked_to 'Safety Case'
linked_to 'google.com/search?q=lsp'
`

	if formatted != expected {
		t.Fatalf("Format() mismatch:\nexpected:\n%s\ngot:\n%s", expected, formatted)
	}
}

func TestFormatPreservesAndNormalizesComments(t *testing.T) {
	input := `// actor declaration
stakeholder Worker   // keep worker
value privacy_pref = 1.58, 0.91
// important requirement
requirement R1 // this is an important requirement
when Worker enters DangerousArea
system shall notify Worker // alert the actor
stakeholders Worker
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	expected := `// actor declaration
stakeholder Worker // keep worker

value privacy_pref = 1.58, 0.91

// important requirement
requirement R1 // this is an important requirement
when Worker enters DangerousArea
system shall notify Worker // alert the actor
stakeholders Worker
`

	if formatted != expected {
		t.Fatalf("Format() mismatch:\nexpected:\n%s\ngot:\n%s", expected, formatted)
	}
}

func TestFormatPreservesAssignmentEntryWithoutArrow(t *testing.T) {
	input := `stakeholder Worker
value privacy_pref = 1.58, 0.91
requirement R1
system shall notify Worker
stakeholders Worker
assignment R1
Worker privacy_pref
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(formatted, "Worker privacy_pref\n") {
		t.Fatalf("expected missing-arrow assignment entry to be preserved, got:\n%s", formatted)
	}
	if strings.Contains(formatted, "Worker privacy_pref ->") {
		t.Fatalf("formatter must not add an assignment arrow to invalid input, got:\n%s", formatted)
	}
}

func TestFormatEchoesMalformedInput(t *testing.T) {
	input := `mystery top
stakeholder Worker
requirement R1
unknown clause
system shall notify Worker
stakeholders Worker
assignment R1
unknown assignment line
Worker -> privacy
`

	formatted, err := Format(input)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	expected := `mystery top

stakeholder Worker

requirement R1
unknown clause
system shall notify Worker
stakeholders Worker

assignment R1
unknown assignment line
Worker -> privacy
`
	if formatted != expected {
		t.Fatalf("expected malformed lines to be preserved in formatted output:\n%s", formatted)
	}
}

func TestFormatForWriteRejectsMalformedInput(t *testing.T) {
	input := `stakeholder Worker
requirement R1
unknown clause
system shall notify Worker
stakeholders Worker
`

	_, err := FormatForWrite(input)
	if err == nil {
		t.Fatal("expected FormatForWrite to reject malformed input")
	}

	var unsafeErr *unsafeWriteError
	if !errors.As(err, &unsafeErr) {
		t.Fatalf("expected UnsafeWriteError, got %T: %v", err, err)
	}
	if unsafeErr.ParseDiagnostics == 0 || unsafeErr.MalformedLines == 0 {
		t.Fatalf("expected parser diagnostics and malformed line counts, got %+v", unsafeErr)
	}
}

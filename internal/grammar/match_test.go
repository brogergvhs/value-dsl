package grammar

import "testing"

func TestTokenizeStringWithEscapedBackslashBeforeQuote(t *testing.T) {
	line := TokenizeRawLine(1, `retention "foo\\" trailing`)
	if len(line.Tokens) != 3 {
		t.Fatalf("tokens = %+v, want 3 tokens", line.Tokens)
	}
	if line.Tokens[1].Text != `"foo\\"` {
		t.Fatalf("string token = %q", line.Tokens[1].Text)
	}
	if line.Tokens[2].Text != "trailing" {
		t.Fatalf("trailing token = %q", line.Tokens[2].Text)
	}
}

func TestTokenizeStringKeepsEscapedQuoteInsideString(t *testing.T) {
	line := TokenizeRawLine(1, `retention "foo\"bar"`)
	if len(line.Tokens) != 2 {
		t.Fatalf("tokens = %+v, want 2 tokens", line.Tokens)
	}
	if line.Tokens[1].Text != `"foo\"bar"` {
		t.Fatalf("string token = %q", line.Tokens[1].Text)
	}
}

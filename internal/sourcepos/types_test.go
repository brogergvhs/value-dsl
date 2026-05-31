package sourcepos

import "testing"

func TestUTF16ConversionsRejectInvalidUTF8(t *testing.T) {
	invalid := string([]byte{'a', 0xff, 'b'})

	if _, ok := ByteOffsetFromUTF16OK(invalid, 2); ok {
		t.Fatal("expected invalid UTF-8 in ByteOffsetFromUTF16OK")
	}
	if _, ok := UTF16OffsetFromByteOK(invalid, len(invalid)); ok {
		t.Fatal("expected invalid UTF-8 in UTF16OffsetFromByteOK")
	}
}

func TestUTF16ConversionsHandleSurrogatePairs(t *testing.T) {
	line := "a😀b"

	offset, ok := ByteOffsetFromUTF16OK(line, 3)
	if !ok {
		t.Fatal("expected valid UTF-8")
	}
	if offset != len("a😀") {
		t.Fatalf("offset = %d, want %d", offset, len("a😀"))
	}

	units, ok := UTF16OffsetFromByteOK(line, len("a😀"))
	if !ok {
		t.Fatal("expected valid UTF-8")
	}
	if units != 3 {
		t.Fatalf("units = %d, want 3", units)
	}
}

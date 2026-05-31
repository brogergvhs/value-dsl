// Package sourcepos defines shared source positions and ranges for the DSL.
package sourcepos

import (
	"unicode/utf16"
	"unicode/utf8"
)

type Position struct {
	Line int
	// Column is a 1-based byte column in the original document text.
	Column int
}

type Range struct {
	Start Position
	End   Position
}

func NewPosition(line, column int) Position { return Position{Line: line, Column: column} }
func NewRange(start, end Position) Range    { return Range{Start: start, End: end} }

func LineRange(line int) Range {
	p := NewPosition(line, 1)
	return NewRange(p, p)
}

func ByteOffsetFromUTF16OK(line string, character int) (int, bool) {
	if character <= 0 {
		return 0, utf8.ValidString(line)
	}

	offset, units := 0, 0
	for offset < len(line) && units < character {
		r, size := utf8.DecodeRuneInString(line[offset:])
		if r == utf8.RuneError && size == 1 {
			return offset, false
		}
		step := utf16.RuneLen(r)
		if step < 0 {
			return offset, false
		}
		if units+step > character {
			break
		}
		units += step
		offset += size
	}

	return offset, true
}

func UTF16OffsetFromByteOK(line string, offset int) (int, bool) {
	if offset <= 0 {
		return 0, utf8.ValidString(line)
	}
	if offset > len(line) {
		offset = len(line)
	}

	units := 0
	for i := 0; i < offset; {
		r, size := utf8.DecodeRuneInString(line[i:])
		if r == utf8.RuneError && size == 1 {
			return units, false
		}
		step := utf16.RuneLen(r)
		if step < 0 {
			return units, false
		}
		units += step
		i += size
	}

	return units, true
}

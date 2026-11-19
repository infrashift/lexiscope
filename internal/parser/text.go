package parser

import (
	"regexp"

	"github.com/infrashift/lexiscope/internal/assert"
)

// Text is prose with the source offset of each byte. Off[i] is the offset
// in the file of the source byte that gave B[i]. A replacement maps each of
// its bytes to the start of the text that it replaced.
type Text struct {
	B   []byte
	Off []int32
}

// add appends s. With fill, each byte of s maps to off. Without fill, byte k
// of s maps to off+k.
func (t *Text) add(s string, off int32, fill bool) {
	assert.True(off >= 0, "an offset is not negative")
	assert.True(len(t.B) == len(t.Off), "each byte has an offset")
	for k := range len(s) {
		t.B = append(t.B, s[k])
		if fill {
			t.Off = append(t.Off, off)
		} else {
			t.Off = append(t.Off, off+int32(k))
		}
	}
}

// String returns the bytes of the text.
func (t Text) String() string {
	assert.True(len(t.B) == len(t.Off), "each byte has an offset")
	assert.True(len(t.B) <= MaxBlockBytes, "a block is bounded")
	return string(t.B)
}

// Sub returns the bytes i to j of the text.
func (t Text) Sub(i, j int) Text {
	assert.True(0 <= i && i <= j && j <= len(t.B), "the range is in the text")
	assert.True(len(t.B) == len(t.Off), "each byte has an offset")
	return Text{B: t.B[i:j], Off: t.Off[i:j]}
}

// Start returns the source offset of byte i, or of the end for i == len.
func (t Text) Start(i int) int32 {
	assert.True(len(t.B) > 0, "the text is not empty")
	assert.True(0 <= i && i <= len(t.B), "the index is in the text")
	if i == len(t.B) {
		return t.Off[i-1] + 1
	}
	return t.Off[i]
}

// End returns the source offset after byte j-1 of the text.
func (t Text) End(j int) int32 {
	assert.True(0 < j && j <= len(t.B), "the end follows the first byte")
	assert.True(len(t.B) == len(t.Off), "each byte has an offset")
	return t.Off[j-1] + 1
}

// ReplaceAll replaces each match of re, left to right, as Python's re.sub
// does for a pattern that can not match the empty string.
func (t Text) ReplaceAll(re *regexp.Regexp, repl string) Text {
	assert.True(re != nil, "the pattern exists")
	assert.True(len(t.B) == len(t.Off), "each byte has an offset")
	var out Text
	last := 0
	for _, m := range re.FindAllIndex(t.B, -1) {
		assert.True(m[1] > m[0], "a cleanup pattern does not match the empty string")
		out.B = append(out.B, t.B[last:m[0]]...)
		out.Off = append(out.Off, t.Off[last:m[0]]...)
		out.add(repl, t.Off[m[0]], true)
		last = m[1]
	}
	out.B = append(out.B, t.B[last:]...)
	out.Off = append(out.Off, t.Off[last:]...)
	return out
}

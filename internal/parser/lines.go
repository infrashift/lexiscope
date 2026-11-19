package parser

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/infrashift/lexiscope/internal/assert"
)

// Lines is the line table of one file. A line ends at "\r\n", "\r" or "\n",
// the universal newlines of Python's read_text.
type Lines struct {
	src    []byte
	starts []int32 // the offset of the first byte of each line
	ends   []int32 // the offset after the last byte of each line, without the newline
}

// NewLines builds the line table of src.
func NewLines(src []byte) (*Lines, error) {
	assert.True(src != nil || len(src) == 0, "the source exists")
	if len(src) > MaxFileBytes {
		return nil, fmt.Errorf("%d bytes, more than %d", len(src), MaxFileBytes)
	}
	l := &Lines{src: src, starts: []int32{0}}
	for i := 0; i < len(src); i++ {
		if src[i] != '\n' && src[i] != '\r' {
			continue
		}
		l.ends = append(l.ends, int32(i))
		if src[i] == '\r' && i+1 < len(src) && src[i+1] == '\n' {
			i++
		}
		l.starts = append(l.starts, int32(i+1))
	}
	l.ends = append(l.ends, int32(len(src)))
	if len(l.starts) > MaxLines {
		return nil, fmt.Errorf("%d lines, more than %d", len(l.starts), MaxLines)
	}
	assert.True(len(l.starts) == len(l.ends), "each line has a start and an end")
	return l, nil
}

// Count returns the number of lines.
func (l *Lines) Count() int {
	assert.True(l != nil, "the table exists")
	assert.True(len(l.starts) > 0, "a file has one line or more")
	return len(l.starts)
}

// Text returns line i (0-based) without its newline.
func (l *Lines) Text(i int) string {
	assert.True(0 <= i && i < len(l.starts), "the line exists")
	assert.True(l.starts[i] <= l.ends[i], "a line does not end before it starts")
	return string(l.src[l.starts[i]:l.ends[i]])
}

// Start returns the offset of the first byte of line i (0-based).
func (l *Lines) Start(i int) int32 {
	assert.True(0 <= i && i < len(l.starts), "the line exists")
	assert.True(l.starts[i] >= 0, "an offset is not negative")
	return l.starts[i]
}

// LineOf returns the 1-based line of off.
func (l *Lines) LineOf(off int32) int {
	assert.True(off >= 0 && int(off) <= len(l.src), "the offset is in the file")
	assert.True(len(l.starts) > 0, "a file has one line or more")
	i, found := slices.BinarySearch(l.starts, off)
	if !found {
		i-- // off is inside line i-1.
	}
	return i + 1
}

// Position returns the 1-based line and the 1-based UTF-16 column of off,
// the column unit of SARIF.
func (l *Lines) Position(off int32) (int, int) {
	assert.True(off >= 0 && int(off) <= len(l.src), "the offset is in the file")
	assert.True(len(l.starts) > 0, "a file has one line or more")
	line := l.LineOf(off)
	col := 1
	rest := l.src[l.starts[line-1]:off]
	for len(rest) > 0 {
		r, size := utf8.DecodeRune(rest)
		col += utf16Len(r)
		rest = rest[size:]
	}
	return line, col
}

func utf16Len(r rune) int {
	assert.True(r >= 0, "a rune is not negative")
	assert.True(r <= utf8.MaxRune || r == utf8.RuneError, "a rune is in range")
	if r >= 0x10000 {
		return 2
	}
	return 1
}

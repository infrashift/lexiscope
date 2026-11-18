// Package parser extracts prose blocks from Markdown files and from the
// comments of source files.
//
// For Markdown, the extractor copies the line rules of scripts/ste-check.py,
// the oracle of the core rules (ADR-0005). It skips the same lines, and it
// makes the same paragraphs and the same text cleanup. The wiki spec "Known differences
// from ste-check" lists each deliberate difference.
//
// Each byte of a block keeps the offset of its source byte, so a finding
// can name the exact line and column. The extractor uses no recursion. Each
// loop runs over a slice or counts toward a limit.
package parser

import (
	"unicode"
	"unicode/utf8"

	"github.com/infrashift/lexis/internal/assert"
)

// Limits of the extractor.
const (
	MaxFileBytes  = 16 << 20 // the source limit of ztree
	MaxBlockBytes = 1 << 20
	MaxLines      = 1 << 20
	MaxBlocks     = 1 << 18
)

// PySpaceClass is the regexp class of Python's \s for str: the characters
// of str.isspace.
const PySpaceClass = `\t\n\v\f\r \x{1c}-\x{1f}\x{85}\p{Z}`

// IsSpace reports whether r is white space for Python's str.isspace, the
// definition of the oracle.
func IsSpace(r rune) bool {
	assert.True(r >= 0, "a rune is not negative")
	if r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' || r == ' ' || r == 0x85 {
		return true
	}
	assert.True(r != ' ', "the ASCII space returns above")
	return (r >= 0x1c && r <= 0x1f) || unicode.In(r, unicode.Z)
}

// trimSpace removes the white space of Python's str.strip from both ends.
func trimSpace(s string) string {
	assert.True(len(s) <= MaxFileBytes, "the text is bounded")
	start, end := 0, len(s)
	for start < end {
		r, size := utf8.DecodeRuneInString(s[start:end])
		if !IsSpace(r) {
			break
		}
		start += size
	}
	for end > start {
		r, size := utf8.DecodeLastRuneInString(s[start:end])
		if !IsSpace(r) {
			break
		}
		end -= size
	}
	assert.True(start <= end, "the ends do not cross")
	return s[start:end]
}

// IsWord reports whether r is a word character for Python's \w on str: a
// letter, a number or an underscore.
func IsWord(r rune) bool {
	assert.True(r >= 0, "a rune is not negative")
	w := r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
	assert.True(!w || !IsSpace(r), "a word character is not white space")
	return w
}

package engine

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/infrashift/lexiscope/internal/assert"
	"github.com/infrashift/lexiscope/internal/parser"
)

// span is the byte range [start, end) of a block text.
type span struct{ start, end int }

// sentences splits s as the oracle does: at a run of white space after
// '.', '!' or '?' that a non-space character follows.
func sentences(s string) []span {
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	var out []span
	start := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != '.' && r != '!' && r != '?' {
			i += size
			continue
		}
		j := spaceRun(s, i+size)
		if j > i+size && j < len(s) {
			out = append(out, span{start, i + size})
			start = j
		}
		i = max(j, i+size)
	}
	out = append(out, span{start, len(s)})
	assert.True(len(out) > 0, "a text has one sentence or more")
	return out
}

// spaceRun returns the end of the run of white space that starts at i.
func spaceRun(s string, i int) int {
	assert.True(0 <= i && i <= len(s), "the index is in the text")
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !parser.IsSpace(r) {
			break
		}
		i += size
	}
	return i
}

// wordCount counts the words of a sentence as the oracle does: the parts
// between runs of white space that hold an ASCII letter or digit.
func wordCount(s string) int {
	assert.True(len(s) <= parser.MaxBlockBytes, "a sentence is bounded")
	n := 0
	counted := false // the current part has an ASCII letter or digit, and n counts it
	for _, r := range s {
		switch {
		case parser.IsSpace(r):
			counted = false
		case !counted && r < utf8.RuneSelf && (unicode.IsLetter(r) || unicode.IsDigit(r)):
			n++
			counted = true
		}
	}
	assert.True(n <= len(s), "each word has a byte or more")
	return n
}

// match is one match of a term in a text.
type match struct {
	span
	text string // the matched text, with its case
	term string // the term of the set, in lower case
}

// boundary reports whether a word boundary of Python's \b separates the
// runes before and after byte i of s.
func boundary(s string, i int) bool {
	assert.True(0 <= i && i <= len(s), "the index is in the text")
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	before, after := false, false
	if i > 0 {
		before = parser.IsWord(lastRune(s[:i]))
	}
	if i < len(s) {
		after = parser.IsWord(firstRune(s[i:]))
	}
	return before != after
}

func firstRune(s string) rune {
	assert.True(s != "", "the text is not empty")
	r, size := utf8.DecodeRuneInString(s)
	assert.True(size > 0, "a rune has a byte or more")
	return r
}

func lastRune(s string) rune {
	assert.True(s != "", "the text is not empty")
	r, size := utf8.DecodeLastRuneInString(s)
	assert.True(size > 0, "a rune has a byte or more")
	return r
}

// runeSize returns the byte length of the first rune of s.
func runeSize(s string) int {
	assert.True(s != "", "the text is not empty")
	r, size := utf8.DecodeRuneInString(s)
	assert.True(size > 0 && (r != utf8.RuneError || size == 1), "a rune has a byte or more")
	return size
}

// matchTerms finds each term of the set in s: left to right, without
// overlap, without case and between word boundaries. The oracle's
// \b(term|...)\b with re.I does the same. A longer term wins at one position.
func matchTerms(s string, terms []string) []match {
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	assert.True(len(terms) <= 1<<16, "the term count is bounded")
	var out []match
	for i := 0; i < len(s); {
		m, ok := matchAt(s, i, terms)
		if ok {
			out = append(out, m)
			i = m.end
			continue
		}
		i += runeSize(s[i:])
	}
	return out
}

// matchAt returns the first term that matches at byte i of s.
func matchAt(s string, i int, terms []string) (match, bool) {
	assert.True(0 <= i && i < len(s), "the index is in the text")
	assert.True(len(terms) <= 1<<16, "the term count is bounded")
	if !boundary(s, i) {
		return match{}, false
	}
	for _, t := range terms {
		n := prefixFold(s[i:], t)
		if n > 0 && boundary(s, i+n) {
			return match{span: span{i, i + n}, text: s[i : i+n], term: t}, true
		}
	}
	return match{}, false
}

// prefixFold returns the byte length of the prefix of s that equals term
// under simple case folding, or 0.
func prefixFold(s, term string) int {
	assert.True(term != "", "a term is not empty")
	assert.True(len(term) < 1<<16, "a term is bounded")
	i := 0
	for _, tr := range term {
		if i >= len(s) {
			return 0
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != tr && !strings.EqualFold(string(r), string(tr)) {
			return 0
		}
		i += size
	}
	return i
}

// longestFirst orders terms by length, longest first, then by text.
func longestFirst(a, b string) int {
	assert.True(a != "" && b != "", "terms are not empty")
	assert.True(len(a) < 1<<16 && len(b) < 1<<16, "terms are bounded")
	if len(a) != len(b) {
		return len(b) - len(a)
	}
	return strings.Compare(a, b)
}

// words returns the runs of word characters of s, with their spans.
func words(s string) []span {
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	var out []span
	start := -1
	for i, r := range s {
		w := parser.IsWord(r)
		if w && start < 0 {
			start = i
		} else if !w && start >= 0 {
			out = append(out, span{start, i})
			start = -1
		}
	}
	if start >= 0 {
		out = append(out, span{start, len(s)})
	}
	assert.True(len(out) <= len(s), "each word has a byte or more")
	return out
}

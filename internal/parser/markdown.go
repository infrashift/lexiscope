package parser

import (
	"bytes"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/infrashift/lexis/internal/assert"
)

// okfwMarker starts the generated okfw block of CLAUDE.md. The oracle does
// not check CLAUDE.md after it.
const okfwMarker = "<!-- BEGIN okfw -->"

// skipPrefixes are the prefixes of the lines that the oracle skips: tables,
// HTML and JSX, MDX imports and exports, and JSX expressions.
var skipPrefixes = []string{"|", "<", "import ", "export ", "/>", "{"}

var headingRE = regexp.MustCompile(`^#{1,6}([` + PySpaceClass + `]|$)`)

// mdState is the state of the line scan of one Markdown file.
type mdState struct {
	lines    *Lines
	fence    byte // '`' or '~' inside a fenced code block, else 0
	comment  bool // inside a multi-line HTML comment
	prose    []proseLine
	headings []Block
}

// Markdown returns the blocks of a Markdown or MDX file. name selects the
// CLAUDE.md rule of the oracle.
func Markdown(name string, src []byte) ([]Block, error) {
	assert.True(name != "", "the file has a name")
	if !utf8.Valid(src) {
		return nil, errors.New("the file is not valid UTF-8")
	}
	if filepath.Base(name) == "CLAUDE.md" {
		if i := bytes.Index(src, []byte(okfwMarker)); i >= 0 {
			src = src[:i]
		}
	}
	lines, err := NewLines(src)
	if err != nil {
		return nil, err
	}
	st := &mdState{lines: lines}
	for i := frontMatterEnd(lines); i < lines.Count(); i++ {
		st.scan(i)
	}
	blocks, err := paragraphs(st.prose, false)
	if err != nil {
		return nil, err
	}
	assert.True(len(st.headings) <= MaxLines, "the heading count is bounded")
	return append(blocks, st.headings...), nil
}

// frontMatterEnd returns the first line after the front matter. The oracle
// knows YAML front matter (---). lexis also knows TOML front matter (+++).
// Without a line that closes it, the front matter runs to the end of the file.
func frontMatterEnd(lines *Lines) int {
	assert.True(lines != nil, "the table exists")
	assert.True(lines.Count() > 0, "a file has one line or more")
	first := trimSpace(lines.Text(0))
	if first != "---" && first != "+++" {
		return 0
	}
	for i := 1; i < lines.Count(); i++ {
		if trimSpace(lines.Text(i)) == first {
			return i + 1
		}
	}
	return lines.Count()
}

// scan reads line i (0-based).
func (st *mdState) scan(i int) {
	assert.True(st.lines != nil, "the table exists")
	assert.True(0 <= i && i < st.lines.Count(), "the line exists")
	raw := st.lines.Text(i)
	s := trimSpace(raw)
	if st.skipFenceOrComment(s) || s == "" {
		return
	}
	off := st.lines.Start(i) + int32(strings.Index(raw, s))
	if strings.HasPrefix(s, "#") {
		if headingRE.MatchString(s) {
			st.addHeading(i+1, s, off)
		}
		return
	}
	for _, p := range skipPrefixes {
		if strings.HasPrefix(s, p) {
			return
		}
	}
	st.prose = append(st.prose, proseLine{num: i + 1, text: s, off: off})
}

// skipFenceOrComment tracks fenced code and multi-line HTML comments. It
// reports whether the line belongs to one of them. The oracle knows ```
// fences only. lexis also knows ~~~ fences and multi-line HTML comments.
func (st *mdState) skipFenceOrComment(s string) bool {
	assert.True(st.fence == 0 || !st.comment, "a fence and a comment do not nest")
	assert.True(!strings.ContainsRune(s, '\n'), "the text is one line")
	switch {
	case st.fence != 0:
		if strings.HasPrefix(s, strings.Repeat(string(st.fence), 3)) {
			st.fence = 0
		}
		return true
	case st.comment:
		st.comment = !strings.Contains(s, "-->")
		return true
	case strings.HasPrefix(s, "```"), strings.HasPrefix(s, "~~~"):
		st.fence = s[0]
		return true
	case strings.HasPrefix(s, "<!--") && !strings.Contains(s[4:], "-->"):
		st.comment = true
		return true
	}
	return false
}

// addHeading adds an ATX heading without its # marks.
func (st *mdState) addHeading(num int, s string, off int32) {
	assert.True(num > 0, "lines start at 1")
	assert.True(strings.HasPrefix(s, "#"), "a heading starts with #")
	text := strings.TrimLeft(s, "#")
	trimmed := trimSpace(text)
	start := off + int32(len(s)-len(text)) + int32(strings.Index(text, trimmed))
	var t Text
	t.add(trimmed, start, false)
	st.headings = append(st.headings, Block{Kind: Heading, Line: num, Text: clean(t)})
}

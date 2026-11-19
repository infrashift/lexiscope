package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/infrashift/lexiscope/internal/assert"
)

// Kind is the kind of a block.
type Kind uint8

// Block kinds.
const (
	Paragraph Kind = iota
	Heading
)

// Block is one paragraph, list item, quote or heading after cleanup.
type Block struct {
	Kind     Kind
	Comment  bool // the block comes from a source code comment
	Numbered bool // the block starts with "1. ": a step for the oracle
	Bullet   bool // the block starts with "- ", "* " or "+ "
	Line     int  // the 1-based first line: the line that the oracle reports
	Text     Text // the text after cleanup, without the list marker
}

// proseLine is one line of prose. It has a 1-based number, its text without
// white space at the ends, and the offset of the first byte of that text. A
// paragraph does not cross a change of group: each source comment has its
// own group, and all Markdown lines have group 0.
type proseLine struct {
	num   int
	text  string
	off   int32
	group int
}

// The cleanup patterns of the oracle (scripts/ste-check.py, clean and
// check_paragraph), with Python's Unicode \s and \d.
var (
	codeRE   = regexp.MustCompile("`[^`]*`")
	linkRE   = regexp.MustCompile(`\]\([^)]*\)`)
	tagRE    = regexp.MustCompile(`<[^>]+>`)
	urlRE    = regexp.MustCompile(`https?://[^` + PySpaceClass + `]+`)
	boldRE   = regexp.MustCompile(`\*\*`)
	starRE   = regexp.MustCompile(`\*`)
	markerRE = regexp.MustCompile(`^[` + PySpaceClass + `]*([-*+]|\p{Nd}+\.)[` + PySpaceClass + `]+(\[[ x]\][` + PySpaceClass + `]+)?`)
	stepRE   = regexp.MustCompile(`^\p{Nd}+\.[` + PySpaceClass + `]`)
	bulletRE = regexp.MustCompile(`^[-*+][` + PySpaceClass + `]`)
	itemRE   = regexp.MustCompile(`^([-*+]|\p{Nd}+\.)[` + PySpaceClass + `]`)
)

// clean applies the cleanup of the oracle. Inline code becomes CODE, and a
// link keeps its text. An HTML tag becomes a space, a URL becomes URL, and
// the emphasis markers go.
func clean(t Text) Text {
	assert.True(len(t.B) == len(t.Off), "each byte has an offset")
	assert.True(len(t.B) <= MaxBlockBytes, "a block is bounded")
	t = t.ReplaceAll(codeRE, "CODE")
	t = t.ReplaceAll(linkRE, "]")
	t = t.ReplaceAll(tagRE, " ")
	t = t.ReplaceAll(urlRE, "URL")
	t = t.ReplaceAll(boldRE, "")
	return t.ReplaceAll(starRE, "")
}

// paragraphs groups prose lines as the oracle does. A new paragraph starts
// after a gap in the line numbers, at a list item, and at a quote line.
func paragraphs(lines []proseLine, comment bool) ([]Block, error) {
	assert.True(len(lines) <= MaxLines, "the line count is bounded")
	var blocks []Block
	var para []proseLine
	for k, pl := range lines {
		newItem := itemRE.MatchString(pl.text)
		if k == 0 || pl.num != lines[k-1].num+1 || pl.group != lines[k-1].group || newItem || strings.HasPrefix(pl.text, ">") {
			if len(para) > 0 {
				blocks = append(blocks, paragraph(para, comment))
			}
			para = para[:0]
		}
		trimmed := strings.TrimLeft(pl.text, "> ")
		para = append(para, proseLine{num: pl.num, text: trimmed, off: pl.off + int32(len(pl.text)-len(trimmed))})
	}
	if len(para) > 0 {
		blocks = append(blocks, paragraph(para, comment))
	}
	if len(blocks) > MaxBlocks {
		return nil, fmt.Errorf("%d blocks, more than %d", len(blocks), MaxBlocks)
	}
	return blocks, nil
}

// paragraph joins the lines with one space, as " ".join does, then cleans
// the text and removes the list marker and a checkbox.
func paragraph(para []proseLine, comment bool) Block {
	assert.True(len(para) > 0, "a paragraph has a line")
	var raw Text
	for k, pl := range para {
		if k > 0 {
			raw.add(" ", para[k-1].off+int32(len(para[k-1].text)), true)
		}
		raw.add(pl.text, pl.off, false)
	}
	b := Block{
		Kind: Paragraph, Comment: comment, Line: para[0].num,
		Numbered: stepRE.Match(raw.B), Bullet: bulletRE.Match(raw.B),
	}
	b.Text = clean(raw).ReplaceAll(markerRE, "")
	assert.True(b.Line > 0, "lines start at 1")
	return b
}

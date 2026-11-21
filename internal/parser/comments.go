package parser

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	ztree "github.com/infrashift/ztree/bindings/go"

	"github.com/infrashift/lexiscope/internal/assert"
)

// directiveRE matches a tool directive after //, for example //go:build,
// //nolint or //export. A directive is not prose.
var directiveRE = regexp.MustCompile(`^(?:[a-z]+:[a-z]|nolint|export |line |extern |\+build)`)

// docTagRE matches the XML doc tags of C# at the start of a /// line, for
// example <summary> or <param name="x">. The prose after the tags counts.
var docTagRE = regexp.MustCompile(`^[ \t]*(?:</?(?:summary|remarks|param|typeparam|returns|value|exception|example|para|see|seealso|paramref|typeparamref|inheritdoc|c|list|listheader|item|term|description|permission|include)\b[^>]*>[ \t]*)+`)

// LangOf returns the grammar of a source file name.
func LangOf(name string) (ztree.Lang, bool) {
	assert.True(name != "", "the file has a name")
	assert.True(ztree.LangGo != ztree.LangZig, "the grammars differ")
	switch filepath.Ext(name) {
	case ".go":
		return ztree.LangGo, true
	case ".zig":
		return ztree.LangZig, true
	case ".py":
		return ztree.LangPython, true
	case ".c", ".h":
		return ztree.LangC, true
	case ".cc", ".cpp", ".cxx", ".hh", ".hpp", ".hxx":
		return ztree.LangCPP, true
	case ".rs":
		return ztree.LangRust, true
	case ".cs":
		return ztree.LangCSharp, true
	}
	return 0, false
}

// Comments returns the prose blocks of the comments of src. Line comments
// on their own lines and on adjacent lines form one comment. A trailing
// comment after code stands alone. Indented lines are code examples.
func Comments(lang ztree.Lang, src []byte) ([]Block, error) {
	assert.True(len(src) <= MaxFileBytes || src != nil, "the source exists")
	if !utf8.Valid(src) {
		return nil, errors.New("the file is not valid UTF-8")
	}
	lines, err := NewLines(src)
	if err != nil {
		return nil, err
	}
	spans, err := ztree.ExtractComments(lang, src)
	if err != nil {
		return nil, err // Each ztree error starts with "ztree:".
	}
	var prose []proseLine
	group, prevEnd, prevOwnLine := 0, int32(-1), false
	doc := docCode{xml: lang == ztree.LangCSharp}
	for _, s := range spans {
		start, end := int32(s.StartByte), int32(s.EndByte)
		own := ownLine(src, start)
		lineComment := !strings.HasPrefix(string(src[start:end]), "/*")
		if lineComment {
			end = trimLineBreak(src, start, end)
		}
		if !own || !prevOwnLine || !lineComment || !adjacent(lines, prevEnd, start) {
			group++
		}
		if !doc.skip(string(src[start:end])) {
			prose = append(prose, commentLines(lines, src, start, end, group, doc.xml)...)
		}
		prevEnd, prevOwnLine = end, own && lineComment
	}
	return paragraphs(prose, true)
}

// docCode follows the <code> blocks of the XML doc comments of C#. Each
// /// line is a separate comment, so the state lasts from line to line.
type docCode struct {
	xml    bool // the comments are C# comments with XML doc tags
	inCode bool // the previous /// line opened a <code> block
}

// skip reports whether the comment text is a line of a <code> block. The
// lines with the <code> and </code> tags are part of the block.
func (d *docCode) skip(text string) bool {
	assert.True(text != "", "a comment has text")
	assert.True(d.xml || !d.inCode, "only C# has code blocks")
	if !d.xml || !strings.HasPrefix(text, "///") {
		return false
	}
	opens, closes := strings.Contains(text, "<code"), strings.Contains(text, "</code>")
	skip := d.inCode || opens
	d.inCode = (d.inCode || opens) && !closes
	return skip
}

// trimLineBreak returns the end of the line comment src[start:end] without
// its line break. A Rust doc comment (/// or //!) ends after the line break.
func trimLineBreak(src []byte, start, end int32) int32 {
	assert.True(start < end && int(end) <= len(src), "the comment is in the source")
	for end > start+1 && (src[end-1] == '\n' || src[end-1] == '\r') {
		end--
	}
	assert.True(end > start, "a line comment keeps its marker")
	return end
}

// ownLine reports whether only white space precedes off on its line.
func ownLine(src []byte, off int32) bool {
	assert.True(off >= 0 && int(off) <= len(src), "the offset is in the source")
	assert.True(len(src) <= MaxFileBytes, "the source is bounded")
	for i := off - 1; i >= 0 && src[i] != '\n' && src[i] != '\r'; i-- {
		if src[i] != ' ' && src[i] != '\t' {
			return false
		}
	}
	return true
}

// adjacent reports whether start is on the line after the line of prevEnd.
func adjacent(lines *Lines, prevEnd, start int32) bool {
	assert.True(lines != nil, "the table exists")
	assert.True(start > prevEnd, "the comments are in order")
	if prevEnd < 0 {
		return false
	}
	return lines.LineOf(start) == lines.LineOf(prevEnd)+1
}

// commentLines returns the prose lines of the comment src[start:end]. With
// xml, a /// line loses the XML doc tags at its start.
func commentLines(lines *Lines, src []byte, start, end int32, group int, xml bool) []proseLine {
	assert.True(start < end && int(end) <= len(src), "the comment is in the source")
	assert.True(group > 0, "each comment has a group")
	text := string(src[start:end])
	var out []proseLine
	if body, ok := strings.CutPrefix(text, "/*"); ok {
		body = strings.TrimSuffix(body, "*/")
		off := start + 2
		for k, line := range strings.Split(body, "\n") {
			content := line
			if k > 0 {
				content = strings.TrimLeft(content, " \t")
				content = strings.TrimPrefix(content, "*")
			}
			out = appendProse(out, lines, content, off+int32(len(line)-len(content)), group)
			off += int32(len(line)) + 1
		}
		return out
	}
	marker := "#"
	if strings.HasPrefix(text, "//") {
		marker = "//"
	}
	// In C and C++, a backslash at the end of a line comment continues the
	// comment on the next line. Each line is a separate comment line.
	parts := strings.Split(text, "\n")
	off := start
	for k, line := range parts {
		body := line
		if k > 0 {
			body = strings.TrimLeft(body, " \t")
		}
		lead := int32(len(line) - len(body))
		if k < len(parts)-1 {
			body = strings.TrimSuffix(strings.TrimSuffix(body, "\r"), "\\")
		}
		out = appendLineComment(out, lines, body, marker, off+lead, group, xml)
		off += int32(len(line)) + 1
	}
	assert.True(off == end+1, "the lines cover the comment")
	return out
}

// appendLineComment adds the prose of one line of a line comment. The line
// starts at off. A tool directive and a shebang line are not prose. With
// xml, a doc comment (///) loses the XML doc tags at its start.
func appendLineComment(out []proseLine, lines *Lines, line, marker string, off int32, group int, xml bool) []proseLine {
	assert.True(off >= 0, "an offset is not negative")
	assert.True(marker == "//" || marker == "#", "the marker is known")
	content := strings.TrimPrefix(line, marker)
	if marker == "//" && directiveRE.MatchString(content) || strings.HasPrefix(content, "!") && marker == "#" {
		return out
	}
	docLine := xml && strings.HasPrefix(content, "/")
	content = strings.TrimLeft(content, "/!") // doc comments: ///, //!
	if docLine {
		content = content[len(docTagRE.FindString(content)):]
	}
	return appendProse(out, lines, content, off+int32(len(line)-len(content)), group)
}

// appendProse adds one comment line unless it is empty, code or a skipped
// line. A code line starts with a tab or with two spaces or more.
func appendProse(out []proseLine, lines *Lines, content string, off int32, group int) []proseLine {
	assert.True(off >= 0, "an offset is not negative")
	assert.True(!strings.Contains(content, "\n"), "the content is one line")
	content = strings.TrimSuffix(content, "\r")
	body := strings.TrimPrefix(content, " ")
	if strings.HasPrefix(body, "\t") || strings.HasPrefix(body, " ") {
		return out
	}
	s := trimSpace(body)
	if s == "" || strings.HasPrefix(s, "#") {
		return out
	}
	for _, p := range skipPrefixes {
		if strings.HasPrefix(s, p) {
			return out
		}
	}
	at := off + int32(strings.Index(content, s))
	return append(out, proseLine{num: lines.LineOf(at), text: s, off: at, group: group})
}

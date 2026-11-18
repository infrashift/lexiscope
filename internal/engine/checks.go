package engine

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/parser"
)

// minGerund is the shortest word that the -ing check reads, so "ring" and
// "king" do not count.
const minGerund = 5

// blockCheck runs the rules on one block.
type blockCheck struct {
	path       string
	lines      *parser.Lines
	block      parser.Block
	vocabulary map[string]bool
}

// run dispatches on the check of the rule.
func (bc *blockCheck) run(c *compiled) []Finding {
	assert.True(c != nil, "the rule exists")
	assert.True(len(bc.block.Text.B) > 0, "the block has text")
	s := bc.block.Text.String()
	p := c.rule.Params
	switch c.rule.Check {
	case "max_words":
		return bc.maxWords(c, s)
	case "banned_terms", "possessive_pronouns":
		return bc.terms(c, matchTerms(s, c.terms), nil)
	case "term_substitution":
		return bc.terms(c, matchTerms(s, c.terms), p.Swap)
	case "passive_voice":
		return bc.terms(c, matchPassive(s, c.terms, c.suffixes, p.Participles, p.Exceptions), nil)
	case "nominalizations":
		return bc.terms(c, bc.wordMatches(s, c.suffixes, p.Swap, p.Exceptions, 3), p.Swap)
	case "ambiguous_gerunds":
		return bc.terms(c, bc.wordMatches(s, []string{"ing"}, nil, p.Exceptions, minGerund-3), nil)
	case "one_instruction_per_sentence":
		return bc.instructions(c, s)
	case "regex_ban":
		return bc.regex(c, s)
	}
	assert.True(false, "the schema allows each check above only")
	return nil
}

// maxWords reports each sentence with more words than the limit.
func (bc *blockCheck) maxWords(c *compiled, s string) []Finding {
	assert.True(c.rule.Params.MaxWords > 0, "the schema requires max_words")
	assert.True(s != "", "the block has text")
	var out []Finding
	for _, sp := range sentences(s) {
		n := wordCount(s[sp.start:sp.end])
		if n <= c.rule.Params.MaxWords {
			continue
		}
		f := bc.finding(c, sp, strconv.Itoa(n), "", map[string]string{
			"{words}": strconv.Itoa(n), "{max_words}": strconv.Itoa(c.rule.Params.MaxWords),
		})
		out = append(out, f)
	}
	return out
}

// terms turns matches into findings. With swap, each finding names the
// replacement of its term.
func (bc *blockCheck) terms(c *compiled, ms []match, swap map[string]string) []Finding {
	assert.True(c != nil, "the rule exists")
	assert.True(len(ms) <= parser.MaxBlockBytes, "the match count is bounded")
	out := make([]Finding, 0, len(ms))
	for _, m := range ms {
		repl := lookupFold(swap, m.term)
		f := bc.finding(c, m.span, m.text, m.text, map[string]string{"{term}": m.text, "{replacement}": repl})
		f.Replacement = repl
		out = append(out, f)
	}
	return out
}

// lookupFold returns the value of key in m without case, or "".
func lookupFold(m map[string]string, key string) string {
	assert.True(key != "" || len(m) == 0, "a key is not empty")
	assert.True(len(m) <= 1<<16, "the map is bounded")
	if v, ok := m[key]; ok {
		return v
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

// matchPassive finds an auxiliary, white space, then a word. The word is a
// participle, or it ends in a suffix with one character or more before it.
// With the auxiliaries of the oracle and the suffix "ed", it matches as
// \b(is|are|...)\s+\w+ed\b with re.I does.
func matchPassive(s string, aux, suffixes []string, participles, exceptions map[string]bool) []match {
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	assert.True(len(aux) > 0, "the schema requires auxiliaries")
	var out []match
	for i := 0; i < len(s); {
		if m, ok := matchAt(s, i, aux); ok {
			j := spaceRun(s, m.end)
			k := wordEnd(s, j)
			w := strings.ToLower(s[j:k])
			if j > m.end && k > j && !exceptions[w] && (participles[w] || hasSuffix(w, suffixes, 1)) {
				out = append(out, match{span: span{m.start, k}, text: s[m.start:k], term: w})
				i = k
				continue
			}
		}
		i += runeSize(s[i:])
	}
	return out
}

// wordEnd returns the end of the run of word characters at i.
func wordEnd(s string, i int) int {
	assert.True(0 <= i && i <= len(s), "the index is in the text")
	assert.True(len(s) <= parser.MaxBlockBytes, "a block is bounded")
	for i < len(s) {
		r := firstRune(s[i:])
		if !parser.IsWord(r) {
			break
		}
		i += runeSize(s[i:])
	}
	return i
}

// hasSuffix reports whether w ends in a suffix with at least min characters
// before it.
func hasSuffix(w string, suffixes []string, minBefore int) bool {
	assert.True(minBefore >= 0, "the minimum is not negative")
	assert.True(len(suffixes) <= 1<<16, "the suffix list is bounded")
	for _, suf := range suffixes {
		if strings.HasSuffix(w, suf) && len([]rune(w))-len([]rune(suf)) >= minBefore {
			return true
		}
	}
	return false
}

// wordMatches returns the words that swap names, or that end in a suffix
// with at least min characters before it. It skips exceptions and the
// vocabulary.
func (bc *blockCheck) wordMatches(s string, suffixes []string, swap map[string]string, exceptions map[string]bool, minBefore int) []match {
	assert.True(len(suffixes) > 0, "the check has a suffix")
	assert.True(minBefore >= 0, "the minimum is not negative")
	var out []match
	for _, sp := range words(s) {
		w := strings.ToLower(s[sp.start:sp.end])
		if exceptions[w] || bc.vocabulary[w] || identifier(s[sp.start:sp.end]) {
			continue
		}
		if lookupFold(swap, w) != "" || hasSuffix(w, suffixes, minBefore) {
			out = append(out, match{span: sp, text: s[sp.start:sp.end], term: w})
		}
	}
	return out
}

// identifier reports whether a word looks like a code name: a capital
// letter after the first letter, a digit or an underscore. The word checks
// skip code names, for example "skipFenceOrComment" in a comment.
func identifier(w string) bool {
	assert.True(w != "", "a word is not empty")
	assert.True(len(w) <= parser.MaxBlockBytes, "a word is bounded")
	for i, r := range w {
		if r == '_' || unicode.IsDigit(r) || (i > 0 && unicode.IsUpper(r)) {
			return true
		}
	}
	return false
}

// instructions reports a sentence that starts with a verb of the list. A
// joiner and a second verb of the list follow later, for example "Run the
// build and copy the file."
func (bc *blockCheck) instructions(c *compiled, s string) []Finding {
	assert.True(len(c.rule.Params.Verbs) > 0, "the schema requires verbs")
	assert.True(s != "", "the block has text")
	verbs, joiners := lowerSet(c.rule.Params.Verbs), lowerSet(c.rule.Params.Joiners)
	var out []Finding
	for _, sp := range sentences(s) {
		ws := words(s[sp.start:sp.end])
		if len(ws) == 0 || !verbs[strings.ToLower(s[sp.start+ws[0].start:sp.start+ws[0].end])] {
			continue
		}
		for k := 1; k+1 < len(ws); k++ {
			a := strings.ToLower(s[sp.start+ws[k].start : sp.start+ws[k].end])
			b := strings.ToLower(s[sp.start+ws[k+1].start : sp.start+ws[k+1].end])
			if joiners[a] && verbs[b] {
				out = append(out, bc.finding(c, sp, a+" "+b, "", nil))
				break
			}
		}
	}
	return out
}

// regex reports each match of the pattern of the rule.
func (bc *blockCheck) regex(c *compiled, s string) []Finding {
	assert.True(c.re != nil, "the pattern compiled")
	assert.True(s != "", "the block has text")
	var out []Finding
	for _, m := range c.re.FindAllStringIndex(s, -1) {
		if m[1] == m[0] {
			continue // an empty match marks no text
		}
		text := s[m[0]:m[1]]
		out = append(out, bc.finding(c, span{m[0], m[1]}, text, text, map[string]string{"{term}": text}))
	}
	return out
}

// finding builds a finding for the bytes sp of the block text.
func (bc *blockCheck) finding(c *compiled, sp span, detail, term string, vars map[string]string) Finding {
	assert.True(sp.start < sp.end || sp.end == 0, "the span is in order")
	assert.True(sp.end <= len(bc.block.Text.B), "the span is in the block")
	t := bc.block.Text
	start, end := t.Start(min(sp.start, len(t.B)-1)), t.End(max(sp.end, 1))
	line, col := bc.lines.Position(start)
	endLine, endCol := bc.lines.Position(max(end, start))
	msg := c.rule.Message
	for k, v := range vars {
		msg = strings.ReplaceAll(msg, k, v)
	}
	return Finding{
		Path: bc.path, Line: line, Column: col, EndLine: endLine, EndColumn: endCol,
		RuleID: c.rule.ID, Severity: c.rule.Severity, Message: msg, Term: term,
		BlockLine: bc.block.Line, Detail: detail,
	}
}

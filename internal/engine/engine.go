// Package engine runs the rules of a resolved ruleset on the prose blocks of
// a file and returns the findings.
//
// The checks of the core rules copy scripts/ste-check.py (ADR-0005). They
// copy its sentence split, its word count, its \b boundaries and its re.I
// flag. They use the Unicode classes of Python's str patterns.
package engine

import (
	"cmp"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/parser"
	"github.com/infrashift/lexis/internal/ruleset"
)

// Finding is one result of one rule. Line and Column are 1-based. Columns
// count UTF-16 code units, as SARIF does.
type Finding struct {
	Path        string `json:"path"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column"`
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Term        string `json:"term,omitempty"`
	Replacement string `json:"replacement,omitempty"`
	// BlockLine is the first line of the block: the line that the oracle
	// reports. Detail is the matched text or the word count, as the oracle
	// reports it. The parity test compares both.
	BlockLine int    `json:"-"`
	Detail    string `json:"-"`
}

// compiled is a rule with its term lists in match order.
type compiled struct {
	rule     ruleset.Rule
	terms    []string // tokens, swap keys or auxiliaries, longest first
	suffixes []string
	re       *regexp.Regexp
}

// Engine checks files against the active rules of one ruleset.
type Engine struct {
	rules        []compiled
	vocabulary   map[string]bool
	stepNumbered bool
	stepBullet   bool
}

// New compiles the active rules of rs.
func New(rs *ruleset.Ruleset) (*Engine, error) {
	assert.True(rs != nil, "the ruleset exists")
	assert.True(len(rs.Rules) <= ruleset.MaxRules, "the rule count is bounded")
	e := &Engine{vocabulary: lowerSet(rs.Vocabulary)}
	for _, r := range rs.Active() {
		c := compiled{rule: r, suffixes: sortedTerms(r.Params.Suffixes)}
		switch r.Check {
		case "banned_terms", "possessive_pronouns":
			c.terms = sortedTerms(r.Params.Tokens)
		case "term_substitution":
			c.terms = sortedTerms(boolSet(slices.Collect(maps.Keys(r.Params.Swap))))
		case "passive_voice":
			c.terms = sortedTerms(r.Params.Auxiliaries)
		case "regex_ban":
			re, err := regexp.Compile(r.Params.Pattern)
			if err != nil {
				return nil, fmt.Errorf("rule %s: pattern: %w", r.ID, err)
			}
			c.re = re
		}
		if r.Scope == "procedural_sentence" {
			e.stepNumbered = e.stepNumbered || slices.Contains(markers(r), "numbered")
			e.stepBullet = e.stepBullet || slices.Contains(markers(r), "bullet")
		}
		e.rules = append(e.rules, c)
	}
	return e, nil
}

// markers returns the procedural markers of a rule. The default is the rule
// of the oracle: numbered items only.
func markers(r ruleset.Rule) []string {
	assert.True(r.ID != "", "the rule has an ID")
	assert.True(len(r.Params.ProceduralMarkers) <= 2, "a rule has two markers or fewer")
	if len(r.Params.ProceduralMarkers) == 0 {
		return []string{"numbered"}
	}
	return r.Params.ProceduralMarkers
}

// Check returns the findings of one file in order of position. A file that
// is neither Markdown nor a known source file has no findings.
func (e *Engine) Check(path string, src []byte) ([]Finding, error) {
	assert.True(e != nil, "the engine exists")
	assert.True(path != "", "the file has a path")
	blocks, err := extract(path, src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	lines, err := parser.NewLines(src)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var out []Finding
	for _, b := range blocks {
		if len(b.Text.B) == 0 {
			continue
		}
		bc := &blockCheck{path: path, lines: lines, block: b, vocabulary: e.vocabulary}
		for k := range e.rules {
			if e.applies(&e.rules[k].rule, &b) {
				out = append(out, bc.run(&e.rules[k])...)
			}
		}
	}
	slices.SortStableFunc(out, compareFindings) // nasa-lint: allow(R9) slices takes a comparison function.
	return out, nil
}

// extract selects the extractor by the file extension.
func extract(path string, src []byte) ([]parser.Block, error) {
	assert.True(path != "", "the file has a path")
	assert.True(len(src) <= parser.MaxFileBytes || src != nil, "the source exists")
	if IsMarkdown(path) {
		return parser.Markdown(path, src)
	}
	if lang, ok := parser.LangOf(path); ok {
		return parser.Comments(lang, src)
	}
	return nil, nil
}

// IsMarkdown reports whether lexis reads path as Markdown.
func IsMarkdown(path string) bool {
	assert.True(path != "", "the file has a path")
	ext := filepath.Ext(path)
	assert.True(!strings.Contains(ext, "/"), "an extension has no slash")
	return ext == ".md" || ext == ".mdx" || ext == ".markdown"
}

// applies reports whether the scope of r covers the block.
func (e *Engine) applies(r *ruleset.Rule, b *parser.Block) bool {
	assert.True(r != nil && b != nil, "the rule and the block exist")
	assert.True(b.Line > 0, "lines start at 1")
	step := (e.stepNumbered && b.Numbered) || (e.stepBullet && b.Bullet)
	if !e.stepNumbered && !e.stepBullet {
		step = b.Numbered
	}
	para := b.Kind == parser.Paragraph
	switch r.Scope {
	case "procedural_sentence":
		return para && step
	case "descriptive_sentence":
		return para && !step
	case "any_sentence":
		return para
	case "heading":
		return b.Kind == parser.Heading
	case "comment":
		return b.Comment
	}
	return true // token: each block
}

func compareFindings(a, b Finding) int {
	assert.True(a.Path == b.Path, "the findings of one file")
	assert.True(a.Line > 0 && b.Line > 0, "lines start at 1")
	return cmp.Or(cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column), strings.Compare(a.RuleID, b.RuleID))
}

// sortedTerms returns the lower-case terms of a set, longest first.
func sortedTerms(set map[string]bool) []string {
	assert.True(len(set) <= 1<<16, "the set is bounded")
	out := slices.Collect(maps.Keys(lowerSet(set)))
	slices.SortFunc(out, longestFirst) // nasa-lint: allow(R9) slices takes a comparison function.
	assert.True(len(out) <= len(set), "lower case can merge terms")
	return out
}

func lowerSet(set map[string]bool) map[string]bool {
	assert.True(len(set) <= 1<<16, "the set is bounded")
	out := make(map[string]bool, len(set))
	for t, on := range set {
		if on {
			out[strings.ToLower(t)] = true
		}
	}
	assert.True(len(out) <= len(set), "lower case can merge terms")
	return out
}

func boolSet(terms []string) map[string]bool {
	assert.True(len(terms) <= 1<<16, "the list is bounded")
	out := make(map[string]bool, len(terms))
	for _, t := range terms {
		out[t] = true
	}
	assert.True(len(out) <= len(terms), "a set has no duplicates")
	return out
}

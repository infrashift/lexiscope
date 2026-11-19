// Package ruleset loads CUE rulesets (ADR-0004). It unifies the schema in
// schema.cue with the built-in ruleset, each user file and the CLI flags,
// then decodes the result into a Ruleset.
//
// CUE unification is commutative, so the order of the files does not change
// the result. A built-in value that a user can change is a CUE default. A
// user file or a flag that gives a second concrete value for a field fails
// with the path of the field.
package ruleset

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"

	"github.com/infrashift/lexiscope/internal/assert"
)

// Limits of one load.
const (
	MaxSources   = 64
	MaxFiles     = 256 // CUE files in one directory
	MaxFileBytes = 1 << 20
	MaxRules     = 1024
)

//go:embed schema.cue
var schemaSource []byte

// Concept is one entry of the lexicon.
type Concept struct {
	Approved        string   `json:"approved"`
	Unapproved      []string `json:"unapproved"`
	Nominalizations []string `json:"nominalizations"`
}

// Params holds the parameters of a rule. A check reads only its own fields.
type Params struct {
	MaxWords          int               `json:"max_words,omitempty"`
	ProceduralMarkers []string          `json:"procedural_markers,omitempty"`
	Tokens            map[string]bool   `json:"tokens,omitempty"`
	Swap              map[string]string `json:"swap,omitempty"`
	Exceptions        map[string]bool   `json:"exceptions,omitempty"`
	Suffixes          map[string]bool   `json:"suffixes,omitempty"`
	Auxiliaries       map[string]bool   `json:"auxiliaries,omitempty"`
	Participles       map[string]bool   `json:"participles,omitempty"`
	Verbs             map[string]bool   `json:"verbs,omitempty"`
	Joiners           map[string]bool   `json:"joiners,omitempty"`
	Pattern           string            `json:"pattern,omitempty"`
}

// Rule is one rule of a resolved ruleset.
type Rule struct {
	ID       string `json:"id"`
	Enabled  bool   `json:"enabled"`
	Severity string `json:"severity"`
	Scope    string `json:"scope"`
	Check    string `json:"check"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Params   Params `json:"params"`
}

// Ruleset is a resolved ruleset.
type Ruleset struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Lexicon     map[string]Concept `json:"lexicon"`
	Vocabulary  map[string]bool    `json:"vocabulary"`
	Rules       map[string]Rule    `json:"rules"`
}

// Override is one --set flag: rule_id.param=value. Value is CUE syntax, for
// example 18, "text", ["numbered", "bullet"] or {term: true}. A value that
// is not valid CUE is a string.
type Override struct {
	RuleID string
	Param  string
	Value  string
}

// Options selects the sources and the changes of a load.
type Options struct {
	Rulesets   []string // built-in names, CUE files or directories
	Enable     []string
	Omit       []string
	Overrides  []Override
	AllowTerms []string
}

// Load resolves the ruleset that the options select.
func Load(o Options) (*Ruleset, error) {
	assert.True(len(o.Rulesets) > 0, "the options name a ruleset")
	assert.True(len(o.Rulesets) <= MaxSources, "the source count is bounded")
	srcs, err := resolve(o.Rulesets)
	if err != nil {
		return nil, err
	}
	ctx := cuecontext.New()
	schema := ctx.CompileBytes(schemaSource, cue.Filename("schema.cue"))
	if err := schema.Err(); err != nil {
		return nil, cueError("schema.cue", err)
	}
	v := schema.LookupPath(cue.ParsePath("#Ruleset"))
	for _, src := range srcs {
		f := ctx.CompileBytes(src.data, cue.Filename(src.name), cue.Scope(schema))
		if err := f.Err(); err != nil {
			return nil, cueError(src.name, err)
		}
		v = v.Unify(f)
	}
	if v, err = applyFlags(ctx, v, o); err != nil {
		return nil, err
	}
	return decode(v, sourceNames(srcs))
}

// decode validates the unified value and decodes it.
func decode(v cue.Value, names string) (*Ruleset, error) {
	assert.True(v.Exists(), "the value exists")
	assert.True(names != "", "the sources have names")
	if err := v.Validate(cue.Concrete(true)); err != nil {
		return nil, cueError(names, err)
	}
	var rs Ruleset
	if err := v.Decode(&rs); err != nil {
		return nil, cueError(names, err)
	}
	if len(rs.Rules) > MaxRules {
		return nil, fmt.Errorf("ruleset %s: %d rules, more than %d", rs.Name, len(rs.Rules), MaxRules)
	}
	return &rs, nil
}

// cueError formats the details of a CUE error, one problem on each line.
func cueError(where string, err error) error {
	assert.True(where != "", "the error has a source")
	assert.True(err != nil, "the error exists")
	details := strings.TrimSpace(errors.Details(err, nil))
	details = strings.ReplaceAll(details, "#Ruleset.", "")
	return fmt.Errorf("ruleset %s: %s", where, details)
}

// Sorted returns the rules in the order of their IDs.
func (rs *Ruleset) Sorted() []Rule {
	assert.True(rs != nil, "the ruleset exists")
	assert.True(len(rs.Rules) <= MaxRules, "the rule count is bounded")
	ids := make([]string, 0, len(rs.Rules))
	for id := range rs.Rules {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	out := make([]Rule, len(ids))
	for i, id := range ids {
		out[i] = rs.Rules[id]
	}
	return out
}

// Active returns the enabled rules in the order of their IDs.
func (rs *Ruleset) Active() []Rule {
	assert.True(rs != nil, "the ruleset exists")
	all := rs.Sorted()
	out := make([]Rule, 0, len(all))
	for _, r := range all {
		if r.Enabled {
			out = append(out, r)
		}
	}
	assert.True(len(out) <= len(all), "the active rules are a subset")
	return out
}

package cli

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/pflag"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/ruleset"
)

// maxListItems bounds each list flag.
const maxListItems = 1 << 12

var (
	ruleIDRE = regexp.MustCompile(`^[A-Z][A-Z0-9]*(-[A-Z0-9]+)*$`)
	paramRE  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// rulesetFlags are the flags that select and change a ruleset. lexis scan
// and lexis rules list share them.
type rulesetFlags struct {
	rulesets, enable, omit, sets, allowTerms []string
}

// register adds the flags to fl.
func (f *rulesetFlags) register(fl *pflag.FlagSet) {
	assert.True(f != nil, "the flags exist")
	assert.True(fl != nil, "the flag set exists")
	fl.StringSliceVarP(&f.rulesets, "ruleset", "r", []string{"ste-100"}, "built-in ruleset name, or CUE file or directory")
	fl.StringSliceVar(&f.enable, "enable-rule", nil, "turn on these rule IDs")
	fl.StringSliceVar(&f.omit, "omit-rule", nil, "turn off these rule IDs (alias: --disable-rule)")
	fl.StringArrayVar(&f.sets, "set", nil, "change a rule parameter: rule_id.param=value")
	fl.StringSliceVar(&f.allowTerms, "allow-term", nil, "add approved technical terms")
	fl.SetNormalizeFunc(normalizeRuleFlag) // nasa-lint: allow(R9) pflag maps flag names through this function.
}

// normalizeRuleFlag makes --disable-rule an alias of --omit-rule.
func normalizeRuleFlag(fs *pflag.FlagSet, name string) pflag.NormalizedName {
	assert.True(fs != nil, "pflag gives the flag set")
	assert.True(name != "", "a flag has a name")
	if name == "disable-rule" {
		name = "omit-rule"
	}
	return pflag.NormalizedName(name)
}

// options checks the flag values and returns the ruleset options.
func (f *rulesetFlags) options() (ruleset.Options, error) {
	assert.True(f != nil, "the flags exist")
	assert.True(len(f.enable) <= maxListItems && len(f.omit) <= maxListItems, "the rule lists are bounded")
	o := ruleset.Options{Rulesets: f.rulesets, Enable: f.enable, Omit: f.omit, AllowTerms: f.allowTerms}
	if len(o.Rulesets) == 0 || slices.Contains(o.Rulesets, "") {
		return ruleset.Options{}, errors.New("--ruleset: give a ruleset name or a CUE path")
	}
	if len(o.Rulesets) > ruleset.MaxSources {
		return ruleset.Options{}, fmt.Errorf("--ruleset: more than %d rulesets", ruleset.MaxSources)
	}
	for _, id := range slices.Concat(o.Enable, o.Omit) {
		if !ruleIDRE.MatchString(id) {
			return ruleset.Options{}, fmt.Errorf("%q is not a rule ID such as STE-100-01", id)
		}
	}
	for _, id := range o.Enable {
		if slices.Contains(o.Omit, id) {
			return ruleset.Options{}, fmt.Errorf("%s: --enable-rule and --omit-rule name the same rule", id)
		}
	}
	for _, term := range o.AllowTerms {
		if strings.TrimSpace(term) == "" {
			return ruleset.Options{}, errors.New("--allow-term: a term is empty")
		}
	}
	for _, s := range f.sets {
		ov, err := parseOverride(s)
		if err != nil {
			return ruleset.Options{}, err
		}
		o.Overrides = append(o.Overrides, ov)
	}
	return o, nil
}

// parseOverride reads rule_id.param=value. The ruleset package reads the
// value as CUE.
func parseOverride(s string) (ruleset.Override, error) {
	assert.True(len(s) < 1<<16, "the flag value is bounded")
	key, value, ok := strings.Cut(s, "=")
	if !ok {
		return ruleset.Override{}, fmt.Errorf("--set %q: use rule_id.param=value", s)
	}
	id, param, ok := strings.Cut(key, ".")
	if !ok || !ruleIDRE.MatchString(id) || !paramRE.MatchString(param) {
		return ruleset.Override{}, fmt.Errorf("--set %q: use rule_id.param=value, for example STE-100-01.max_words=18", s)
	}
	ov := ruleset.Override{RuleID: id, Param: param, Value: value}
	assert.True(ov.RuleID != "" && ov.Param != "", "the override names a rule and a parameter")
	return ov, nil
}

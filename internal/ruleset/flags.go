package ruleset

import (
	"fmt"

	"cuelang.org/go/cue"

	"github.com/infrashift/lexis/internal/assert"
)

// maxFlagItems bounds each list of CLI changes.
const maxFlagItems = 1 << 12

// applyFlags fills the CLI changes into the unified value: --enable-rule,
// --omit-rule, --set and --allow-term. A change that conflicts with a file
// fails when decode validates the value.
func applyFlags(ctx *cue.Context, v cue.Value, o Options) (cue.Value, error) {
	assert.True(ctx != nil, "the CUE context exists")
	assert.True(len(o.Enable)+len(o.Omit)+len(o.Overrides)+len(o.AllowTerms) <= 4*maxFlagItems, "the flag lists are bounded")
	for _, id := range o.Enable {
		if err := requireRule(v, id); err != nil {
			return v, err
		}
		v = v.FillPath(rulePath(id, "enabled"), true)
	}
	for _, id := range o.Omit {
		if err := requireRule(v, id); err != nil {
			return v, err
		}
		v = v.FillPath(rulePath(id, "enabled"), false)
	}
	for _, ov := range o.Overrides {
		if err := requireRule(v, ov.RuleID); err != nil {
			return v, err
		}
		v = v.FillPath(rulePath(ov.RuleID, "params", ov.Param), parseValue(ctx, ov.Value))
	}
	for _, term := range o.AllowTerms {
		v = v.FillPath(cue.MakePath(cue.Str("vocabulary"), cue.Str(term)), true)
	}
	return v, nil
}

// requireRule fails when the ruleset has no rule id.
func requireRule(v cue.Value, id string) error {
	assert.True(v.Exists(), "the value exists")
	assert.True(id != "", "the rule ID is set")
	if !v.LookupPath(rulePath(id)).Exists() {
		return fmt.Errorf("ruleset: no rule %s", id)
	}
	return nil
}

// parseValue reads s as a concrete CUE value. Text that is not a concrete
// CUE value, for example a bare word, becomes a CUE string.
func parseValue(ctx *cue.Context, s string) cue.Value {
	assert.True(ctx != nil, "the CUE context exists")
	assert.True(len(s) < 1<<16, "the value is bounded")
	v := ctx.CompileString(s)
	if v.Err() == nil && v.Validate(cue.Concrete(true)) == nil {
		return v
	}
	str := ctx.Encode(s)
	assert.True(str.Err() == nil, "a Go string encodes")
	return str
}

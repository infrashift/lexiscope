// Package ste100 embeds the CUE source of the ste-100 ruleset and its
// lexicon (ste100.cue). internal/ruleset loads it.
package ste100

import (
	_ "embed"
	"slices"

	"github.com/infrashift/lexiscope/internal/assert"
)

// Name is the name of the built-in ruleset.
const Name = "ste-100"

//go:embed ste100.cue
var source []byte

// Source returns a copy of the CUE source of the ruleset.
func Source() []byte {
	assert.True(len(source) > 0, "the build embeds ste100.cue")
	out := slices.Clone(source)
	assert.True(len(out) == len(source), "the copy is complete")
	return out
}

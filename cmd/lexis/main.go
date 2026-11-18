// Command lexis checks Markdown files and source code comments against CUE
// rulesets. Phase 2 of PLAN.md adds the commands.
package main

import (
	"fmt"
	"os"

	"github.com/infrashift/lexis/internal/assert"
)

func main() {
	assert.True(len(os.Args) > 0, "the program has a name")
	assert.True(os.Stderr != nil, "stderr exists")
	fmt.Fprintln(os.Stderr, "lexis: no commands yet")
	os.Exit(2)
}

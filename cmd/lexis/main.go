// Command lexis checks Markdown files and source code comments against CUE
// rulesets. Phase 2 of PLAN.md adds the commands.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "lexis: no commands yet")
	os.Exit(2)
}

// Command lexis checks Markdown files and source code comments against CUE
// rulesets. Package internal/cli holds the commands.
package main

import (
	"os"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/cli"
)

func main() {
	assert.True(len(os.Args) > 0, "the program has a name")
	assert.True(os.Stdout != nil && os.Stderr != nil, "the output streams exist")
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}

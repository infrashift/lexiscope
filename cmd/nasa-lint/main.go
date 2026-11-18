// Command nasa-lint checks the Go source of the module in the working
// directory against the NASA Power of 10 rules (ADR-0003). It is the Go port
// of scripts/nasa-lint.py. Package internal/nasalint lists the checks.
//
// Usage: go run ./cmd/nasa-lint [--verbose]
//
// With --verbose, it lists each function with fewer than 2 assertions. The
// exit status is 0 without findings, 1 with findings and 2 on a failure.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/nasalint"
)

func main() {
	assert.True(len(os.Args) > 0, "the program has a name")
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) (code int) {
	assert.True(stdout != nil && stderr != nil, "the output streams exist")
	assert.True(len(args) < 1<<16, "the argument count is bounded")
	defer assert.Recover(stderr, "nasa-lint", &code)
	verbose := false
	for _, a := range args {
		if a != "--verbose" && a != "-v" {
			fmt.Fprintf(stderr, "nasa-lint: unknown argument %q\nusage: nasa-lint [--verbose]\n", a)
			return 2
		}
		verbose = true
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "nasa-lint: %v\n", err)
		return 2
	}
	res, err := nasalint.Run(root)
	if err != nil {
		fmt.Fprintf(stderr, "nasa-lint: %v\n", err)
		return 2
	}
	write(stdout, res, verbose)
	if len(res.Findings) > 0 {
		return 1
	}
	return 0
}

// write prints the findings and the summary in the format of
// scripts/nasa-lint.py.
func write(w io.Writer, res *nasalint.Result, verbose bool) {
	assert.True(w != nil, "the writer exists")
	assert.True(res != nil, "the result exists")
	if verbose {
		for _, f := range res.Funcs {
			if f.Assertions < nasalint.MinAssertAvg {
				fmt.Fprintf(w, "%s:%d: info: func %s has %d assertion(s)\n", f.Path, f.Line, f.Name, f.Assertions)
			}
		}
	}
	for _, f := range res.Findings {
		fmt.Fprintln(w, f.String())
	}
	fmt.Fprintf(w, "nasa-lint: %d error(s); %d functions, %d assertions (average %.2f)\n",
		len(res.Findings), len(res.Funcs), res.Assertions, res.Average())
}

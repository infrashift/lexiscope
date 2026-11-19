// Command nasa-lint checks the Go source of the module in the current
// directory against the NASA Power of 10 rules (ADR-0003). It is the Go port
// of scripts/nasa-lint.py. Package internal/nasalint lists the checks.
//
// Usage: go run ./cmd/nasa-lint [--verbose]
//
// With --verbose, it lists each function with fewer than 2 assertions. The
// exit status is 0 without findings, 1 with findings and 2 on a failure.
package main

import (
	"io"
	"os"

	"github.com/infrashift/lexiscope/internal/assert"
	"github.com/infrashift/lexiscope/internal/nasalint"
	"github.com/infrashift/lexiscope/internal/output"
)

// exitFailure is the exit status of a failure that is not a finding.
const exitFailure = 2

func main() {
	assert.True(len(os.Args) > 0, "the program has a name")
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run returns the exit status. It writes a failure to stderr.
func run(args []string, stdout, stderr io.Writer) int {
	assert.True(stdout != nil && stderr != nil, "the output streams exist")
	errOut := output.New(stderr)
	code, failure := lint(args, output.New(stdout))
	if failure != "" {
		errOut.Printf("nasa-lint: %s\n", failure)
	}
	assert.True((code == exitFailure) == (failure != ""), "each failure has a message")
	return code
}

func lint(args []string, out *output.Writer) (code int, failure string) {
	assert.True(out != nil, "the writer exists")
	assert.True(len(args) < 1<<16, "the argument count is bounded")
	defer assert.Recover(&code, &failure)
	verbose := false
	for _, a := range args {
		if a != "--verbose" && a != "-v" {
			return exitFailure, "unknown argument " + a + "\nusage: nasa-lint [--verbose]"
		}
		verbose = true
	}
	root, err := os.Getwd()
	if err != nil {
		return exitFailure, err.Error()
	}
	res, err := nasalint.Run(root)
	if err != nil {
		return exitFailure, err.Error()
	}
	write(out, res, verbose)
	if err := out.Err(); err != nil {
		return exitFailure, err.Error()
	}
	if len(res.Findings) > 0 {
		return 1, ""
	}
	return 0, ""
}

// write prints the findings and the summary in the format of
// scripts/nasa-lint.py.
func write(out *output.Writer, res *nasalint.Result, verbose bool) {
	assert.True(out != nil, "the writer exists")
	assert.True(res != nil, "the result exists")
	if verbose {
		for _, f := range res.Funcs {
			if f.Assertions < nasalint.MinAssertAvg {
				out.Printf("%s:%d: info: func %s has %d assertion(s)\n", f.Path, f.Line, f.Name, f.Assertions)
			}
		}
	}
	for _, f := range res.Findings {
		out.Printf("%s\n", f.String())
	}
	out.Printf("nasa-lint: %d error(s); %d functions, %d assertions (average %.2f)\n",
		len(res.Findings), len(res.Funcs), res.Assertions, res.Average())
}

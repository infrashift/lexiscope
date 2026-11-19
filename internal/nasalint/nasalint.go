// Package nasalint checks the Go source of a module against the NASA Power of
// 10 rules, as ADR-0003 of the lexiscope wiki adapts them. It is the Go port of
// scripts/nasa-lint.py, which reads Zig only.
//
// The checks:
//
//	R1  no direct or indirect recursion in one package, and no goto
//	R2  no loop without a condition (for {})
//	R4  functions of 60 lines or fewer, from the func keyword to the last brace
//	R5  an average of 2 or more assert calls per function
//	R7  no discarded value (_ = x), and no discarded call result (x, _ := f())
//	R9  no function types, no function literals and no function values
//
// Test files (_test.go) are exempt. The functions of the assert package do
// not count for R5. Suppress one finding with a comment on the same line or
// the line above: // nasa-lint: allow(R7) <reason>.
//
// An import cycle is a compile error in Go, so a recursion across packages
// needs a function value, and R9 reports each function value. So the call graph
// of R1 covers one package at a time.
package nasalint

import (
	"fmt"
	"slices"
	"strings"

	"github.com/infrashift/lexiscope/internal/assert"
)

// Limits of one run. A run that exceeds a limit stops with an error.
const (
	MaxFnLines      = 60
	MinAssertAvg    = 2.0
	MaxDirs         = 4096
	MaxFiles        = 8192
	MaxNodesPerFile = 1 << 20
	MaxFuncs        = 1 << 16
)

// AssertPathSuffix selects the assert package by the end of its import path.
const AssertPathSuffix = "/internal/assert"

// Finding is one rule violation. Line is 0 for a finding about the whole
// module.
type Finding struct {
	Path    string
	Line    int
	Rule    string
	Message string
}

// String formats the finding as path:line: rule: message.
func (f Finding) String() string {
	assert.True(f.Rule != "", "a finding has a rule")
	assert.True(f.Path != "", "a finding has a path")
	if f.Line == 0 {
		return fmt.Sprintf("%s: %s: %s", f.Path, f.Rule, f.Message)
	}
	return fmt.Sprintf("%s:%d: %s: %s", f.Path, f.Line, f.Rule, f.Message)
}

// FuncStat is the assertion count of one function.
type FuncStat struct {
	Path       string
	Line       int
	Name       string
	Assertions int
}

// Result is the outcome of a run.
type Result struct {
	Findings   []Finding
	Funcs      []FuncStat
	Assertions int
}

// Average returns the assertions per function, or 0 without functions.
func (r *Result) Average() float64 {
	assert.True(r != nil, "the result exists")
	assert.True(r.Assertions >= 0, "the assertion count is not negative")
	if len(r.Funcs) == 0 {
		return 0
	}
	return float64(r.Assertions) / float64(len(r.Funcs))
}

// Run checks each package of the module in root, without the test files.
func Run(root string) (*Result, error) {
	assert.True(root != "", "the root is set")
	module, err := modulePath(root)
	if err != nil {
		return nil, err
	}
	dirs, err := packageDirs(root)
	if err != nil {
		return nil, err
	}
	assert.True(len(dirs) <= MaxDirs, "the walk keeps the directory limit")
	res := &Result{}
	files := 0
	for _, dir := range dirs {
		n, err := checkPackage(root, module, dir, res)
		if err != nil {
			return nil, err
		}
		files += n
		if files > MaxFiles {
			return nil, fmt.Errorf("more than %d Go files", MaxFiles)
		}
	}
	finish(res)
	return res, nil
}

// finish sorts the findings and adds the R5 finding of the module last, as
// scripts/nasa-lint.py does.
func finish(res *Result) {
	assert.True(res != nil, "the result exists")
	assert.True(len(res.Funcs) <= MaxFuncs*MaxFiles, "the function count is bounded")
	slices.SortStableFunc(res.Findings, compareFindings) // nasa-lint: allow(R9) slices takes a comparison function.
	if avg := res.Average(); len(res.Funcs) > 0 && avg < MinAssertAvg {
		res.Findings = append(res.Findings, Finding{
			Path: ".", Rule: "R5",
			Message: fmt.Sprintf("%d assertions in %d functions (average %.2f, min %.1f)",
				res.Assertions, len(res.Funcs), avg, MinAssertAvg),
		})
	}
}

func compareFindings(a, b Finding) int {
	assert.True(a.Path != "" && b.Path != "", "findings have paths")
	assert.True(a.Line >= 0 && b.Line >= 0, "lines are not negative")
	if c := strings.Compare(a.Path, b.Path); c != 0 {
		return c
	}
	return a.Line - b.Line
}

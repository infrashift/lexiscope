package cli

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/ruleset"
	"github.com/infrashift/lexis/internal/version"
)

// Output formats of lexis scan.
var formats = []string{"text", "json", "sarif"}

// Severities of --fail-on, from the most severe.
var severities = []string{"error", "warning", "note"}

// ScanOptions is the checked command line of lexis scan.
type ScanOptions struct {
	Paths        []string
	Ruleset      ruleset.Options
	Format       string
	SARIFVersion string
	Output       string   // "-" is stdout
	Excludes     []string // path.Match patterns of paths to skip
	FailOn       string   // the least severe finding that gives exit status 1
}

// scanFlags holds the raw flag values of lexis scan.
type scanFlags struct {
	rules                                rulesetFlags
	format, sarifVersion, output, failOn string
	excludes                             []string
}

func newScanCmd() *cobra.Command {
	f := &scanFlags{}
	cmd := &cobra.Command{
		Use:   "scan [paths...]",
		Short: "Check files and directories against the active rules",
		Long: "lexis scan checks Markdown files and the comments of Go, Zig, Python, C and Rust files.\n" +
			"Without paths, it checks the current directory.",
		RunE: f.run, // nasa-lint: allow(R9) cobra calls the command through RunE.
	}
	f.register(cmd.Flags())
	return cmd
}

// register adds the flags of lexis scan to fl.
func (f *scanFlags) register(fl *pflag.FlagSet) {
	assert.True(f != nil, "the flags exist")
	assert.True(fl != nil, "the flag set exists")
	f.rules.register(fl)
	fl.StringVar(&f.format, "format", "text", "report format: text, json or sarif")
	fl.StringVar(&f.sarifVersion, "sarif-version", version.DefaultSARIFVersion, "SARIF version: 2.1.0 or 2.2")
	fl.StringVarP(&f.output, "output", "o", "-", "report file; - is stdout")
	fl.StringArrayVar(&f.excludes, "exclude", nil, "skip the paths that match this pattern")
	fl.StringVar(&f.failOn, "fail-on", "error", "exit with status 1 on a finding of this severity or higher: error, warning or note")
}

// options checks the flag values and returns the scan options.
func (f *scanFlags) options(sarifVersionSet bool, args []string) (ScanOptions, error) {
	assert.True(f != nil, "the flags exist")
	assert.True(len(args) <= maxArgs, "the command line is bounded")
	o := ScanOptions{Paths: args, Format: f.format, SARIFVersion: f.sarifVersion, Output: f.output, Excludes: f.excludes, FailOn: f.failOn}
	if len(o.Paths) == 0 {
		o.Paths = []string{"."}
	}
	if err := checkFormat(o.Format, o.SARIFVersion, sarifVersionSet); err != nil {
		return ScanOptions{}, err
	}
	rs, err := f.rules.options()
	if err != nil {
		return ScanOptions{}, err
	}
	o.Ruleset = rs
	if !slices.Contains(severities, o.FailOn) {
		return ScanOptions{}, fmt.Errorf("--fail-on: %q is not one of %s", o.FailOn, strings.Join(severities, ", "))
	}
	if o.Output == "" {
		return ScanOptions{}, errors.New("--output: the value is empty; use - for stdout")
	}
	return o, nil
}

func checkFormat(format, sarifVersion string, sarifVersionSet bool) error {
	assert.True(len(formats) == 3, "lexis has three report formats")
	assert.True(sarifVersion != "" || sarifVersionSet, "the SARIF version has a default")
	if !slices.Contains(formats, format) {
		return fmt.Errorf("--format: %q is not one of %s", format, strings.Join(formats, ", "))
	}
	if !slices.Contains(version.SARIFVersions(), sarifVersion) {
		return fmt.Errorf("--sarif-version: %q is not one of %s", sarifVersion, strings.Join(version.SARIFVersions(), ", "))
	}
	if sarifVersionSet && format != "sarif" {
		return errors.New("--sarif-version needs --format sarif")
	}
	return nil
}

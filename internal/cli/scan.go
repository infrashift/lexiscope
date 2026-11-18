package cli

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/version"
)

// Output formats of lexis scan.
var formats = []string{"text", "json", "sarif"}

// maxListItems bounds each list flag.
const maxListItems = 1 << 12

var (
	ruleIDRE = regexp.MustCompile(`^[A-Z][A-Z0-9]*(-[A-Z0-9]+)*$`)
	paramRE  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// errNotAvailable marks a command that a later phase of PLAN.md delivers.
var errNotAvailable = errors.New("not available yet")

// Override is one --set flag: rule_id.param=value.
type Override struct {
	RuleID string
	Param  string
	Value  string
}

// ScanOptions is the checked command line of lexis scan.
type ScanOptions struct {
	Paths        []string
	Rulesets     []string
	Enable       []string
	Omit         []string
	Overrides    []Override
	AllowTerms   []string
	Format       string
	SARIFVersion string
	Output       string // "-" is stdout
}

// scanFlags holds the raw flag values of lexis scan.
type scanFlags struct {
	rulesets, enable, omit, sets, allowTerms []string
	format, sarifVersion, output             string
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
	fl := cmd.Flags()
	fl.StringSliceVarP(&f.rulesets, "ruleset", "r", []string{"ste-100"}, "built-in ruleset name, or CUE file or directory")
	fl.StringSliceVar(&f.enable, "enable-rule", nil, "turn on these rule IDs")
	fl.StringSliceVar(&f.omit, "omit-rule", nil, "turn off these rule IDs (alias: --disable-rule)")
	fl.StringArrayVar(&f.sets, "set", nil, "change a rule parameter: rule_id.param=value")
	fl.StringSliceVar(&f.allowTerms, "allow-term", nil, "add approved technical terms")
	fl.StringVar(&f.format, "format", "text", "report format: text, json or sarif")
	fl.StringVar(&f.sarifVersion, "sarif-version", version.DefaultSARIFVersion, "SARIF version: 2.1.0 or 2.2")
	fl.StringVarP(&f.output, "output", "o", "-", "report file; - is stdout")
	fl.SetNormalizeFunc(normalizeScanFlag) // nasa-lint: allow(R9) pflag maps flag names through this function.
	assert.True(fl.Lookup("omit-rule") != nil, "the omit flag exists")
	return cmd
}

// normalizeScanFlag makes --disable-rule an alias of --omit-rule.
func normalizeScanFlag(fs *pflag.FlagSet, name string) pflag.NormalizedName {
	assert.True(fs != nil, "pflag gives the flag set")
	assert.True(name != "", "a flag has a name")
	if name == "disable-rule" {
		name = "omit-rule"
	}
	return pflag.NormalizedName(name)
}

func (f *scanFlags) run(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(f != nil, "the flags exist")
	opts, err := f.options(cmd.Flags().Changed("sarif-version"), args)
	if err != nil {
		return err
	}
	return fmt.Errorf("scan %s: %w: the rule engine arrives in Phase 4 of PLAN.md",
		strings.Join(opts.Paths, " "), errNotAvailable)
}

// options checks the flag values and returns the scan options.
func (f *scanFlags) options(sarifVersionSet bool, args []string) (ScanOptions, error) {
	assert.True(f != nil, "the flags exist")
	assert.True(len(args) <= maxArgs, "the command line is bounded")
	o := ScanOptions{Paths: args, Rulesets: f.rulesets, Enable: f.enable, Omit: f.omit,
		AllowTerms: f.allowTerms, Format: f.format, SARIFVersion: f.sarifVersion, Output: f.output}
	if len(o.Paths) == 0 {
		o.Paths = []string{"."}
	}
	if err := checkFormat(o.Format, o.SARIFVersion, sarifVersionSet); err != nil {
		return ScanOptions{}, err
	}
	if err := checkLists(o); err != nil {
		return ScanOptions{}, err
	}
	for _, s := range f.sets {
		ov, err := parseOverride(s)
		if err != nil {
			return ScanOptions{}, err
		}
		o.Overrides = append(o.Overrides, ov)
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

// checkLists checks the rulesets, the rule IDs and the terms.
func checkLists(o ScanOptions) error {
	assert.True(len(o.Paths) > 0, "the paths have a default")
	assert.True(len(o.Enable) <= maxListItems && len(o.Omit) <= maxListItems, "the rule lists are bounded")
	if len(o.Rulesets) == 0 || slices.Contains(o.Rulesets, "") {
		return errors.New("--ruleset: give a ruleset name or a CUE path")
	}
	for _, id := range slices.Concat(o.Enable, o.Omit) {
		if !ruleIDRE.MatchString(id) {
			return fmt.Errorf("%q is not a rule ID such as STE-100-01", id)
		}
	}
	for _, id := range o.Enable {
		if slices.Contains(o.Omit, id) {
			return fmt.Errorf("%s: --enable-rule and --omit-rule name the same rule", id)
		}
	}
	for _, term := range o.AllowTerms {
		if strings.TrimSpace(term) == "" {
			return errors.New("--allow-term: a term is empty")
		}
	}
	return nil
}

// parseOverride reads rule_id.param=value. The value stays text; the
// ruleset engine checks its type.
func parseOverride(s string) (Override, error) {
	assert.True(len(s) < 1<<16, "the flag value is bounded")
	key, value, ok := strings.Cut(s, "=")
	if !ok {
		return Override{}, fmt.Errorf("--set %q: use rule_id.param=value", s)
	}
	id, param, ok := strings.Cut(key, ".")
	if !ok || !ruleIDRE.MatchString(id) || !paramRE.MatchString(param) {
		return Override{}, fmt.Errorf("--set %q: use rule_id.param=value, for example STE-100-01.max_words=18", s)
	}
	ov := Override{RuleID: id, Param: param, Value: value}
	assert.True(ov.RuleID != "" && ov.Param != "", "the override names a rule and a parameter")
	return ov, nil
}

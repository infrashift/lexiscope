package cli

import (
	"bytes"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/infrashift/lexis/internal/version"
)

func runLexis(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Main(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

var versionLine = regexp.MustCompile(`^lexis version \S+ \(commit: \S+, built: \S+\)\n$`)

func TestVersion(t *testing.T) {
	for _, args := range [][]string{{"--version"}, {"-v"}, {"version"}} {
		code, out, errOut := runLexis(t, args...)
		if code != ExitOK || errOut != "" || !versionLine.MatchString(out) {
			t.Errorf("lexis %v = %d, %q, %q", args, code, out, errOut)
		}
	}
	t.Cleanup(func() { version.Version, version.Commit, version.BuildDate = "", "", "" })
	version.Version, version.Commit, version.BuildDate = "v1.0.0", "abc1234", "2026-11-18T00:00:00Z"
	want := "lexis version v1.0.0 (commit: abc1234, built: 2026-11-18T00:00:00Z)\n"
	for _, args := range [][]string{{"--version"}, {"version"}} {
		if _, out, _ := runLexis(t, args...); out != want {
			t.Errorf("lexis %v = %q, want %q", args, out, want)
		}
	}
}

func TestUsageErrors(t *testing.T) {
	cases := []struct {
		args []string
		msg  string
	}{
		{[]string{"bogus"}, `unknown command "bogus"`},
		{[]string{"version", "x"}, "unknown command"},
		{[]string{"scan", "--bogus"}, "unknown flag: --bogus"},
		{[]string{"scan", "--format", "xml"}, `--format: "xml" is not one of text, json, sarif`},
		{[]string{"scan", "--format", "sarif", "--sarif-version", "3.0"}, `--sarif-version: "3.0" is not one of 2.1.0, 2.2`},
		{[]string{"scan", "--sarif-version", "2.2"}, "--sarif-version needs --format sarif"},
		{[]string{"scan", "--enable-rule", "ste-100-01"}, `"ste-100-01" is not a rule ID`},
		{[]string{"scan", "--enable-rule", "STE-100-05", "--disable-rule", "STE-100-05"}, "name the same rule"},
		{[]string{"scan", "--set", "STE-100-01.max_words"}, "use rule_id.param=value"},
		{[]string{"scan", "--set", "STE-100-01=18"}, "use rule_id.param=value"},
		{[]string{"scan", "--set", "STE-100-01.Max=18"}, "use rule_id.param=value"},
		{[]string{"scan", "--allow-term", " "}, "--allow-term: a term is empty"},
		{[]string{"scan", "--ruleset", ""}, "--ruleset: give a ruleset name"},
		{[]string{"scan", "--output", ""}, "--output: the value is empty"},
		{[]string{"rules", "inspect"}, "accepts 1 arg(s)"},
	}
	for _, c := range cases {
		code, out, errOut := runLexis(t, c.args...)
		if code != ExitFailure || out != "" || !strings.HasPrefix(errOut, "lexis: ") || !strings.Contains(errOut, c.msg) {
			t.Errorf("lexis %v = %d, %q, %q; want status 2 and %q", c.args, code, out, errOut, c.msg)
		}
	}
}

func TestLaterPhases(t *testing.T) {
	cases := map[string][]string{
		"Phase 4": {"scan", "--format", "sarif", "--sarif-version", "2.2", "--set", "STE-100-01.max_words=18", "docs"},
		"Phase 3": {"rules", "list"},
	}
	for phase, args := range cases {
		code, _, errOut := runLexis(t, args...)
		if code != ExitFailure || !strings.Contains(errOut, "not available yet") || !strings.Contains(errOut, phase) {
			t.Errorf("lexis %v = %d, %q", args, code, errOut)
		}
	}
	if code, _, errOut := runLexis(t, "rules", "inspect", "ste-100"); code != ExitFailure || !strings.Contains(errOut, "Phase 3") {
		t.Errorf("rules inspect = %d, %q", code, errOut)
	}
}

func TestScanOptions(t *testing.T) {
	cmd := newScanCmd()
	err := cmd.Flags().Parse([]string{
		"-r", "ste-100,my.cue", "--enable-rule", "STE-100-05", "--disable-rule", "STE-100-07,STE-100-08",
		"--omit-rule", "DOC-CUSTOM-01", "--set", "STE-100-01.max_words=18", "--set", "STE-100-05.swap=a,b",
		"--allow-term", "okfw,zmake", "--format", "sarif", "--sarif-version", "2.2", "-o", "out.sarif",
	})
	if err != nil {
		t.Fatal(err)
	}
	f := scanFlagsOf(t, cmd)
	got, err := f.options(cmd.Flags().Changed("sarif-version"), []string{"docs", "README.md"})
	if err != nil {
		t.Fatal(err)
	}
	want := ScanOptions{
		Paths:    []string{"docs", "README.md"},
		Rulesets: []string{"ste-100", "my.cue"},
		Enable:   []string{"STE-100-05"},
		Omit:     []string{"STE-100-07", "STE-100-08", "DOC-CUSTOM-01"},
		Overrides: []Override{
			{RuleID: "STE-100-01", Param: "max_words", Value: "18"},
			{RuleID: "STE-100-05", Param: "swap", Value: "a,b"},
		},
		AllowTerms:   []string{"okfw", "zmake"},
		Format:       "sarif",
		SARIFVersion: "2.2",
		Output:       "out.sarif",
	}
	if !equalOptions(got, want) {
		t.Fatalf("options =\n%+v\nwant\n%+v", got, want)
	}
}

func TestScanDefaults(t *testing.T) {
	cmd := newScanCmd()
	if err := cmd.Flags().Parse(nil); err != nil {
		t.Fatal(err)
	}
	got, err := scanFlagsOf(t, cmd).options(false, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := ScanOptions{Paths: []string{"."}, Rulesets: []string{"ste-100"}, Format: "text", SARIFVersion: "2.1.0", Output: "-"}
	if !equalOptions(got, want) {
		t.Fatalf("defaults =\n%+v\nwant\n%+v", got, want)
	}
}

// scanFlagsOf returns the scanFlags behind the RunE method value of cmd.
func scanFlagsOf(t *testing.T, cmd *cobra.Command) *scanFlags {
	t.Helper()
	f := &scanFlags{}
	get := func(name string) []string {
		v, err := cmd.Flags().GetStringSlice(name)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	f.rulesets, f.enable, f.omit, f.allowTerms = get("ruleset"), get("enable-rule"), get("omit-rule"), get("allow-term")
	sets, err := cmd.Flags().GetStringArray("set")
	if err != nil {
		t.Fatal(err)
	}
	f.sets = sets
	f.format, _ = cmd.Flags().GetString("format")
	f.sarifVersion, _ = cmd.Flags().GetString("sarif-version")
	f.output, _ = cmd.Flags().GetString("output")
	return f
}

func equalOptions(a, b ScanOptions) bool {
	return slices.Equal(a.Paths, b.Paths) && slices.Equal(a.Rulesets, b.Rulesets) &&
		slices.Equal(a.Enable, b.Enable) && slices.Equal(a.Omit, b.Omit) &&
		slices.Equal(a.Overrides, b.Overrides) && slices.Equal(a.AllowTerms, b.AllowTerms) &&
		a.Format == b.Format && a.SARIFVersion == b.SARIFVersion && a.Output == b.Output
}

// TestDocsListEachCommandAndFlag keeps the CLI reference page in step with
// the command tree.
func TestDocsListEachCommandAndFlag(t *testing.T) {
	page, err := os.ReadFile("../../docs/src/content/docs/reference/cli.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(page)
	root := newRoot(&bytes.Buffer{}, &bytes.Buffer{})
	root.InitDefaultHelpCmd()
	stack := []*cobra.Command{root}
	for len(stack) > 0 {
		cmd := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		stack = append(stack, cmd.Commands()...)
		if cmd.Name() == "help" {
			continue
		}
		if !strings.Contains(text, "## `"+cmd.CommandPath()) {
			t.Errorf("cli.md has no section for %q", cmd.CommandPath())
		}
		cmd.LocalFlags().VisitAll(func(fl *pflag.Flag) {
			if fl.Name != "help" && !strings.Contains(text, "`--"+fl.Name+"") {
				t.Errorf("cli.md does not list --%s of %q", fl.Name, cmd.CommandPath())
			}
		})
	}
}

package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/infrashift/lexis/internal/ruleset"
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

func TestScan(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("bad.md", "# Title\n\nIt will work. Processing is slow.\n")
	write("ok.md", "The tool reads the file.\n")
	write("skip.md", "This should not count.\n")
	t.Chdir(dir)

	code, out, errOut := runLexis(t, "scan", "--exclude", "skip.md")
	want := "bad.md:3:4: error: STE-100-04: Do not use \"will\".\n" +
		"bad.md:3:15: warning: STE-100-07: \"Processing\" is an -ing form. Use a verb or a noun.\n" +
		"lexis: 1 error(s), 1 warning(s), 0 note(s) in 2 file(s), ruleset ste-100\n"
	if code != ExitFindings || out != want || errOut != "" {
		t.Fatalf("scan = %d\n%s\nwant\n%s\nstderr %q", code, out, want, errOut)
	}
	if code, out, _ := runLexis(t, "scan", "ok.md"); code != ExitOK || !strings.HasPrefix(out, "lexis: 0 error(s)") {
		t.Fatalf("clean scan = %d, %q", code, out)
	}
	if code, _, _ := runLexis(t, "scan", "--omit-rule", "STE-100-04", "bad.md"); code != ExitOK {
		t.Fatalf("a scan with warnings only = %d, want 0", code)
	}
	code, _, errOut = runLexis(t, "scan", "--format", "json", "-o", "report.json", "bad.md")
	data, err := os.ReadFile("report.json")
	if code != ExitFindings || errOut != "" || err != nil {
		t.Fatalf("json scan = %d, %q, %v", code, errOut, err)
	}
	var doc struct {
		Summary  struct{ Errors, Warnings, Files int }
		Findings []struct {
			RuleID string `json:"rule_id"`
			Line   int
			Column int
		}
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Summary.Errors != 1 || doc.Summary.Files != 1 || len(doc.Findings) != 2 || doc.Findings[0].RuleID != "STE-100-04" || doc.Findings[0].Column != 4 {
		t.Fatalf("json report = %+v", doc)
	}
	code, _, errOut = runLexis(t, "scan", "--format", "sarif")
	if code != ExitFailure || !strings.Contains(errOut, "SARIF arrives in Phase 5") {
		t.Fatalf("sarif scan = %d, %q", code, errOut)
	}
	if code, _, errOut := runLexis(t, "scan", "missing.md"); code != ExitFailure || !strings.Contains(errOut, "missing.md") {
		t.Fatalf("missing file = %d, %q", code, errOut)
	}
}

// parseScan parses args with the flags of lexis scan.
func parseScan(t *testing.T, args ...string) (*scanFlags, *pflag.FlagSet) {
	t.Helper()
	f := &scanFlags{}
	fl := pflag.NewFlagSet("scan", pflag.ContinueOnError)
	f.register(fl)
	if err := fl.Parse(args); err != nil {
		t.Fatal(err)
	}
	return f, fl
}

func TestScanOptions(t *testing.T) {
	f, fl := parseScan(t,
		"-r", "ste-100,my.cue", "--enable-rule", "STE-100-05", "--disable-rule", "STE-100-07,STE-100-08",
		"--omit-rule", "DOC-CUSTOM-01", "--set", "STE-100-01.max_words=18", "--set", "STE-100-05.swap=a,b",
		"--allow-term", "okfw,zmake", "--format", "sarif", "--sarif-version", "2.2", "-o", "out.sarif",
		"--exclude", "a/*", "--exclude", "b.md", "docs", "README.md")
	got, err := f.options(fl.Changed("sarif-version"), fl.Args())
	if err != nil {
		t.Fatal(err)
	}
	want := ScanOptions{
		Paths: []string{"docs", "README.md"},
		Ruleset: ruleset.Options{
			Rulesets: []string{"ste-100", "my.cue"},
			Enable:   []string{"STE-100-05"},
			Omit:     []string{"STE-100-07", "STE-100-08", "DOC-CUSTOM-01"},
			Overrides: []ruleset.Override{
				{RuleID: "STE-100-01", Param: "max_words", Value: "18"},
				{RuleID: "STE-100-05", Param: "swap", Value: "a,b"},
			},
			AllowTerms: []string{"okfw", "zmake"},
		},
		Format:       "sarif",
		SARIFVersion: "2.2",
		Output:       "out.sarif",
		Excludes:     []string{"a/*", "b.md"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options =\n%+v\nwant\n%+v", got, want)
	}
}

func TestScanDefaults(t *testing.T) {
	f, fl := parseScan(t)
	got, err := f.options(fl.Changed("sarif-version"), fl.Args())
	if err != nil {
		t.Fatal(err)
	}
	want := ScanOptions{Paths: []string{"."}, Ruleset: ruleset.Options{Rulesets: []string{"ste-100"}},
		Format: "text", SARIFVersion: "2.1.0", Output: "-"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("defaults =\n%+v\nwant\n%+v", got, want)
	}
}

func TestRulesList(t *testing.T) {
	code, out, errOut := runLexis(t, "rules", "list", "--omit-rule", "STE-100-07", "--set", "STE-100-01.max_words=18",
		"-r", "ste-100,../ruleset/testdata/add_rule.cue")
	if code != ExitOK || errOut != "" {
		t.Fatalf("rules list = %d, %q", code, errOut)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if lines[0] != "ruleset ste-100: 12 rules, 11 active, 1 omitted" || len(lines) != 14 {
		t.Fatalf("output:\n%s", out)
	}
	for _, want := range []string{
		`^ID +STATUS +SEVERITY +CHECK +LIMIT +TITLE$`,
		`^DOC-CUSTOM-01 +active +error +regex_ban +- +Do not write TODO`,
		`^STE-100-01 +active +error +max_words +18 +Limit the length of a procedural step$`,
		`^STE-100-07 +omitted +warning +ambiguous_gerunds +- `,
	} {
		if !slices.ContainsFunc(lines, regexp.MustCompile(want).MatchString) {
			t.Errorf("no line matches %s in:\n%s", want, out)
		}
	}
}

func TestRulesInspect(t *testing.T) {
	code, out, errOut := runLexis(t, "rules", "inspect", "ste-100")
	if code != ExitOK || errOut != "" {
		t.Fatalf("rules inspect = %d, %q", code, errOut)
	}
	for _, want := range []string{
		"ruleset ste-100: A subset of ASD-STE100",
		"\nSTE-100-04 (error, active)\n",
		"  tokens: are expected to, can be seen, easily, is expected to, just, please, really, should, simply, very, will, won't, would\n",
		"utilize -> use",
		"  verify: not validate, confirm, ensure; nouns verification, validation, confirmation\n",
		"\nvocabulary: 0 terms\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output has no %q", want)
		}
	}
	code, _, errOut = runLexis(t, "rules", "inspect", "no-such-ruleset")
	if code != ExitFailure || !strings.Contains(errOut, "not a built-in ruleset") {
		t.Fatalf("unknown ruleset = %d, %q", code, errOut)
	}
}

func TestRulesetErrorsReachTheCLI(t *testing.T) {
	code, _, errOut := runLexis(t, "scan", "--set", "STE-100-01.max_words=many", "../../PLAN.md")
	if code != ExitFailure || !strings.Contains(errOut, "max_words: conflicting values") {
		t.Fatalf("scan with a bad --set = %d, %q", code, errOut)
	}
	code, _, errOut = runLexis(t, "rules", "list", "--omit-rule", "NO-SUCH-01")
	if code != ExitFailure || !strings.Contains(errOut, "no rule NO-SUCH-01") {
		t.Fatalf("rules list with an unknown rule = %d, %q", code, errOut)
	}
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

package engine

// The parity test of ADR-0005. lexiscope and scripts/ste-check.py (the oracle)
// report the same core findings (STE-100-01 to 04) in both directions. The
// test reads each fixture and each document of this repository. The oracle
// reports the first line of a paragraph, so the test compares the block
// line of each lexiscope finding.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/infrashift/lexiscope/internal/ruleset"
)

var (
	oracleLine  = regexp.MustCompile(`^(.+?):(\d+): (error|warning): (.*)$`)
	oracleWords = regexp.MustCompile(`^(procedural step|sentence) has (\d+) words \(max \d+\): `)
	oracleTerm  = regexp.MustCompile(`^(banned word or phrase|possible passive voice): (.*)$`)
)

// repoRoot is the root of the lexiscope repository.
const repoRoot = "../.."

// oracleScript returns the absolute path of the oracle.
func oracleScript(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(repoRoot, "scripts", "ste-check.py"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// runOracle runs a copy of the oracle in dir and returns its findings as
// keys: "path:line: RULE: detail".
func runOracle(t *testing.T, script, dir string) []string {
	t.Helper()
	cmd := exec.Command("python3", "-I", script)
	cmd.Dir = dir
	out, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
		t.Fatalf("ste-check.py: %v", err)
	}
	var keys []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		m := oracleLine.FindStringSubmatch(line)
		if m == nil {
			continue // the summary line
		}
		keys = append(keys, oracleKey(t, m[1], m[2], m[4]))
	}
	slices.Sort(keys)
	return keys
}

func oracleKey(t *testing.T, path, line, msg string) string {
	t.Helper()
	if m := oracleWords.FindStringSubmatch(msg); m != nil {
		rule := map[string]string{"procedural step": "STE-100-01", "sentence": "STE-100-02"}[m[1]]
		return fmt.Sprintf("%s:%s: %s: %s", path, line, rule, m[2])
	}
	if m := oracleTerm.FindStringSubmatch(msg); m != nil {
		rule := map[string]string{"banned word or phrase": "STE-100-04", "possible passive voice": "STE-100-03"}[m[1]]
		return fmt.Sprintf("%s:%s: %s: %s", path, line, rule, m[2])
	}
	t.Fatalf("unknown oracle message: %s", msg)
	return ""
}

// coreEngine returns an engine with the core rules only, after the given
// changes.
func coreEngine(t *testing.T, o ruleset.Options) *Engine {
	t.Helper()
	o.Rulesets = []string{"ste-100"}
	for i := 5; i <= 11; i++ {
		o.Omit = append(o.Omit, fmt.Sprintf("STE-100-%02d", i))
	}
	rs, err := ruleset.Load(o)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(rs)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// lexisKeys checks the files (relative to dir) and returns the keys of the
// findings in the format of oracleKey.
func lexisKeys(t *testing.T, e *Engine, dir string, files []string) []string {
	t.Helper()
	var keys []string
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			t.Fatal(err)
		}
		fs, err := e.Check(rel, src)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range fs {
			detail := f.Detail
			if f.RuleID == "STE-100-03" || f.RuleID == "STE-100-04" {
				detail = pyRepr(detail)
			}
			keys = append(keys, fmt.Sprintf("%s:%d: %s: %s", rel, f.BlockLine, f.RuleID, detail))
		}
	}
	slices.Sort(keys)
	return keys
}

// pyRepr formats s as Python's repr of a str.
func pyRepr(s string) string {
	quote := byte('\'')
	if strings.Contains(s, "'") && !strings.Contains(s, `"`) {
		quote = '"'
	}
	var b strings.Builder
	b.WriteByte(quote)
	for _, r := range s {
		switch {
		case r == '\\' || r == rune(quote):
			b.WriteString(`\` + string(r))
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == ' ' || (unicode.IsPrint(r) && !unicode.In(r, unicode.Z)):
			b.WriteRune(r)
		case r < 0x100:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r < 0x10000:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// stage copies the fixtures of one testdata directory into a temporary
// docs/src/content/docs tree, where the oracle reads them. It returns the
// temporary root and the relative paths of the fixtures.
func stage(t *testing.T, dirs ...string) (string, []string) {
	t.Helper()
	root := t.TempDir()
	var files []string
	for _, dir := range dirs {
		src := filepath.Join("testdata", "ste100", dir)
		err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(src, p)
			if err != nil {
				return err
			}
			dst := filepath.Join("docs", "src", "content", "docs", dir, rel)
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Join(root, filepath.Dir(dst)), 0o755); err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(dst))
			return os.WriteFile(filepath.Join(root, dst), data, 0o644)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return root, files
}

func diff(t *testing.T, name string, oracle, lexis []string) {
	t.Helper()
	var missing, extra []string
	for _, k := range oracle {
		if i := slices.Index(lexis, k); i >= 0 {
			lexis = slices.Delete(lexis, i, i+1)
		} else {
			missing = append(missing, k)
		}
	}
	extra = lexis
	if len(missing)+len(extra) > 0 {
		t.Errorf("%s: lexiscope differs from ste-check.py\nonly the oracle:\n  %s\nonly lexiscope:\n  %s",
			name, strings.Join(missing, "\n  "), strings.Join(extra, "\n  "))
	}
}

func TestParityFixtures(t *testing.T) {
	root, files := stage(t, "violations", "compliant")
	oracle := runOracle(t, oracleScript(t), root)
	if len(oracle) < 40 {
		t.Fatalf("the oracle reports %d findings for the fixtures; want at least 40", len(oracle))
	}
	diff(t, "fixtures", oracle, lexisKeys(t, coreEngine(t, ruleset.Options{}), root, files))
}

// TestParityRepository compares each document that the oracle reads in this
// repository.
func TestParityRepository(t *testing.T) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	files := oracleFiles(t, root)
	if len(files) < 20 {
		t.Fatalf("found %d documents; want at least 20", len(files))
	}
	diff(t, "repository", runOracle(t, oracleScript(t), root), lexisKeys(t, coreEngine(t, ruleset.Options{}), root, files))
}

// oracleFiles lists the files of files() in the oracle.
func oracleFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	for _, n := range []string{"README.md", "PLAN.md", "CLAUDE.md"} {
		if _, err := os.Stat(filepath.Join(root, n)); err == nil {
			files = append(files, n)
		}
	}
	walk := func(dir string, keep func(rel string) bool) {
		err := filepath.WalkDir(filepath.Join(root, dir), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(root, p)
			if err == nil && keep(filepath.ToSlash(rel)) {
				files = append(files, filepath.ToSlash(rel))
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	walk("docs/src/content/docs", func(rel string) bool { return strings.Contains(filepath.Base(rel), ".md") })
	walk(".okfw", func(rel string) bool {
		parts := strings.Split(strings.TrimPrefix(rel, ".okfw/"), "/")
		base := parts[len(parts)-1]
		return strings.HasSuffix(base, ".md") && parts[0] != "daily" && parts[0] != "ephemeral" && base != "index.md" && base != "log.md"
	})
	return files
}

// TestParityChangedRules changes the oracle and lexiscope in the same way: a
// word limit of 15 for steps, and no passive check.
func TestParityChangedRules(t *testing.T) {
	root, files := stage(t, "violations")
	src, err := os.ReadFile(oracleScript(t))
	if err != nil {
		t.Fatal(err)
	}
	patched := strings.Replace(string(src), "MAX_PROCEDURAL = 20", "MAX_PROCEDURAL = 15", 1)
	if patched == string(src) {
		t.Fatal("ste-check.py has no MAX_PROCEDURAL = 20")
	}
	script := filepath.Join(t.TempDir(), "ste-check.py")
	if err := os.WriteFile(script, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}
	var oracle []string
	for _, k := range runOracle(t, script, root) {
		if !strings.Contains(k, ": STE-100-03: ") {
			oracle = append(oracle, k)
		}
	}
	e := coreEngine(t, ruleset.Options{
		Omit:      []string{"STE-100-03"},
		Overrides: []ruleset.Override{{RuleID: "STE-100-01", Param: "max_words", Value: "15"}},
	})
	diff(t, "changed rules", oracle, lexisKeys(t, e, root, files))
}

// TestDivergences checks each deliberate difference (spec "Known differences
// from ste-check"): the oracle reports the fixture, and lexiscope does not.
func TestDivergences(t *testing.T) {
	root, files := stage(t, "divergent")
	oracle := runOracle(t, oracleScript(t), root)
	lexis := lexisKeys(t, coreEngine(t, ruleset.Options{}), root, files)
	for _, f := range files {
		var o []string
		for _, k := range oracle {
			if strings.HasPrefix(k, f+":") {
				o = append(o, k)
			}
		}
		if len(o) == 0 {
			t.Errorf("%s: the oracle reports nothing, so the fixture shows no difference", f)
		}
	}
	if len(lexis) != 0 {
		t.Errorf("lexiscope reports findings in the divergent fixtures: %q", lexis)
	}
}

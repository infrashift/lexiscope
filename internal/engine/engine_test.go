package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/infrashift/lexis/internal/ruleset"
)

func engineFor(t *testing.T, o ruleset.Options) *Engine {
	t.Helper()
	if len(o.Rulesets) == 0 {
		o.Rulesets = []string{"ste-100"}
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

// brief formats findings as "line:col RULE term>replacement".
func brief(fs []Finding) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = fmt.Sprintf("%d:%d %s %s", f.Line, f.Column, f.RuleID, f.Term)
		if f.Replacement != "" {
			out[i] += ">" + f.Replacement
		}
	}
	return out
}

func check(t *testing.T, e *Engine, path, src string) []string {
	t.Helper()
	fs, err := e.Check(path, []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	return brief(fs)
}

func TestExtendedRules(t *testing.T) {
	e := engineFor(t, ruleset.Options{})
	src := "Utilize the tool.\n\n" + // 05
		"The verification takes time and the usage is low.\n\n" + // 06
		"Processing the string is slow.\n\n" + // 07, "string" is an exception
		"Open your file.\n\n" + // 08
		"The tool is seamless.\n\n" + // 09
		"The code is written.\n\n" + // 10
		"1. Run the build and copy the file.\n" // 11
	want := []string{
		"1:1 STE-100-05 Utilize>use",
		"3:5 STE-100-06 verification>verify",
		"3:37 STE-100-06 usage>use",
		"5:1 STE-100-07 Processing",
		"7:6 STE-100-08 your",
		"9:13 STE-100-09 seamless",
		"11:10 STE-100-10 is written",
		"13:4 STE-100-11 ",
	}
	if got := check(t, e, "x.md", src); !slices.Equal(got, want) {
		t.Fatalf("findings:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestVocabularyAndExceptions(t *testing.T) {
	e := engineFor(t, ruleset.Options{AllowTerms: []string{"logging2", "Rendering"}})
	got := check(t, e, "x.md", "The rendering and the configuration and the string stay.\n")
	if len(got) != 0 {
		t.Fatalf("findings = %q, want none: vocabulary and exceptions", got)
	}
}

func TestPositions(t *testing.T) {
	e := engineFor(t, ruleset.Options{})
	// A UTF-16 column counts an astral character as 2 units.
	got := check(t, e, "x.md", "# Title\n\n\U0001F600 é you should go.\n")
	if want := []string{"3:10 STE-100-04 should"}; !slices.Equal(got, want) {
		t.Fatalf("findings = %q, want %q", got, want)
	}
	fs, err := e.Check("x.md", []byte("A line that\nwill wrap.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 || fs[0].Line != 2 || fs[0].Column != 1 || fs[0].EndLine != 2 || fs[0].EndColumn != 5 || fs[0].BlockLine != 1 {
		t.Fatalf("finding = %+v", fs)
	}
	if fs[0].Message != `Do not use "will".` {
		t.Fatalf("message = %q", fs[0].Message)
	}
}

func TestMessages(t *testing.T) {
	e := engineFor(t, ruleset.Options{})
	fs, err := e.Check("x.md", []byte("1. "+strings.Repeat("go ", 21)+"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 || fs[0].Message != "The step has 21 words. Use 20 words or fewer." || fs[0].Severity != "error" {
		t.Fatalf("findings = %+v", fs)
	}
}

func TestScopes(t *testing.T) {
	dir := t.TempDir()
	rule := `rules: "H-01": {severity: "note", scope: "%s", check: "regex_ban", title: "t", message: "found {term}", params: pattern: "XYZ"}`
	cases := map[string][]string{
		"heading":      {"1:3 H-01 XYZ"},
		"comment":      {"7:4 H-01 XYZ"},
		"any_sentence": {"3:1 H-01 XYZ", "7:4 H-01 XYZ"},
		"token":        {"1:3 H-01 XYZ", "3:1 H-01 XYZ", "7:4 H-01 XYZ"},
	}
	md := "# XYZ\n\nXYZ text.\n"
	goSrc := "package p\n\n// XYZ text.\n"
	for scope, want := range cases {
		p := filepath.Join(dir, scope+".cue")
		if err := os.WriteFile(p, []byte(fmt.Sprintf("name: \"s\"\n"+rule, scope)), 0o644); err != nil {
			t.Fatal(err)
		}
		e := engineFor(t, ruleset.Options{Rulesets: []string{p}})
		got := append(check(t, e, "x.md", md), check(t, e, "x.go", strings.Repeat("\n", 4)+goSrc)...)
		if !slices.Equal(got, want) {
			t.Errorf("scope %s: findings = %q, want %q", scope, got, want)
		}
	}
}

func TestBulletSteps(t *testing.T) {
	src := "- " + strings.Repeat("go ", 21) + "\n"
	if got := check(t, engineFor(t, ruleset.Options{}), "x.md", src); len(got) != 0 {
		t.Fatalf("a bullet is not a step by default: %q", got)
	}
	e := engineFor(t, ruleset.Options{Overrides: []ruleset.Override{
		{RuleID: "STE-100-01", Param: "procedural_markers", Value: `["numbered", "bullet"]`},
	}})
	if got := check(t, e, "x.md", src); !slices.Equal(got, []string{"1:3 STE-100-01 "}) {
		t.Fatalf("a bullet with the bullet marker: %q", got)
	}
}

func TestSourceComments(t *testing.T) {
	e := engineFor(t, ruleset.Options{})
	src := `package p

//go:generate stringer should not count
// The parser will read
// the file.
//
//	code will not count
func f() {
	x := 1 // A side note should count.
	_ = x  // nasa-lint: allow(R7) the test needs it.
}

/* A block comment
 * will count. */
`
	want := []string{"4:15 STE-100-04 will", "9:24 STE-100-04 should", "14:4 STE-100-04 will"}
	if got := check(t, e, "x.go", src); !slices.Equal(got, want) {
		t.Fatalf("findings = %q, want %q", got, want)
	}
	if got := check(t, e, "x.py", "# The script should run.\nx = '# not a comment, will'\n"); !slices.Equal(got, []string{"1:14 STE-100-04 should"}) {
		t.Fatalf("python findings = %q", got)
	}
	if got := check(t, e, "x.txt", "This should not count.\n"); len(got) != 0 {
		t.Fatalf("a text file has findings: %q", got)
	}
}

func TestBadInput(t *testing.T) {
	e := engineFor(t, ruleset.Options{})
	if _, err := e.Check("x.md", []byte{0xff, 0xfe}); err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("invalid UTF-8: err = %v", err)
	}
	o := ruleset.Options{Rulesets: []string{"ste-100"}, Overrides: []ruleset.Override{{RuleID: "STE-100-04", Param: "pattern", Value: `"("`}}}
	rs, err := ruleset.Load(o)
	if err != nil {
		t.Fatal(err)
	}
	rs.Rules["STE-100-04"] = func() ruleset.Rule { r := rs.Rules["STE-100-04"]; r.Check = "regex_ban"; return r }()
	if _, err := New(rs); err == nil || !strings.Contains(err.Error(), "STE-100-04: pattern") {
		t.Fatalf("bad pattern: err = %v", err)
	}
}

func TestCollect(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"a.md", "b.go", "c.txt", "sub/d.mdx", ".hidden/e.md", "node_modules/f.md", "testdata/g.md", "skip/h.md", "sub/index.md"} {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
	got, err := Collect([]string{".", ".hidden"}, []string{"skip", "index.md"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{".hidden/e.md", "a.md", "b.go", "sub/d.mdx"}; !slices.Equal(got, want) {
		t.Fatalf("Collect = %q, want %q", got, want)
	}
	if _, err := Collect([]string{"c.txt"}, nil); err == nil {
		t.Fatal("a named text file gives no error")
	}
	if _, err := Collect([]string{"."}, []string{"["}); err == nil {
		t.Fatal("a bad pattern gives no error")
	}
	if _, err := Collect([]string{"missing"}, nil); err == nil {
		t.Fatal("a missing path gives no error")
	}
}

func TestIdentifiersAreNotWords(t *testing.T) {
	e := engineFor(t, ruleset.Options{})
	got := check(t, e, "x.go", "package p\n\n// skipFenceOrComment and TestLexiconProjection and max_mapping2 stay.\n")
	if len(got) != 0 {
		t.Fatalf("findings = %q, want none for code names", got)
	}
	if got := check(t, e, "x.md", "The Processing step is slow.\n"); len(got) != 1 {
		t.Fatalf("a capital first letter is a word: %q", got)
	}
}

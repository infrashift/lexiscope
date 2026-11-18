package nasalint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// writeModule writes a module example.com/m with the given files into a
// temporary directory and returns the directory.
func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	files["go.mod"] = "module example.com/m\n\ngo 1.26\n"
	for name, src := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

var wantRE = regexp.MustCompile(`// want((?: R[0-9]+)+)`)

// wants returns "path:line: Rn" for each `// want Rn` marker in the files.
func wants(files map[string]string) []string {
	var out []string
	for name, src := range files {
		for i, line := range strings.Split(src, "\n") {
			m := wantRE.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			for _, rule := range strings.Fields(m[1]) {
				out = append(out, fmt.Sprintf("%s:%d: %s", name, i+1, rule))
			}
		}
	}
	slices.Sort(out)
	return out
}

func gots(res *Result) []string {
	var out []string
	for _, f := range res.Findings {
		if f.Rule != "R5" {
			out = append(out, fmt.Sprintf("%s:%d: %s", f.Path, f.Line, f.Rule))
		}
	}
	slices.Sort(out)
	return out
}

// check runs the linter on the files and compares the findings, R5 aside,
// with the want markers.
func check(t *testing.T, files map[string]string) *Result {
	t.Helper()
	want := wants(files)
	res, err := Run(writeModule(t, files))
	if err != nil {
		t.Fatal(err)
	}
	if got := gots(res); !slices.Equal(got, want) {
		t.Fatalf("findings:\n%s\nwant:\n%s\nall:\n%v", strings.Join(got, "\n"), strings.Join(want, "\n"), res.Findings)
	}
	return res
}

func TestR1Recursion(t *testing.T) {
	check(t, map[string]string{"p/r1.go": `package p

func fact(n int) int { // want R1
	if n == 0 {
		return 1
	}
	return n * fact(n-1)
}

func even(n int) bool { // want R1
	if n == 0 {
		return true
	}
	return odd(n - 1)
}

func odd(n int) bool { return n != 0 && even(n-1) }

type walker interface{ walk() }

type node struct{ next walker }

func (n node) walk() { n.next.walk() } // want R1

func gen[T any](x T, n int) { // want R1
	if n > 0 {
		gen[T](x, n-1)
	}
}

func jump() {
	goto end // want R1
end:
}

// Calls without a cycle, and calls into other packages.
func a() { b(); b() }
func b() { c() }
func c() { println("c") }
`, "p/r1b.go": `package p

// A cycle across two files of one package.
func ping(n int) { // want R1
	if n > 0 {
		pong(n - 1)
	}
}
`, "p/r1c.go": `package p

func pong(n int) { ping(n) }

func useGen() { gen(1, 2) }
`})
}

func TestR1GenericRecursion(t *testing.T) {
	check(t, map[string]string{"p/g.go": `package p

func gen[T any](x T, n int) { // want R1
	if n > 0 {
		gen[T](x, n-1)
	}
}
`})
}

func TestR2LoopsWithoutCondition(t *testing.T) {
	check(t, map[string]string{"p/r2.go": `package p

func loops(xs []int, n int) {
	for { // want R2
		break
	}
	for ;; { // want R2
		break
	}
	for i := 0; i < n; i++ {
	}
	for range n {
	}
	for _, x := range xs {
		println(x)
	}
	for n > 0 {
		n--
	}
}
`})
}

func TestR4FunctionLength(t *testing.T) {
	body := strings.Repeat("\tprintln()\n", 58)
	check(t, map[string]string{"p/r4.go": "package p\n\n" +
		"func sixty() {\n" + body + "}\n\n" + // 60 lines: allowed
		"func sixtyOne() { // want R4\n" + body + "\tprintln()\n}\n"})
}

func TestR7Discards(t *testing.T) {
	check(t, map[string]string{"p/r7.go": `package p

import "strconv"

type I interface{ M() }
type T struct{}

func (T) M() {}

func two() (int, error) { return 0, nil }

var _ I = T{}
var _ = two // want R7 R9

func discards(m map[string]int, x any) {
	_ = strconv.Itoa(1) // want R7
	_ = m // want R7
	n, _ := two() // want R7
	_, err := two() // want R7
	v, _ := m["k"]
	s, _ := x.(string)
	a, _ := strconv.Atoi("1") // want R7
	b, _ := 1, two // want R9
	println(n, err, v, s, a, b)
}
`})
}

func TestR9FunctionTypes(t *testing.T) {
	check(t, map[string]string{"p/r9.go": `package p

type F func() // want R9

type S struct {
	cb func() // want R9
}

type I interface {
	M(x int) int
}

func helper() {}

func (S) method() {}

func takes(f func()) {} // want R9

func values(s S) {
	f := helper // want R9
	g := s.method // want R9
	h := S.method // want R9
	k := func() {} // want R9
	helper()
	s.method()
	println(f, g, h, k)
}
`})
}

func TestAllowComments(t *testing.T) {
	check(t, map[string]string{"p/allow.go": `package p

func spin() {
	for { // nasa-lint: allow(R2) the test needs it.
		break
	}
	// nasa-lint: allow(R2) the comment on the line above counts too.
	for {
		break
	}
	for { // nasa-lint: allow(R7) a different rule does not count. // want R2
		break
	}
}
`})
}

func TestSkippedFilesAndDirectories(t *testing.T) {
	bad := "package p\n\nfunc f() {\n\tfor {\n\t}\n}\n"
	res := check(t, map[string]string{
		"p/x_test.go":            bad,
		"p/testdata/t.go":        bad,
		"p/.hidden/h.go":         bad,
		"p/_skip/s.go":           bad,
		"p/vendor/v.go":          bad,
		"docs/node_modules/n.go": bad,
		"nested/go.mod":          "module example.com/nested\n",
		"nested/n.go":            bad,
		"p/ok.go":                "package p\n",
	})
	if len(res.Funcs) != 0 {
		t.Fatalf("functions = %v, want none", res.Funcs)
	}
}

func TestR5Average(t *testing.T) {
	files := map[string]string{
		"internal/assert/assert.go": "package assert\n\nfunc True(c bool, m string) {\n\tif !c {\n\t\tpanic(m)\n\t}\n}\n",
		"p/p.go": `package p

import "example.com/m/internal/assert"

func three(x int) {
	assert.True(x > 0, "a")
	assert.True(x < 9, "b")
	func() { assert.True(x != 5, "c") }() // want R9
}

func none() {}
`,
	}
	res := check(t, files)
	if res.Assertions != 3 || len(res.Funcs) != 2 {
		t.Fatalf("assertions = %d in %d functions, want 3 in 2 (the assert package does not count)", res.Assertions, len(res.Funcs))
	}
	last := res.Findings[len(res.Findings)-1]
	want := Finding{Path: ".", Rule: "R5", Message: "3 assertions in 2 functions (average 1.50, min 2.0)"}
	if last != want {
		t.Fatalf("R5 finding = %+v, want %+v", last, want)
	}
	if got := last.String(); got != ".: R5: 3 assertions in 2 functions (average 1.50, min 2.0)" {
		t.Fatalf("String() = %q", got)
	}
}

func TestFindingFormat(t *testing.T) {
	f := Finding{Path: "a/b.go", Line: 7, Rule: "R2", Message: "loop without a condition `for {}`"}
	if got := f.String(); got != "a/b.go:7: R2: loop without a condition `for {}`" {
		t.Fatalf("String() = %q", got)
	}
}

func TestErrors(t *testing.T) {
	if _, err := Run(t.TempDir()); err == nil {
		t.Fatal("a directory without go.mod gives no error")
	}
	root := writeModule(t, map[string]string{"p/bad.go": "package p\n\nfunc {\n"})
	if _, err := Run(root); err == nil {
		t.Fatal("a syntax error gives no error")
	}
}

func TestGuessName(t *testing.T) {
	cases := map[string]string{
		"fmt":                               "fmt",
		"github.com/spf13/cobra":            "cobra",
		"github.com/owenrumney/go-sarif/v3": "sarif",
		"cuelang.org/go/cue/cuecontext":     "cuecontext",
		"example.com/some-lib":              "some_lib",
	}
	for in, want := range cases {
		if got := guessName(in); got != want {
			t.Errorf("guessName(%q) = %q, want %q", in, got, want)
		}
	}
}

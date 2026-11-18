package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func module(t *testing.T, src string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "m.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestExitStatus(t *testing.T) {
	cases := []struct {
		name string
		src  string
		args []string
		code int
		out  string
	}{
		{"clean", "package m\n", nil, 0, "nasa-lint: 0 error(s); 0 functions, 0 assertions (average 0.00)\n"},
		{"finding", "package m\n\nvar _ = f()\n\nfunc f() int { return 1 }\n", nil, 1,
			"m.go:3: R7: discarded value `var _ = ...`\n" +
				".: R5: 0 assertions in 1 functions (average 0.00, min 2.0)\n" +
				"nasa-lint: 2 error(s); 1 functions, 0 assertions (average 0.00)\n"},
		{"verbose", "package m\n\nfunc f() {}\n", []string{"--verbose"}, 1,
			"m.go:3: info: func f has 0 assertion(s)\n" +
				".: R5: 0 assertions in 1 functions (average 0.00, min 2.0)\n" +
				"nasa-lint: 1 error(s); 1 functions, 0 assertions (average 0.00)\n"},
		{"bad argument", "package m\n", []string{"--fix"}, 2, ""},
		{"syntax error", "package m\n\nfunc {\n", nil, 2, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Chdir(module(t, c.src))
			var stdout, stderr bytes.Buffer
			if code := run(c.args, &stdout, &stderr); code != c.code {
				t.Fatalf("code = %d, want %d; stderr: %s", code, c.code, stderr.String())
			}
			if got := stdout.String(); got != c.out {
				t.Fatalf("stdout:\n%s\nwant:\n%s", got, c.out)
			}
			if c.code == 2 && stderr.Len() == 0 {
				t.Fatal("a failure writes no message to stderr")
			}
		})
	}
}

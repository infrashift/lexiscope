package engine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/parser"
)

// MaxFiles bounds the files of one scan.
const MaxFiles = 1 << 16

// maxWalkSteps bounds the directory walk.
const maxWalkSteps = 1 << 20

// skippedDirs are the directory names that a walk does not enter. A walk
// also skips each hidden directory, unless the user names it as a path.
var skippedDirs = []string{"node_modules", "vendor", "testdata", "dist"}

// IsChecked reports whether lexis reads the file: Markdown, or a source file
// of a ztree grammar.
func IsChecked(name string) bool {
	assert.True(name != "", "the file has a name")
	lang, src := parser.LangOf(name)
	assert.True(src || lang == 0, "an unknown file has no grammar")
	return IsMarkdown(name) || src
}

// Collect returns the files to check under the paths, sorted and without
// duplicates. A named file must be a checked file. An exclude pattern
// matches a slash path, each parent directory of it, or its base name, with
// the syntax of path.Match.
func Collect(paths, excludes []string) ([]string, error) {
	assert.True(len(paths) > 0, "a scan has a path")
	assert.True(len(excludes) <= 1<<12, "the exclude list is bounded")
	for _, p := range excludes {
		if err := checkPattern(p); err != nil {
			return nil, fmt.Errorf("--exclude %q: %w", p, err)
		}
	}
	var files []string
	stack := slices.Clone(paths)
	for range maxWalkSteps {
		if len(stack) == 0 {
			slices.Sort(files)
			return slices.Compact(files), nil
		}
		p := filepath.Clean(stack[len(stack)-1])
		stack = stack[:len(stack)-1]
		named := slices.Contains(paths, p) || slices.Contains(paths, p+string(filepath.Separator))
		next, file, err := visit(p, named, excludes)
		if err != nil {
			return nil, err
		}
		stack = append(stack, next...)
		if file {
			files = append(files, p)
		}
		if len(files) > MaxFiles {
			return nil, fmt.Errorf("more than %d files", MaxFiles)
		}
	}
	return nil, fmt.Errorf("the walk exceeded %d steps", maxWalkSteps)
}

// visit returns the entries of a directory to walk, or reports whether p
// is a file to check.
func visit(p string, named bool, excludes []string) ([]string, bool, error) {
	assert.True(p != "", "the path is set")
	assert.True(len(excludes) <= 1<<12, "the exclude list is bounded")
	if excluded(p, excludes) {
		return nil, false, nil
	}
	info, err := os.Stat(p)
	if err != nil {
		return nil, false, err
	}
	if !info.IsDir() {
		if named && !IsChecked(p) {
			return nil, false, fmt.Errorf("%s: lexis reads Markdown files and .go, .zig, .py, .c, .h, .rs and C++ files", p)
		}
		return nil, IsChecked(p), nil
	}
	base := filepath.Base(p)
	if !named && (slices.Contains(skippedDirs, base) || (strings.HasPrefix(base, ".") && base != ".")) {
		return nil, false, nil
	}
	entries, err := os.ReadDir(p)
	if err != nil {
		return nil, false, err
	}
	next := make([]string, 0, len(entries))
	for _, e := range entries {
		next = append(next, filepath.Join(p, e.Name()))
	}
	return next, false, nil
}

// excluded reports whether a pattern matches p, a parent directory of p,
// or the base name of p.
func excluded(p string, excludes []string) bool {
	assert.True(p != "", "the path is set")
	assert.True(len(excludes) <= 1<<12, "the exclude list is bounded")
	slash := filepath.ToSlash(p)
	for _, pat := range excludes {
		if matchPattern(pat, path.Base(slash)) {
			return true
		}
		for dir := slash; dir != "." && dir != "/" && dir != ""; dir = path.Dir(dir) {
			if matchPattern(pat, dir) {
				return true
			}
		}
	}
	return false
}

// checkPattern reports a syntax error of a path.Match pattern.
func checkPattern(p string) error {
	assert.True(len(p) < 1<<12, "a pattern is bounded")
	matched, err := path.Match(p, "")
	assert.True(err == nil || !matched, "a bad pattern matches nothing")
	return err
}

// matchPattern is path.Match after Collect checked the pattern syntax.
func matchPattern(pat, name string) bool {
	assert.True(pat != "", "the pattern is set")
	ok, err := path.Match(pat, name)
	assert.True(err == nil, "Collect checks each pattern first")
	return ok
}

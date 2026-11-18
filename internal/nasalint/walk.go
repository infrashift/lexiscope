package nasalint

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/infrashift/lexis/internal/assert"
)

// modulePath reads the module path from root/go.mod.
func modulePath(root string) (string, error) {
	assert.True(root != "", "the root is set")
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	assert.True(len(lines) > 0, "Split gives one line or more")
	for _, line := range lines {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.Trim(strings.TrimSpace(rest), `"`), nil
		}
	}
	return "", errors.New("go.mod has no module line")
}

// packageDirs returns the directories under root that hold Go source, as
// slash-separated paths relative to root, in sorted order. It walks with an
// explicit stack. It skips hidden directories, testdata, vendor,
// node_modules and each directory with its own go.mod.
func packageDirs(root string) ([]string, error) {
	assert.True(root != "", "the root is set")
	stack := []string{"."}
	var dirs []string
	for range MaxDirs {
		if len(stack) == 0 {
			slices.Sort(dirs)
			return dirs, nil
		}
		assert.True(len(dirs) < MaxDirs, "each step adds one directory or none")
		rel := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		sub, hasGo := splitEntries(root, rel, entries)
		stack = append(stack, sub...)
		if hasGo {
			dirs = append(dirs, rel)
		}
	}
	return nil, fmt.Errorf("more than %d directories", MaxDirs)
}

// splitEntries returns the subdirectories to walk, and whether the
// directory holds a Go source file that is not a test.
func splitEntries(root, rel string, entries []fs.DirEntry) ([]string, bool) {
	assert.True(root != "" && rel != "", "the paths are set")
	assert.True(len(entries) <= MaxFiles*MaxDirs, "the entry count is bounded")
	var sub []string
	hasGo := false
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() {
			hasGo = hasGo || isSource(name)
			continue
		}
		if skipDir(name) || isModule(filepath.Join(root, filepath.FromSlash(rel), name)) {
			continue
		}
		sub = append(sub, path.Join(rel, name))
	}
	return sub, hasGo
}

func skipDir(name string) bool {
	assert.True(name != "", "the name is set")
	assert.True(!strings.Contains(name, "/"), "the name has one element")
	switch name {
	case "testdata", "vendor", "node_modules":
		return true
	}
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func isModule(dir string) bool {
	assert.True(dir != "", "the directory is set")
	info, err := os.Stat(filepath.Join(dir, "go.mod"))
	assert.True(err == nil || info == nil, "Stat gives info or an error")
	return err == nil && !info.IsDir()
}

func isSource(name string) bool {
	assert.True(name != "", "the name is set")
	assert.True(!strings.Contains(name, "/"), "the name has one element")
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") &&
		!strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

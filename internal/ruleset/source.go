package ruleset

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"cuelang.org/go/cue"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/rules/ste100"
)

// source is one CUE file of a load.
type source struct {
	name string // the built-in name or the file path
	data []byte
}

// Builtins returns the names of the built-in rulesets.
func Builtins() []string {
	names := []string{ste100.Name}
	assert.True(len(names) == 1, "lexis has one built-in ruleset")
	assert.True(names[0] != "", "a built-in ruleset has a name")
	return names
}

// resolve reads the sources that the names select: a built-in ruleset, a
// CUE file, or each CUE file of a directory (not its subdirectories).
func resolve(names []string) ([]source, error) {
	assert.True(len(names) > 0, "a load has a source")
	assert.True(len(names) <= MaxSources, "the source count is bounded")
	var out []source
	for _, name := range names {
		if name == ste100.Name {
			out = append(out, source{name: name, data: ste100.Source()})
			continue
		}
		files, err := cueFiles(name)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			src, err := readSource(f)
			if err != nil {
				return nil, err
			}
			out = append(out, src)
		}
	}
	return out, nil
}

// cueFiles returns name when it is a CUE file, or the sorted CUE files of
// the directory name.
func cueFiles(name string) ([]string, error) {
	assert.True(len(Builtins()) > 0, "the built-in names exist")
	info, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("ruleset %q: not a built-in ruleset (%s) and not a CUE file or directory", name, strings.Join(Builtins(), ", "))
	}
	if !info.IsDir() {
		if filepath.Ext(name) != ".cue" {
			return nil, fmt.Errorf("ruleset %q: the file name does not end in .cue", name)
		}
		return []string{name}, nil
	}
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".cue" {
			files = append(files, filepath.Join(name, e.Name()))
		}
	}
	if len(files) == 0 || len(files) > MaxFiles {
		return nil, fmt.Errorf("ruleset %q: the directory has %d CUE files; it needs 1 to %d", name, len(files), MaxFiles)
	}
	slices.Sort(files)
	return files, nil
}

// readSource reads one CUE file of at most MaxFileBytes.
func readSource(path string) (source, error) {
	assert.True(path != "", "the path is set")
	info, err := os.Stat(path)
	if err != nil {
		return source{}, err
	}
	if info.Size() > MaxFileBytes {
		return source{}, fmt.Errorf("ruleset %q: %d bytes, more than %d", path, info.Size(), MaxFileBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return source{}, err
	}
	assert.True(len(data) <= MaxFileBytes || int64(len(data)) > info.Size(), "the size check holds unless the file grew")
	return source{name: path, data: data}, nil
}

// sourceNames joins the names of the sources for an error message.
func sourceNames(srcs []source) string {
	assert.True(len(srcs) > 0, "a load has a source")
	names := make([]string, len(srcs))
	for i, s := range srcs {
		names[i] = s.name
	}
	assert.True(len(names) == len(srcs), "each source has a name")
	return strings.Join(names, " + ")
}

// rulePath returns the path rules."<id>".<elem>...
func rulePath(id string, elems ...string) cue.Path {
	assert.True(id != "", "the rule ID is set")
	assert.True(len(elems) <= 2, "a rule path has two elements or fewer")
	sels := []cue.Selector{cue.Str("rules"), cue.Str(id)}
	for _, e := range elems {
		sels = append(sels, cue.Str(e))
	}
	return cue.MakePath(sels...)
}

package nasalint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/infrashift/lexiscope/internal/assert"
)

// pkgCheck holds the state of the check of one package.
type pkgCheck struct {
	fset   *token.FileSet
	pkg    *types.Package
	info   *types.Info
	files  []*ast.File
	paths  []string // slash-separated path of each file, relative to the root
	allows []map[int][]string
	res    *Result
	// typeErrors counts the errors of go/types. They are normal: the
	// importer gives empty packages.
	typeErrors int
}

// checkPackage parses and checks the package in root/rel. It returns the
// number of files it read.
func checkPackage(root, module, rel string, res *Result) (int, error) {
	assert.True(root != "" && module != "" && rel != "", "the paths are set")
	assert.True(res != nil, "the result exists")
	pc := &pkgCheck{fset: token.NewFileSet(), res: res}
	if err := pc.parseDir(root, rel); err != nil {
		return 0, err
	}
	if len(pc.files) == 0 {
		return 0, nil
	}
	if err := pc.typeCheck(path.Join(module, rel)); err != nil {
		return 0, err
	}
	for i, f := range pc.files {
		if err := pc.checkFile(i, f); err != nil {
			return 0, err
		}
	}
	pc.checkGraph()
	return len(pc.files), nil
}

// parseDir parses each Go file of the directory that is not a test.
func (pc *pkgCheck) parseDir(root, rel string) error {
	assert.True(pc.fset != nil, "the file set exists")
	assert.True(len(pc.files) == 0, "the directory is parsed once")
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	if len(entries) > MaxFiles {
		return fmt.Errorf("%s: more than %d entries", rel, MaxFiles)
	}
	for _, e := range entries {
		if e.IsDir() || !isSource(e.Name()) {
			continue
		}
		p := path.Join(rel, e.Name())
		f, err := parser.ParseFile(pc.fset, filepath.Join(root, filepath.FromSlash(p)), nil,
			parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		pc.files = append(pc.files, f)
		pc.paths = append(pc.paths, p)
		pc.allows = append(pc.allows, allowComments(pc.fset, f))
	}
	return nil
}

// typeCheck resolves the names of the package. Imports resolve to empty
// packages, so a call into another package has no target. R1 needs the
// targets in this package only.
func (pc *pkgCheck) typeCheck(pkgPath string) error {
	assert.True(pkgPath != "", "the package path is set")
	assert.True(len(pc.files) > 0, "the package has files")
	pc.info = &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	conf := types.Config{
		Importer:    emptyImporter{},
		FakeImportC: true,
		Error:       pc.countTypeError, // nasa-lint: allow(R9) go/types reports each error through this callback.
	}
	pkg, err := conf.Check(pkgPath, pc.fset, pc.files, pc.info)
	if err != nil && pc.typeErrors == 0 {
		return err
	}
	pc.pkg = pkg
	return nil
}

func (pc *pkgCheck) countTypeError(err error) {
	assert.True(err != nil, "go/types reports an error")
	assert.True(pc.typeErrors >= 0, "the count is not negative")
	pc.typeErrors++
}

// emptyImporter gives an empty package for each import path.
type emptyImporter struct{}

// Import returns types.Unsafe for "unsafe" and an empty package otherwise.
func (emptyImporter) Import(p string) (*types.Package, error) {
	assert.True(p != "", "the import path is set")
	if p == "unsafe" {
		return types.Unsafe, nil
	}
	pkg := types.NewPackage(p, guessName(p))
	pkg.MarkComplete()
	assert.True(pkg.Path() == p, "the package keeps its path")
	return pkg, nil
}

var versionElem = regexp.MustCompile(`^v[0-9]+$`)

// guessName returns the probable package name of an import path: the last
// element, or the element before a major version suffix such as /v3.
func guessName(p string) string {
	assert.True(p != "", "the import path is set")
	elems := strings.Split(p, "/")
	name := elems[len(elems)-1]
	if versionElem.MatchString(name) && len(elems) > 1 {
		name = elems[len(elems)-2]
	}
	name = strings.TrimPrefix(name, "go-")
	name = strings.NewReplacer("-", "_", ".", "_").Replace(name)
	assert.True(name != "", "the name is not empty")
	return name
}

var allowRE = regexp.MustCompile(`nasa-lint: allow\((R[0-9]+)\)`)

// allowComments maps each line with an allow comment to its rules.
func allowComments(fset *token.FileSet, f *ast.File) map[int][]string {
	assert.True(fset != nil && f != nil, "the file exists")
	assert.True(len(f.Comments) <= MaxNodesPerFile, "the comment count is bounded")
	allows := map[int][]string{}
	for _, group := range f.Comments {
		for _, c := range group.List {
			for _, m := range allowRE.FindAllStringSubmatch(c.Text, -1) {
				line := fset.Position(c.Pos()).Line
				allows[line] = append(allows[line], m[1])
			}
		}
	}
	return allows
}

// report adds a finding unless an allow comment on its line or the line
// above names its rule.
func (pc *pkgCheck) report(file int, pos token.Pos, rule, msg string) {
	assert.True(file >= 0 && file < len(pc.files), "the file index is valid")
	assert.True(pos.IsValid(), "the position is valid")
	line := pc.fset.Position(pos).Line
	if slices.Contains(pc.allows[file][line], rule) || slices.Contains(pc.allows[file][line-1], rule) {
		return
	}
	pc.res.Findings = append(pc.res.Findings, Finding{Path: pc.paths[file], Line: line, Rule: rule, Message: msg})
}

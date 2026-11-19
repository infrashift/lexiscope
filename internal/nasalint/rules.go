package nasalint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/infrashift/lexiscope/internal/assert"
)

// nodeCtx carries facts from a parent node to its children. ast.Preorder
// visits a parent before its children, so a parent records the facts first.
type nodeCtx struct {
	// sigs holds the function types of declarations, literals and interface
	// methods. Each other function type breaks R9.
	sigs map[*ast.FuncType]bool
	// called holds the expressions in call position. Each other use of a
	// function of this package is a function value.
	called map[ast.Expr]bool
}

// checkFile runs the rules on one file.
func (pc *pkgCheck) checkFile(file int, f *ast.File) error {
	assert.True(pc.info != nil, "the package has type information")
	assert.True(file >= 0 && file < len(pc.files), "the file index is valid")
	nc := nodeCtx{sigs: map[*ast.FuncType]bool{}, called: map[ast.Expr]bool{}}
	count := 0
	for n := range ast.Preorder(f) {
		count++
		if count > MaxNodesPerFile {
			return fmt.Errorf("%s: more than %d syntax nodes", pc.paths[file], MaxNodesPerFile)
		}
		pc.recordContext(n, &nc)
		pc.checkNode(file, n, &nc)
	}
	pc.checkFuncs(file, f)
	return nil
}

// recordContext records the facts of n for its children.
func (pc *pkgCheck) recordContext(n ast.Node, nc *nodeCtx) {
	assert.True(n != nil, "the node exists")
	assert.True(nc.sigs != nil && nc.called != nil, "the context exists")
	switch n := n.(type) {
	case *ast.FuncDecl:
		nc.sigs[n.Type] = true
	case *ast.FuncLit:
		nc.sigs[n.Type] = true
	case *ast.InterfaceType:
		for _, m := range n.Methods.List {
			if ft, ok := m.Type.(*ast.FuncType); ok {
				nc.sigs[ft] = true
			}
		}
	case *ast.CallExpr:
		base := callBase(n.Fun)
		nc.called[base] = true
		if sel, ok := base.(*ast.SelectorExpr); ok {
			nc.called[sel.Sel] = true
		}
	}
}

// checkNode reports the findings of one node.
func (pc *pkgCheck) checkNode(file int, n ast.Node, nc *nodeCtx) {
	assert.True(n != nil, "the node exists")
	assert.True(nc != nil, "the context exists")
	switch n := n.(type) {
	case *ast.BranchStmt:
		if n.Tok == token.GOTO {
			pc.report(file, n.Pos(), "R1", "goto statement")
		}
	case *ast.ForStmt:
		if n.Cond == nil {
			pc.report(file, n.Pos(), "R2", "loop without a condition `for {}`")
		}
	case *ast.AssignStmt:
		pc.checkAssign(file, n)
	case *ast.ValueSpec:
		if n.Type == nil && len(n.Values) > 0 && allBlank(identExprs(n.Names)) {
			pc.report(file, n.Pos(), "R7", "discarded value `var _ = ...`")
		}
	case *ast.FuncLit:
		pc.report(file, n.Pos(), "R9", "function literal")
	case *ast.FuncType:
		if !nc.sigs[n] {
			pc.report(file, n.Pos(), "R9", "function type")
		}
	case *ast.Ident:
		pc.checkFuncValue(file, n, nc)
	}
}

// checkAssign reports a discarded value or a discarded call result.
func (pc *pkgCheck) checkAssign(file int, s *ast.AssignStmt) {
	assert.True(s != nil, "the statement exists")
	assert.True(len(s.Lhs) > 0 && len(s.Rhs) > 0, "an assignment has two sides")
	if allBlank(s.Lhs) {
		pc.report(file, s.Pos(), "R7", "discarded value `_ = ...`")
		return
	}
	for k, l := range s.Lhs {
		if !isBlank(l) {
			continue
		}
		rhs := s.Rhs[0]
		if len(s.Rhs) == len(s.Lhs) {
			rhs = s.Rhs[k]
		}
		if pc.isCall(rhs) {
			pc.report(file, s.Pos(), "R7", "discarded call result `x, _ := f()`")
			return
		}
	}
}

// checkFuncValue reports a function of this package outside call position.
func (pc *pkgCheck) checkFuncValue(file int, id *ast.Ident, nc *nodeCtx) {
	assert.True(id != nil, "the identifier exists")
	assert.True(pc.pkg != nil, "the package exists")
	fn, ok := pc.info.Uses[id].(*types.Func)
	if !ok || fn.Pkg() != pc.pkg || nc.called[id] {
		return
	}
	pc.report(file, id.Pos(), "R9", fmt.Sprintf("function value `%s`", id.Name))
}

// isCall reports whether e is a call and not a type conversion. A call into
// another package counts as a call, because its target is unknown.
func (pc *pkgCheck) isCall(e ast.Expr) bool {
	assert.True(e != nil, "the expression exists")
	assert.True(pc.info != nil, "the package has type information")
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	tv, known := pc.info.Types[call.Fun]
	return !known || !tv.IsType()
}

// checkFuncs checks the length (R4) and counts the assertions (R5) of each
// function declaration with a body.
func (pc *pkgCheck) checkFuncs(file int, f *ast.File) {
	assert.True(f != nil, "the file exists")
	assert.True(pc.pkg != nil, "the package exists")
	countR5 := !strings.HasSuffix(pc.pkg.Path(), AssertPathSuffix)
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.FuncDecl)
		if !ok || d.Body == nil {
			continue
		}
		start := pc.fset.Position(d.Pos()).Line
		length := pc.fset.Position(d.Body.Rbrace).Line - start + 1
		name := declName(d)
		if length > MaxFnLines {
			pc.report(file, d.Pos(), "R4", fmt.Sprintf("func %s has %d lines (max %d)", name, length, MaxFnLines))
		}
		if countR5 {
			n := pc.countAssertions(d.Body)
			pc.res.Assertions += n
			pc.res.Funcs = append(pc.res.Funcs, FuncStat{Path: pc.paths[file], Line: start, Name: name, Assertions: n})
		}
	}
}

// countAssertions counts the calls into the assert package in body,
// function literals included.
func (pc *pkgCheck) countAssertions(body *ast.BlockStmt) int {
	assert.True(body != nil, "the body exists")
	assert.True(pc.info != nil, "the package has type information")
	n := 0
	for node := range ast.Preorder(body) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			continue
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		x, ok := sel.X.(*ast.Ident)
		if !ok {
			continue
		}
		if pn, ok := pc.info.Uses[x].(*types.PkgName); ok && strings.HasSuffix(pn.Imported().Path(), AssertPathSuffix) {
			n++
		}
	}
	return n
}

// callBase removes parentheses and type arguments from the function of a
// call: (f)(x) and f[int](x) call f.
func callBase(e ast.Expr) ast.Expr {
	assert.True(e != nil, "the expression exists")
	e = ast.Unparen(e)
	switch x := e.(type) {
	case *ast.IndexExpr:
		e = ast.Unparen(x.X)
	case *ast.IndexListExpr:
		e = ast.Unparen(x.X)
	}
	assert.True(e != nil, "the base exists")
	return e
}

func isBlank(e ast.Expr) bool {
	assert.True(e != nil, "the expression exists")
	id, ok := e.(*ast.Ident)
	assert.True(!ok || id.Name != "", "an identifier has a name")
	return ok && id.Name == "_"
}

func allBlank(es []ast.Expr) bool {
	assert.True(len(es) > 0, "the list is not empty")
	assert.True(len(es) <= MaxNodesPerFile, "the list is bounded")
	for _, e := range es {
		if !isBlank(e) {
			return false
		}
	}
	return true
}

func identExprs(ids []*ast.Ident) []ast.Expr {
	assert.True(len(ids) > 0, "the list is not empty")
	out := make([]ast.Expr, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	assert.True(len(out) == len(ids), "each identifier is copied")
	return out
}

// declName returns Name, or Recv.Name for a method.
func declName(d *ast.FuncDecl) string {
	assert.True(d != nil && d.Name != nil, "the declaration has a name")
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return d.Name.Name
	}
	t := d.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	t = callBase(t) // A generic receiver T[K] names T.
	recv, ok := t.(*ast.Ident)
	assert.True(ok, "a receiver type is a name")
	return recv.Name + "." + d.Name.Name
}

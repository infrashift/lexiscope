package nasalint

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/infrashift/lexis/internal/assert"
)

// callGraph is the call graph of one package. Node k is the function
// declaration decls[k] in file files[k].
type callGraph struct {
	decls  []*ast.FuncDecl
	files  []int
	funcs  []*types.Func
	index  map[*types.Func]int
	byName map[string][]int // methods by name, for calls through an interface
	edges  [][]int
}

// checkGraph reports each recursion in the package (R1).
func (pc *pkgCheck) checkGraph() {
	assert.True(pc.pkg != nil, "the package exists")
	assert.True(len(pc.files) == len(pc.paths), "each file has a path")
	g := pc.buildGraph()
	for _, cyc := range findCycles(g.edges) {
		names := make([]string, len(cyc))
		for i, k := range cyc {
			names[i] = declName(g.decls[k])
		}
		first := cyc[0]
		pc.report(g.files[first], g.decls[first].Pos(), "R1", "recursion: "+strings.Join(names, " -> "))
	}
}

// buildGraph collects the function declarations of the package and the
// calls between them.
func (pc *pkgCheck) buildGraph() *callGraph {
	assert.True(pc.info != nil, "the package has type information")
	assert.True(len(pc.files) > 0, "the package has files")
	g := &callGraph{index: map[*types.Func]int{}, byName: map[string][]int{}}
	for file, f := range pc.files {
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.FuncDecl)
			if !ok || d.Body == nil {
				continue
			}
			fn, ok := pc.info.Defs[d.Name].(*types.Func)
			if !ok {
				continue // An init or blank function has no object.
			}
			g.index[fn] = len(g.decls)
			if d.Recv != nil {
				g.byName[fn.Name()] = append(g.byName[fn.Name()], len(g.decls))
			}
			g.decls = append(g.decls, d)
			g.files = append(g.files, file)
			g.funcs = append(g.funcs, fn)
		}
	}
	assert.True(len(g.decls) <= MaxFuncs, "the package keeps the function limit")
	g.edges = make([][]int, len(g.decls))
	for k, d := range g.decls {
		g.edges[k] = pc.callees(g, d.Body)
	}
	return g
}

// callees returns the sorted nodes that the calls in body reach. A call
// through an interface of this package reaches each method of that name.
func (pc *pkgCheck) callees(g *callGraph, body *ast.BlockStmt) []int {
	assert.True(g != nil && body != nil, "the graph and the body exist")
	assert.True(pc.pkg != nil, "the package exists")
	var out []int
	for n := range ast.Preorder(body) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			continue
		}
		fn := pc.calledFunc(call)
		if fn == nil {
			continue
		}
		if k, ok := g.index[fn]; ok {
			out = append(out, k)
		} else if fn.Pkg() == pc.pkg && isInterfaceMethod(fn) {
			out = append(out, g.byName[fn.Name()]...)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// calledFunc returns the generic origin of the called function, or nil for
// a call of a function value, a conversion or a builtin.
func (pc *pkgCheck) calledFunc(call *ast.CallExpr) *types.Func {
	assert.True(call != nil, "the call exists")
	assert.True(pc.info != nil, "the package has type information")
	var id *ast.Ident
	switch base := callBase(call.Fun).(type) {
	case *ast.Ident:
		id = base
	case *ast.SelectorExpr:
		id = base.Sel
	default:
		return nil
	}
	fn, ok := pc.info.Uses[id].(*types.Func)
	if !ok {
		return nil
	}
	return fn.Origin()
}

func isInterfaceMethod(fn *types.Func) bool {
	assert.True(fn != nil, "the function exists")
	sig, ok := fn.Type().(*types.Signature)
	assert.True(ok, "a function has a signature")
	return sig.Recv() != nil && types.IsInterface(sig.Recv().Type())
}

// frame is one entry of the explicit stack of findCycles.
type frame struct {
	node int
	next int // index of the next edge to follow
}

// findCycles runs an iterative three-color depth-first search over the
// nodes in order. It returns the path of each back edge, from the target of
// the back edge to its source and back to the target.
func findCycles(edges [][]int) [][]int {
	assert.True(len(edges) <= MaxFuncs, "the graph keeps the function limit")
	color := make([]uint8, len(edges)) // 0 new, 1 on the path, 2 done
	budget := len(edges)
	for _, e := range edges {
		budget += len(e)
	}
	var cycles [][]int
	for root := range edges {
		if color[root] != 0 {
			continue
		}
		var err error
		cycles, budget, err = walkFrom(root, edges, color, cycles, budget)
		assert.True(err == nil, fmt.Sprint(err))
	}
	return cycles
}

// walkFrom runs the search from one root. Each step follows one edge or
// leaves one node, so budget, the sum of the edges and the nodes, bounds it.
func walkFrom(root int, edges [][]int, color []uint8, cycles [][]int, budget int) ([][]int, int, error) {
	assert.True(root >= 0 && root < len(edges), "the root is a node")
	assert.True(color[root] == 0, "the root is new")
	stack := []frame{{node: root}}
	path := []int{root}
	color[root] = 1
	for range budget + 1 {
		if len(stack) == 0 {
			return cycles, budget, nil
		}
		budget--
		top := &stack[len(stack)-1]
		if top.next == len(edges[top.node]) {
			color[top.node] = 2
			stack = stack[:len(stack)-1]
			path = path[:len(path)-1]
			continue
		}
		next := edges[top.node][top.next]
		top.next++
		switch color[next] {
		case 1:
			start := slices.Index(path, next)
			cycles = append(cycles, append(slices.Clone(path[start:]), next))
		case 0:
			color[next] = 1
			stack = append(stack, frame{node: next})
			path = append(path, next)
		}
	}
	return cycles, budget, fmt.Errorf("the search from node %d exceeded its budget", root)
}

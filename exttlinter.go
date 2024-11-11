package exttlinter

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "exttlinter is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "exttlinter",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		buildssa.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	var testingPkg *types.Package
	for _, p := range pass.Pkg.Imports() {
		if p.Path() == "testing" {
			testingPkg = p
			break
		}
	}

	// skip if testing package is not imported because it is not a test file
	if testingPkg == nil {
		return nil, nil
	}
	tbObj := testingPkg.Scope().Lookup("TB")
	if tbObj == nil {
		return nil, fmt.Errorf("testing package not TB interface")
	}
	tbIface, ok := tbObj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("testing.TB is not an interface")
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncLit)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncLit:
			body := n.Body.List
			for _, stmt := range body {
				exprStmt, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}

				callExpr, ok := exprStmt.X.(*ast.CallExpr)
				if !ok {
					continue
				}

				selectExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				ident, ok := selectExpr.X.(*ast.Ident)
				if !ok {
					continue
				}

				obj := pass.TypesInfo.ObjectOf(ident)
				if obj == nil {
					pass.Reportf(ident.Pos(), "object is nil")
				}

				if types.Satisfies(obj.Type(), tbIface) {
					// TODO: check if the object is external scope testing object
					pass.Reportf(ident.Pos(), fmt.Sprintf("should not use external testing object."))
				}
			}
		}
	})

	return nil, nil
}

package exttlinter

import (
	"fmt"
	"go/ast"
	"go/types"

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

func lookupTestingObject(pass *analysis.Pass) (*types.Interface, error) {
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

	return tbIface, nil
}

func run(pass *analysis.Pass) (any, error) {
	tbIface, err := lookupTestingObject(pass)
	if err != nil {
		return nil, err
	}

	// no test file
	if tbIface == nil {
		return nil, nil
	}

	isTestingObject := func(pass *analysis.Pass, ident *ast.Ident) (bool, error) {
		obj := pass.TypesInfo.ObjectOf(ident)
		if obj == nil {
			return false, fmt.Errorf("object not found")
		}

		if types.Implements(obj.Type(), tbIface) {
			return true, nil
		}

		return false, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncLit)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncLit:
			// memorize testing objects
			tObjects := make(map[types.Object]struct{}, 0)
			params := n.Type.Params.List
			for _, p := range params {
				for _, name := range p.Names {
					isTesting, err := isTestingObject(pass, name)
					if err != nil {
						continue
					}
					obj := pass.TypesInfo.ObjectOf(name)
					if isTesting {
						tObjects[obj] = struct{}{}
					}
				}

			}

			// check if the testing object method called
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

				isTesting, err := isTestingObject(pass, ident)
				if err != nil {
					continue
				}
				if isTesting {
					if _, ok := tObjects[pass.TypesInfo.ObjectOf(ident)]; !ok {
						pass.Reportf(ident.Pos(), "should not use external testing object.")
					}
				}
			}
		}
	})

	return nil, nil
}

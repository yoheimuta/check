package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"os"

	"github.com/opennota/check"
)

type visitor struct {
	fset     *token.FileSet
	m        map[*ast.Object][]string
	funcName string
}

var exitStatus int

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.FuncDecl:
		v.funcName = node.Name.Name
		v.m = make(map[*ast.Object][]string)

	case *ast.DeferStmt:
		if sel, ok := node.Call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				if selectors, ok := v.m[ident.Obj]; !ok {
					v.m[ident.Obj] = []string{sel.Sel.Name}
				} else {
					found := false
					for _, selname := range selectors {
						if selname == sel.Sel.Name {
							pos := v.fset.Position(node.Pos())
							fmt.Printf("%s:%d: Repeating defer %s.%s() inside function %s\n",
								pos.Filename, pos.Line,
								ident.Name, selname, v.funcName)
							found = true
							exitStatus = 1
							break
						}
					}
					if !found {
						v.m[ident.Obj] = append(selectors, sel.Sel.Name)
					}
				}
			}
		}
	}
	return v
}

func main() {
	flag.Parse()
	pkgPath := "."
	if len(flag.Args()) > 0 {
		pkgPath = flag.Arg(0)
	}
	visitor := &visitor{}
	fset, astFiles := check.ASTFilesForPackage(pkgPath)
	visitor.fset = fset
	for _, f := range astFiles {
		ast.Walk(visitor, f)
	}
	os.Exit(exitStatus)
}

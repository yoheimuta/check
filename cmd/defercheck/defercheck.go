package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"

	"github.com/opennota/check"
)

type visitor struct {
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
							fmt.Printf("Repeating defer %s.%s() inside function %s\n",
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
	_, astFiles := check.ASTFilesForPackage(pkgPath)
	for _, f := range astFiles {
		ast.Walk(visitor, f)
	}
	os.Exit(exitStatus)
}

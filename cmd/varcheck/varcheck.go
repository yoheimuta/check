package main

import (
	"flag"
	"fmt"
	"go/ast"
	"os"

	"github.com/opennota/check"
)

var (
	reportExported = flag.Bool("e", false, "Report exported variables and constants")
)

type visitor struct {
	m          map[*ast.Object]int
	insideFunc bool
}

func (v *visitor) decl(obj *ast.Object) {
	if _, ok := v.m[obj]; !ok {
		v.m[obj] = 0
	}
}

func (v *visitor) use(obj *ast.Object) {
	if _, ok := v.m[obj]; ok {
		v.m[obj]++
	} else {
		v.m[obj] = 1
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Ident:
		v.use(node.Obj)

	case *ast.ValueSpec:
		if !v.insideFunc {
			for _, ident := range node.Names {
				if ident.Name != "_" {
					v.decl(ident.Obj)
				}
			}
		}
		for _, val := range node.Values {
			ast.Walk(v, val)
		}
		return nil

	case *ast.FuncDecl:
		if node.Body != nil {
			v.insideFunc = true
			ast.Walk(v, node.Body)
			v.insideFunc = false
		}

		return nil
	}

	return v
}

func main() {
	flag.Parse()
	pkgPath := "."
	if len(flag.Args()) > 0 {
		pkgPath = flag.Arg(0)
	}
	visitor := &visitor{m: make(map[*ast.Object]int)}
	_, astFiles := check.ASTFilesForPackage(pkgPath)
	for _, f := range astFiles {
		ast.Walk(visitor, f)
	}
	exitStatus := 0
	for obj, useCount := range visitor.m {
		if useCount == 0 && (*reportExported || !check.IsExported(obj.Name)) {
			fmt.Println(obj.Name)
			exitStatus = 1
		}
	}
	os.Exit(exitStatus)
}

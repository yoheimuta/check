package main

import (
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"

	"code.google.com/p/go.tools/go/types"
	"honnef.co/go/importer"

	"github.com/opennota/check"
)

var (
	reportExported = flag.Bool("e", false, "Report exported variables and constants")
)

type visitor struct {
	pkg        *types.Package
	info       types.Info
	m          map[types.Object]int
	insideFunc bool
}

func (v *visitor) decl(obj types.Object) {
	if _, ok := v.m[obj]; !ok {
		v.m[obj] = 0
	}
}

func (v *visitor) use(obj types.Object) {
	if _, ok := v.m[obj]; ok {
		v.m[obj]++
	} else {
		v.m[obj] = 1
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Ident:
		v.use(v.info.Uses[node])

	case *ast.ValueSpec:
		if !v.insideFunc {
			for _, ident := range node.Names {
				if ident.Name != "_" {
					v.decl(v.info.Defs[ident])
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
	visitor := &visitor{
		info: types.Info{
			Defs: make(map[*ast.Ident]types.Object),
			Uses: make(map[*ast.Ident]types.Object),
		},

		m: make(map[types.Object]int),
	}
	fset, astFiles := check.ASTFilesForPackage(pkgPath)
	imp := importer.New()
	config := types.Config{Import: imp.Import}
	var err error
	visitor.pkg, err = config.Check(pkgPath, fset, astFiles, &visitor.info)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range astFiles {
		ast.Walk(visitor, f)
	}
	exitStatus := 0
	for obj, useCount := range visitor.m {
		if useCount == 0 && (*reportExported || !check.IsExported(obj.Name())) {
			fmt.Println(obj.Name())
			exitStatus = 1
		}
	}
	os.Exit(exitStatus)
}

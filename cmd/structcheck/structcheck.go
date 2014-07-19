package main

import (
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"

	_ "code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/types"
	"honnef.co/go/importer"

	"github.com/opennota/check"
)

var (
	assignmentsOnly = flag.Bool("a", false, "Count assignments only")
	minimumUseCount = flag.Int("n", 1, "Minimum use count")
)

type visitor struct {
	pkg  *types.Package
	info types.Info
	m    map[string]map[string]int
}

func (v *visitor) decl(structName, fieldName string) {
	if _, ok := v.m[structName]; !ok {
		v.m[structName] = make(map[string]int)
	}
	if _, ok := v.m[structName][fieldName]; !ok {
		v.m[structName][fieldName] = 0
	}
}

func (v *visitor) assignment(structName, fieldName string) {
	if _, ok := v.m[structName]; !ok {
		v.m[structName] = make(map[string]int)
	}
	if _, ok := v.m[structName][fieldName]; ok {
		v.m[structName][fieldName]++
	} else {
		v.m[structName][fieldName] = 1
	}
}

func (v *visitor) typeSpec(node *ast.TypeSpec) {
	if strukt, ok := node.Type.(*ast.StructType); ok {
		structName := node.Name.Name
		for _, f := range strukt.Fields.List {
			if len(f.Names) > 0 {
				fieldName := f.Names[0].Name
				v.decl(structName, fieldName)
			}
		}
	}
}

func (v *visitor) names(expr *ast.SelectorExpr) (string, string, bool) {
	selection := v.info.Selections[expr]
	if selection == nil {
		return "", "", false
	}
	recv := selection.Recv()
	if ptr, ok := recv.(*types.Pointer); ok {
		recv = ptr.Elem()
	}
	return types.TypeString(v.pkg, recv), selection.Obj().Name(), true
}

func (v *visitor) assignStmt(node *ast.AssignStmt) {
	for _, lhs := range node.Lhs {
		var selector *ast.SelectorExpr
		switch expr := lhs.(type) {
		case *ast.SelectorExpr:
			selector = expr
		case *ast.IndexExpr:
			if expr, ok := expr.X.(*ast.SelectorExpr); ok {
				selector = expr
			}
		}
		if selector != nil {
			if sn, fn, ok := v.names(selector); ok {
				v.assignment(sn, fn)
			}
		}
	}
}

func (v *visitor) compositeLiteral(node *ast.CompositeLit) {
	t := v.info.Types[node.Type]
	for _, expr := range node.Elts {
		if kv, ok := expr.(*ast.KeyValueExpr); ok { // no support for positional field values yet
			if ident, ok := kv.Key.(*ast.Ident); ok {
				v.assignment(types.TypeString(v.pkg, t.Type), ident.Name)
			}
		}
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.TypeSpec:
		v.typeSpec(node)

	case *ast.AssignStmt:
		if *assignmentsOnly {
			v.assignStmt(node)
		}

	case *ast.SelectorExpr:
		if !*assignmentsOnly {
			if sn, fn, ok := v.names(node); ok {
				v.assignment(sn, fn)
			}
		}

	case *ast.CompositeLit:
		v.compositeLiteral(node)
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
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		},

		m: make(map[string]map[string]int),
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
	for structName := range visitor.m {
		for fieldName, v := range visitor.m[structName] {
			if v < *minimumUseCount {
				fmt.Printf("%s.%s\n", structName, fieldName)
				exitStatus = 1
			}
		}
	}
	os.Exit(exitStatus)
}

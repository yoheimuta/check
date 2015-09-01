// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"log"
	"os"

	"github.com/kisielk/gotool"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/types"
)

var (
	reportExported = flag.Bool("e", false, "Report exported variables and constants")
)

type visitor struct {
	prog       *loader.Program
	pkg        *loader.PackageInfo
	uses       map[types.Object]int
	insideFunc bool
}

func (v *visitor) decl(obj types.Object) {
	if _, ok := v.uses[obj]; !ok {
		v.uses[obj] = 0
	}
}

func (v *visitor) use(obj types.Object) {
	if _, ok := v.uses[obj]; ok {
		v.uses[obj]++
	} else {
		v.uses[obj] = 1
	}
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.Ident:
		v.use(v.pkg.Info.Uses[node])

	case *ast.ValueSpec:
		if !v.insideFunc {
			for _, ident := range node.Names {
				if ident.Name != "_" {
					v.decl(v.pkg.Info.Defs[ident])
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
	exitStatus := 0
	importPaths := gotool.ImportPaths(flag.Args())
	if len(importPaths) == 0 {
		importPaths = []string{"."}
	}

	ctx := build.Default
	loadcfg := loader.Config{
		Build: &ctx,
	}
	rest, err := loadcfg.FromArgs(importPaths, true)
	if err != nil {
		log.Fatalf("could not parse arguments: %s", err)
	}
	if len(rest) > 0 {
		log.Fatalf("unhandled extra arguments: %v", rest)
	}

	program, err := loadcfg.Load()
	if err != nil {
		log.Fatalf("could not type check: %s", err)
	}

	uses := make(map[types.Object]int)

	for _, pkgInfo := range program.InitialPackages() {
		if pkgInfo.Pkg.Path() == "unsafe" {
			continue
		}

		v := &visitor{
			prog: program,
			pkg:  pkgInfo,
			uses: uses,
		}

		for _, f := range v.pkg.Files {
			ast.Walk(v, f)
		}
	}

	for obj, useCount := range uses {
		if useCount == 0 && (*reportExported || !ast.IsExported(obj.Name())) {
			pos := program.Fset.Position(obj.Pos())
			fmt.Printf("%s: %s:%d:%d: %s\n", obj.Pkg().Path(), pos.Filename, pos.Line, pos.Column, obj.Name())
			exitStatus = 1
		}
	}
	os.Exit(exitStatus)
}

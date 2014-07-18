package check

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"reflect"
)

func IsExported(ident string) bool {
	return ident[0] >= 'A' && ident[0] <= 'Z'
}

func ASTFilesForPackage(path string) (*token.FileSet, []*ast.File) {
	ctx := build.Default
	pkg, err := ctx.Import(path, ".", 0)
	if err != nil {
		var err2 error
		pkg, err2 = ctx.ImportDir(path, 0)
		if err2 != nil {
			log.Fatalf("cannot import package %s\n"+
				"Errors are:\n"+
				"    %s\n"+
				"    %s",
				path, err, err2)
		}
	}
	fset := token.NewFileSet()
	var astFiles []*ast.File
	for _, f := range pkg.GoFiles {
		fn := filepath.Join(pkg.Dir, f)
		f, err := parser.ParseFile(fset, fn, nil, 0)
		if err != nil {
			log.Fatalf("cannot parse file '%s'\n"+
				"Error: %s", fn, err)
		}
		astFiles = append(astFiles, f)
	}
	return fset, astFiles
}

func TypeName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

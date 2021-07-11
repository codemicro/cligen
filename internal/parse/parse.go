package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
)

type Parsed struct {
	PackageName string
	Functions map[string]*Function
}

func Directory(dir string) (*Parsed, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, func(info fs.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".cligen.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	if len(packages) != 1 {
		return nil, errors.New("input directory must only have source files with at most a single package in it") // TODO: This error is phrased weird.
	}

	var pkg *ast.Package
	for _, x := range packages {
		pkg = x
		break
	}

	functions, err := getFunctionsFromPackage(pkg)
	if err != nil {
		return nil, err
	}

	return &Parsed{
		PackageName: pkg.Name,
		Functions:   functions,
	}, nil
}

type Function struct {
	Name string
	Directive string
	Signature *Signature
}

func getFunctionsFromPackage(pkg *ast.Package) (map[string]*Function, error) {
	functions := make(map[string]*Function)

	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {

			if funcDecl, ok := declaration.(*ast.FuncDecl); ok {

				// if the function has a receiver, ignore it
				if funcDecl.Recv != nil {
					continue
				}

				// if a comment is included in one of the lines before a function, it will be included here
				if funcDecl.Doc == nil {
					continue
				}

				directiveText, err := getDirective(funcDecl.Doc)
				if err != nil {
					if errors.Is(err, errorNoDirective) {
						continue
					} else {
						return nil, fmt.Errorf("%s:%s: %s", pkg.Name, funcDecl.Name.String(), err.Error())
					}
				}

				s := funcDecl.Name.String()
				functions[s] = &Function{
					Name:      s,
					Directive: directiveText,
					Signature: signatureFromDeclaration(funcDecl),
				}
			}

		}
	}

	return functions, nil
}

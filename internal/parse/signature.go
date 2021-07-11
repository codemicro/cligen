package parse

import (
	"fmt"
	"go/ast"
)

type Signature struct {
	Argument []*Param
	Return []*Param
}

type Param struct {
	Name string
	Type string
	IsPointer bool
}

func signatureFromDeclaration(f *ast.FuncDecl) *Signature {
	return &Signature{
		Argument: unwrapFieldList(f.Type.Params),
		Return:   unwrapFieldList(f.Type.Results),
	}
}

func unwrapFieldList(list *ast.FieldList) []*Param {
	if list == nil {
		return nil
	}

	var params []*Param
	for _, item := range list.List {

		var (
			typeName  string
			isPointer bool
		)

		{
			var ident *ast.Ident

			switch x := item.Type.(type) {
			case *ast.Ident:
				ident = x
			case *ast.StarExpr:
				isPointer = true
				ident = x.X.(*ast.Ident)
			default:
				panic(fmt.Errorf("unknown type %T", item.Type))
			}

			typeName = ident.Name
		}

		if len(item.Names) == 0 {
			// if there are no names associated with this type, we still need to know about it
			params = append(params, &Param{
				Type:      typeName,
				IsPointer: isPointer,
			})
		} else {
			for _, name := range item.Names {
				params = append(params, &Param{
					Name:      name.Name,
					Type:      typeName,
					IsPointer: isPointer,
				})
			}
		}
	}
	return params
}
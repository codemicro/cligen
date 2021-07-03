package parse

import (
	"fmt"
	"go/ast"
)

type signature struct {
	Argument []*param
	Return []*param
}

type param struct {
	Name string
	Type string
	IsPointer bool
}

func signatureFromDeclaration(f *ast.FuncDecl) *signature {
	return &signature{
		Argument: unwrapFieldList(f.Type.Params),
		Return:   unwrapFieldList(f.Type.Results),
	}
}

func unwrapFieldList(list *ast.FieldList) []*param {
	if list == nil {
		return nil
	}

	var params []*param
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
			params = append(params, &param{
				Type:      typeName,
				IsPointer: isPointer,
			})
		} else {
			for _, name := range item.Names {
				params = append(params, &param{
					Name:      name.Name,
					Type:      typeName,
					IsPointer: isPointer,
				})
			}
		}
	}
	return params
}
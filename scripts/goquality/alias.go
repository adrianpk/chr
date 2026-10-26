package main

import (
	"fmt"
	"go/ast"
)

type typeBinding struct {
	key        string
	expression ast.Expr
	defined    bool
}

func (scan *sourceScan) aliasBinding(identifier *ast.Ident) typeBinding {
	key := scan.packagePath + "/" + identifier.Name
	if identifier.Obj == nil {
		return typeBinding{key: key, expression: scan.aliases[key], defined: scan.defined[key]}
	}

	spec, ok := identifier.Obj.Decl.(*ast.TypeSpec)
	if !ok || !spec.Assign.IsValid() {
		return typeBinding{key: key, defined: true}
	}

	return typeBinding{key: fmt.Sprintf("%s/%d", key, spec.Pos()), expression: spec.Type}
}

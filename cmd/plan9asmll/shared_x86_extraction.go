package main

import (
	"go/ast"
	"go/token"
	"strconv"
)

// Shared effect grammars also contain operand/condition switches. Observe
// actual opcode-derived dispatch, not every string found in an x86 file.
func collectSharedX86OpcodeDispatch(supported map[string]struct{}, file *ast.File) {
	add := func(expression ast.Expr) {
		literal, ok := expression.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return
		}
		name, err := strconv.Unquote(literal.Value)
		if err == nil {
			if name = normalizeOp(name); name != "" {
				supported[name] = struct{}{}
			}
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SwitchStmt:
			if sharedX86OpcodeExpression(value.Tag, 0) {
				for _, statement := range value.Body.List {
					for _, expression := range statement.(*ast.CaseClause).List {
						add(expression)
					}
				}
			}
		case *ast.BinaryExpr:
			if value.Op == token.EQL || value.Op == token.NEQ {
				if sharedX86OpcodeExpression(value.X, 0) {
					add(value.Y)
				} else if sharedX86OpcodeExpression(value.Y, 0) {
					add(value.X)
				}
			}
		case *ast.CompositeLit:
			mapType, ok := value.Type.(*ast.MapType)
			if key, valid := mapTypeKeyIdent(mapType); ok && valid && key == "Op" {
				collectSupportedOpcodeMap(supported, value)
			}
		case *ast.IndexExpr:
			if !sharedX86OpcodeExpression(value.Index, 0) {
				break
			}
			name, ok := value.X.(*ast.Ident)
			if !ok || name.Obj == nil {
				break
			}
			declaration, ok := name.Obj.Decl.(*ast.ValueSpec)
			if !ok || len(declaration.Values) != 1 {
				break
			}
			if literal, ok := declaration.Values[0].(*ast.CompositeLit); ok {
				collectSupportedOpcodeMap(supported, literal)
			}
		}
		return true
	})
}

// Follow parser-bound aliases within a finite expression, including the first
// result of strings.Cut. Other tuple results and unrelated identifiers never
// gain an opcode role just because their names resemble instruction names.
func sharedX86OpcodeExpression(expression ast.Expr, depth int) bool {
	if depth > 8 {
		return false
	}
	switch value := expression.(type) {
	case *ast.Ident:
		if value.Name == "op" {
			return true
		}
		if value.Obj == nil {
			return false
		}
		assignment, ok := value.Obj.Decl.(*ast.AssignStmt)
		if !ok || len(assignment.Rhs) != 1 || len(assignment.Lhs) == 0 {
			return false
		}
		first, ok := assignment.Lhs[0].(*ast.Ident)
		return ok && first.Obj == value.Obj && sharedX86OpcodeExpression(assignment.Rhs[0], depth+1)
	case *ast.ParenExpr:
		return sharedX86OpcodeExpression(value.X, depth+1)
	case *ast.SelectorExpr:
		return value.Sel.Name == "Op"
	case *ast.CallExpr:
		if name, ok := value.Fun.(*ast.Ident); ok && name.Name == "string" && len(value.Args) == 1 {
			return sharedX86OpcodeExpression(value.Args[0], depth+1)
		}
		function, ok := value.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := function.X.(*ast.Ident)
		if !ok || pkg.Name != "strings" {
			return false
		}
		switch function.Sel.Name {
		case "ToUpper", "ToLower", "TrimSpace":
			return len(value.Args) == 1 && sharedX86OpcodeExpression(value.Args[0], depth+1)
		case "Cut":
			if len(value.Args) != 2 {
				return false
			}
			separator, ok := value.Args[1].(*ast.BasicLit)
			if !ok || separator.Kind != token.STRING {
				return false
			}
			text, err := strconv.Unquote(separator.Value)
			return err == nil && text == "." && sharedX86OpcodeExpression(value.Args[0], depth+1)
		}
	}
	return false
}

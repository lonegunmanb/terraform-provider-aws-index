package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"go/ast"
	"go/token"
	"strings"
)







// findResourceFunction locates any function declaration that returns *pluginsdk.Resource
func findResourceFunction(node *ast.File) *ast.FuncDecl {
	var resourceFunc *ast.FuncDecl

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Type.Results == nil {
			return true
		}

		// Check if function returns *pluginsdk.Resource
		for _, result := range fn.Type.Results.List {
			starExpr, ok := result.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			ident, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				continue
			}
			if ident.Name == "pluginsdk" && selectorExpr.Sel.Name == "Resource" {
				resourceFunc = fn
				return false // Found it, stop searching
			}
		}
		return true
	})

	return resourceFunc
}

// extractCRUDFromFunction extracts CRUD method names from a resource function body
func extractCRUDFromFunction(fn *ast.FuncDecl) *LegacyResourceCRUDFunctions {
	methods := &LegacyResourceCRUDFunctions{}

	if fn.Body == nil {
		return methods
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		// Look for return statements
		returnStmt, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		// Process each return expression
		for _, result := range returnStmt.Results {
			unaryExpr, ok := result.(*ast.UnaryExpr)
			// Handle direct return of composite literal
			if !ok || unaryExpr.Op != token.AND {
				return true
			}
			if compLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
				extractFromResourceLiteral(compLit, methods)
			}
		}

		return true
	})

	// Also look for variable assignments in case of Pattern 2 (variable assignment)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assignStmt, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for _, rhs := range assignStmt.Rhs {
			unaryExpr, ok := rhs.(*ast.UnaryExpr)
			if !ok || unaryExpr.Op != token.AND {
				return true
			}
			if compLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
				extractFromResourceLiteral(compLit, methods)
			}
		}

		return true
	})

	return methods
}

// extractFunctionReference extracts function name from various AST patterns:
// - Direct identifier: funcName
// - Selector expression: package.FuncName
func extractFunctionReference(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		// Direct identifier: funcName
		return e.Name
	case *ast.SelectorExpr:
		// Selector expression: package.FuncName
		return e.Sel.Name
	default:
		return ""
	}
}

func mergeMap[TK comparable, TV any](m1, m2 map[TK]TV) map[TK]TV {
	m := make(map[TK]TV)
	for tk, tv := range m1 {
		m[tk] = tv
	}
	for tk, tv := range m2 {
		m[tk] = tv
	}
	return m
}

// extractResourceTerraformTypes extracts Terraform types from ResourceType methods for each resource struct
func extractResourceTerraformTypes(packageInfo *gophon.PackageInfo, resourceStructs []string) map[string]string {
	terraformTypes := make(map[string]string)

	for _, structName := range resourceStructs {
		if terraformType := extractTerraformTypeFromResourceTypeMethod(packageInfo, structName); terraformType != "" {
			terraformTypes[structName] = terraformType
		}
	}

	return terraformTypes
}

// extractDataSourceTerraformTypes extracts Terraform types from ResourceType methods for each data source struct
func extractDataSourceTerraformTypes(packageInfo *gophon.PackageInfo, dataSourceStructs []string) map[string]string {
	terraformTypes := make(map[string]string)

	for _, structName := range dataSourceStructs {
		if terraformType := extractTerraformTypeFromResourceTypeMethod(packageInfo, structName); terraformType != "" {
			terraformTypes[structName] = terraformType
		}
	}

	return terraformTypes
}

// extractEphemeralTerraformTypes extracts Terraform types from Metadata methods for each ephemeral struct
func extractEphemeralTerraformTypes(packageInfo *gophon.PackageInfo, ephemeralStructs []string) map[string]string {
	terraformTypes := make(map[string]string)

	for _, structName := range ephemeralStructs {
		if terraformType := extractTerraformTypeFromMetadataMethod(packageInfo, structName); terraformType != "" {
			terraformTypes[structName] = terraformType
		}
	}

	return terraformTypes
}

// extractTerraformTypeFromResourceTypeMethod extracts terraform type from ResourceType method of a struct
func extractTerraformTypeFromResourceTypeMethod(packageInfo *gophon.PackageInfo, structName string) string {
	for _, fileInfo := range packageInfo.Files {
		if fileInfo.File == nil {
			continue
		}

		var result string
		// Look for ResourceType method on the struct
		ast.Inspect(fileInfo.File, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "ResourceType" {
				return true
			}

			// Check if this method belongs to our struct
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				var receiverTypeName string

				// Handle both pointer receiver (*StructName) and value receiver (StructName)
				switch recvType := fn.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					// Pointer receiver: *StructName
					if ident, ok := recvType.X.(*ast.Ident); ok {
						receiverTypeName = ident.Name
					}
				case *ast.Ident:
					// Value receiver: StructName
					receiverTypeName = recvType.Name
				}

				if receiverTypeName == structName {
					// Found the ResourceType method for our struct
					result = extractStringReturnValue(fn)
					return false // Stop traversing
				}
			}
			return true
		})
		if result != "" {
			return result
		}
	}
	return ""
}

// extractTerraformTypeFromMetadataMethod extracts terraform type from Metadata method of an ephemeral struct
func extractTerraformTypeFromMetadataMethod(packageInfo *gophon.PackageInfo, structName string) string {
	for _, fileInfo := range packageInfo.Files {
		if fileInfo.File == nil {
			continue
		}

		var result string
		// Look for Metadata method on the struct
		ast.Inspect(fileInfo.File, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "Metadata" {
				return true
			}

			// Check if this method belongs to our struct
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				var receiverTypeName string

				// Handle both pointer receiver (*StructName) and value receiver (StructName)
				switch recvType := fn.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					// Pointer receiver: *StructName
					if ident, ok := recvType.X.(*ast.Ident); ok {
						receiverTypeName = ident.Name
					}
				case *ast.Ident:
					// Value receiver: StructName
					receiverTypeName = recvType.Name
				}

				if receiverTypeName == structName {
					// Found the Metadata method for our struct
					result = extractTypeNameFromMetadataMethod(fn)
					return false // Stop traversing
				}
			}
			return true
		})
		if result != "" {
			return result
		}
	}
	return ""
}

// extractStringReturnValue extracts a string literal return value from a function
func extractStringReturnValue(fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}

	for _, stmt := range fn.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		if len(retStmt.Results) == 0 {
			continue
		}
		basicLit, ok := retStmt.Results[0].(*ast.BasicLit)
		if !ok {
			continue
		}
		if basicLit.Kind == token.STRING {
			// Remove quotes from string literal
			return strings.Trim(basicLit.Value, `"`)
		}
	}
	return ""
}

// extractTypeNameFromMetadataMethod extracts TypeName assignment from Metadata method, used by ephemeral
func extractTypeNameFromMetadataMethod(fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}

	for _, stmt := range fn.Body.List {
		assignStmt, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		// Look for resp.TypeName = "something"
		if len(assignStmt.Lhs) == 0 || len(assignStmt.Rhs) == 0 {
			continue
		}
		selectorExpr, ok := assignStmt.Lhs[0].(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if selectorExpr.Sel.Name != "TypeName" {
			continue
		}
		basicLit, ok := assignStmt.Rhs[0].(*ast.BasicLit)
		if !ok {
			continue
		}
		if basicLit.Kind == token.STRING {
			// Remove quotes from string literal
			return strings.Trim(basicLit.Value, `"`)
		}
	}
	return ""
}

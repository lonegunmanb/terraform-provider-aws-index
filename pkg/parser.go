package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"go/ast"
	"go/token"
	"strings"
)

// extractSupportedResourcesMappings extracts mappings from SupportedResources method in the AST
func extractSupportedResourcesMappings(node *ast.File) map[string]string {
	return extractMappingsFromMethod(node, "SupportedResources")
}

// extractSupportedDataSourcesMappings extracts mappings from SupportedDataSources method in the AST
func extractSupportedDataSourcesMappings(node *ast.File) map[string]string {
	return extractMappingsFromMethod(node, "SupportedDataSources")
}

// extractDataSourcesStructTypes extracts struct type names from DataSources method in the AST
func extractDataSourcesStructTypes(node *ast.File) []string {
	return extractStructTypesFromMethod(node, "DataSources")
}

// extractResourcesStructTypes extracts struct type names from Resources method in the AST
func extractResourcesStructTypes(node *ast.File) []string {
	return extractStructTypesFromMethod(node, "Resources")
}

// extractEphemeralResourcesFunctions extracts function names from EphemeralResources method in the AST
func extractEphemeralResourcesFunctions(node *ast.File) []string {
	return extractFunctionNamesFromMethod(node, "EphemeralResources")
}

// extractMappingsFromMethod extracts mappings from any method that returns map[string]*pluginsdk.Resource
func extractMappingsFromMethod(node *ast.File, methodName string) map[string]string {
	mappings := make(map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != methodName {
			return true
		}

		// Look for return statements in the function body
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			// Process each return expression
			for _, result := range returnStmt.Results {
				// Handle direct map literal return
				if mapLit, ok := result.(*ast.CompositeLit); ok {
					mappings = mergeMap(mappings, extractFromMapLiteral(mapLit))
				}

				// Handle variable reference (like "resources" or "dataSources" variable)
				ident, ok := result.(*ast.Ident)
				if !ok {
					continue
				}
				// Find the variable definition in the function
				ast.Inspect(fn.Body, func(varNode ast.Node) bool {
					assignStmt, ok := varNode.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for i, lhs := range assignStmt.Lhs {
						lhsIdent, ok := lhs.(*ast.Ident)
						if !ok || lhsIdent.Name != ident.Name {
							return true
						}
						if i >= len(assignStmt.Rhs) {
							return true
						}
						if mapLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit); ok {
							mappings = mergeMap(mappings, extractFromMapLiteral(mapLit))
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return mappings
}

// extractStructTypesFromMethod extracts struct type names from any method that returns []sdk.DataSource or []sdk.Resource
func extractStructTypesFromMethod(node *ast.File, methodName string) []string {
	var types []string

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != methodName {
			return true
		}

		// Look for return statements in the function body
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			// Process each return expression
			for _, result := range returnStmt.Results {
				// Handle direct slice literal return
				if sliceLit, ok := result.(*ast.CompositeLit); ok {
					types = append(types, extractFromSliceLiteral(sliceLit)...)
				}

				// Handle variable reference (like "dataSources" variable)
				ident, ok := result.(*ast.Ident)
				if !ok {
					continue
				}
				// Find the variable definition in the function
				ast.Inspect(fn.Body, func(varNode ast.Node) bool {
					assignStmt, ok := varNode.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for i, lhs := range assignStmt.Lhs {
						lhsIdent, ok := lhs.(*ast.Ident)
						if !ok || lhsIdent.Name != ident.Name {
							return true
						}
						if i >= len(assignStmt.Rhs) {
							return true
						}
						if sliceLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit); ok {
							types = append(types, extractFromSliceLiteral(sliceLit)...)
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return types
}

// extractFunctionNamesFromMethod extracts function names from any method that returns []func() ephemeral.EphemeralResource
func extractFunctionNamesFromMethod(node *ast.File, methodName string) []string {
	var functions []string

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != methodName {
			return true
		}

		// Look for return statements in the function body
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			// Process each return expression
			for _, result := range returnStmt.Results {
				// Handle direct slice literal return
				if sliceLit, ok := result.(*ast.CompositeLit); ok {
					functions = append(functions, extractFromFunctionSliceLiteral(sliceLit)...)
				}

				// Handle variable reference (like "ephemeralResources" variable)
				ident, ok := result.(*ast.Ident)
				if !ok {
					continue
				}
				// Find the variable definition in the function
				ast.Inspect(fn.Body, func(varNode ast.Node) bool {
					assignStmt, ok := varNode.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for i, lhs := range assignStmt.Lhs {
						lhsIdent, ok := lhs.(*ast.Ident)
						if !ok || lhsIdent.Name != ident.Name {
							return true
						}
						if i >= len(assignStmt.Rhs) {
							return true
						}
						if sliceLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit); ok {
							functions = append(functions, extractFromFunctionSliceLiteral(sliceLit)...)
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return functions
}

// extractFromMapLiteral extracts key-value pairs from a map literal
func extractFromMapLiteral(mapLit *ast.CompositeLit) map[string]string {
	mappings := make(map[string]string)
	for _, elt := range mapLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		// Extract the key (terraform resource type)
		var key string
		if keyLit, ok := kv.Key.(*ast.BasicLit); ok && keyLit.Kind == token.STRING {
			key = strings.Trim(keyLit.Value, `"`)
		}

		// Extract the value (function call name)
		var value string
		callExpr, ok := kv.Value.(*ast.CallExpr)
		if !ok {
			continue
		}
		if fnIdent, ok := callExpr.Fun.(*ast.Ident); ok {
			value = fnIdent.Name
		}

		if key != "" && value != "" {
			mappings[key] = value
		}
	}
	return mappings
}

// extractFromSliceLiteral extracts struct type names from a slice literal
func extractFromSliceLiteral(sliceLit *ast.CompositeLit) []string {
	var types []string
	for _, elt := range sliceLit.Elts {
		// Handle struct literals like StructName{}
		compLit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}

		// Extract the struct type name
		if ident, ok := compLit.Type.(*ast.Ident); ok {
			types = append(types, ident.Name)
		}
	}
	return types
}

// extractFromFunctionSliceLiteral extracts function names from a slice literal
func extractFromFunctionSliceLiteral(sliceLit *ast.CompositeLit) []string {
	var functions []string
	for _, elt := range sliceLit.Elts {
		// Handle function identifiers like FuncName (without parentheses)
		if ident, ok := elt.(*ast.Ident); ok {
			functions = append(functions, ident.Name)
		}
	}
	return functions
}

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

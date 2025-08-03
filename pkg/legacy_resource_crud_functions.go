package pkg

import (
	pkg2 "github.com/lonegunmanb/gophon/pkg"
	"go/ast"
)

// LegacyResourceCRUDFunctions represents CRUD methods extracted from legacy plugin SDK resources
type LegacyResourceCRUDFunctions struct {
	CreateMethod string `json:"create_method,omitempty"` // "keyVaultCreateFunc"
	ReadMethod   string `json:"read_method,omitempty"`   // "keyVaultReadFunc"
	UpdateMethod string `json:"update_method,omitempty"` // "keyVaultUpdateFunc"
	DeleteMethod string `json:"delete_method,omitempty"` // "keyVaultDeleteFunc"
}

// extractLegacyResourceCRUDMethods analyzes a legacy plugin SDK resource function
// and extracts CRUD method names from the returned pluginsdk.Resource struct
// The input ast.File should contain the registration function's source code
// It will find any function that returns *pluginsdk.Resource and parse its CRUD methods
func extractLegacyResourceCRUDMethods(node *ast.File) (*LegacyResourceCRUDFunctions, error) {
	// Find the resource function that returns *pluginsdk.Resource
	resourceFunc := findResourceFunction(node)
	if resourceFunc == nil {
		return &LegacyResourceCRUDFunctions{}, nil // No resource function found, return empty
	}

	// Extract CRUD methods from the function body
	crudMethods := extractCRUDFromFunction(resourceFunc)
	return crudMethods, nil
}

// extractFromResourceLiteral parses a pluginsdk.Resource composite literal
// and extracts only CRUD method names
func extractFromResourceLiteral(compLit *ast.CompositeLit, methods *LegacyResourceCRUDFunctions) {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Get the field name
		var fieldName string
		if ident, ok := kv.Key.(*ast.Ident); ok {
			fieldName = ident.Name
		}

		// Extract function reference from the value
		funcName := extractFunctionReference(kv.Value)
		if funcName == "" {
			continue
		}

		// Map field names to CRUD methods
		switch fieldName {
		case "Create", "CreateContext", "CreateFunc", "CreateWithoutTimeout":
			methods.CreateMethod = funcName
		case "Read", "ReadContext", "ReadFunc", "ReadWithoutTimeout":
			methods.ReadMethod = funcName
		case "Update", "UpdateContext", "UpdateFunc", "UpdateWithoutTimeout":
			methods.UpdateMethod = funcName
		case "Delete", "DeleteContext", "DeleteFunc", "DeleteWithoutTimeout":
			methods.DeleteMethod = funcName
		}
	}
}

// extractCRUDFromPackage extracts CRUD methods from a gophon PackageInfo by finding the registration function
func extractCRUDFromPackage(registrationMethod string, packageInfo *pkg2.PackageInfo) *LegacyResourceCRUDFunctions {
	if packageInfo == nil || packageInfo.Functions == nil {
		return nil
	}

	// Find the registration function in the gophon function data
	for _, funcInfo := range packageInfo.Functions {
		if funcInfo.Name == registrationMethod && funcInfo.FuncDecl != nil {
			return extractCRUDFromFunction(funcInfo.FuncDecl)
		}
	}

	return nil
}

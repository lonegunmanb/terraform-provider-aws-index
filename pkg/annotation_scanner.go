package pkg

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

// annotationRegex matches Terraform provider annotations
// Examples:
// @SDKResource("aws_lambda_function", name="Function")
// @FrameworkDataSource("aws_bedrock_custom_model", name="Custom Model")
var annotationRegex = regexp.MustCompile(`@(SDKResource|SDKDataSource|FrameworkResource|FrameworkDataSource|EphemeralResource)\("([^"]+)"(?:,\s*name="([^"]+)")?[^)]*\)`)

// ScanPackageForAnnotations scans all files in the package for annotations
// and returns structured results mapping annotations to their context
func ScanPackageForAnnotations(packageInfo *gophon.PackageInfo) (*AnnotationResults, error) {
	results := NewAnnotationResults()

	// Scan each file in the package
	for _, fileInfo := range packageInfo.Files {
		fileResults, err := scanFileForAnnotations(fileInfo)
		if err != nil {
			// Log error but continue with other files
			continue
		}

		// Merge file results into package results
		for _, result := range fileResults {
			results.Add(result)
		}
	}

	return results, nil
}

// scanFileForAnnotations scans a single Go file for annotations and extracts relevant info
func scanFileForAnnotations(fileInfo *gophon.FileInfo) ([]AnnotationResult, error) {
	var results []AnnotationResult

	if fileInfo.File == nil {
		return results, fmt.Errorf("no AST available for file %s", fileInfo.FilePath)
	}

	// First, scan for any annotations in the file
	annotations := findAnnotationsInFile(fileInfo.File)
	if len(annotations) == 0 {
		return results, nil // No annotations found
	}

	// For each annotation found, extract the full context from the file
	for _, annotation := range annotations {
		result := AnnotationResult{
			Type:          annotation.Type,
			TerraformType: annotation.TerraformType,
			Name:          annotation.Name,
			FilePath:      fileInfo.FilePath,
			RawAnnotation: annotation.RawAnnotation,
		}

		// Extract type-specific information from the file
		switch annotation.Type {
		case AnnotationSDKResource:
			result.CRUDMethods = extractSDKResourceCRUDFromFile(fileInfo.File)
		case AnnotationSDKDataSource:
			result.CRUDMethods = extractSDKDataSourceMethodsFromFile(fileInfo.File)
		case AnnotationFrameworkResource, AnnotationFrameworkDataSource, AnnotationEphemeralResource:
			// Find struct type by Schema method - the struct that implements framework interfaces
			result.StructType = extractFrameworkStructTypeBySchemaMethod(fileInfo.File)
			result.FrameworkMethods = inferFrameworkMethods(annotation.Type)
		}

		results = append(results, result)
	}

	return results, nil
}

// basicAnnotation represents a simple annotation found in comments
type basicAnnotation struct {
	Type          AnnotationType
	TerraformType string
	Name          string
	RawAnnotation string
	FunctionName  string // Added to track which function has the annotation
}

// findAnnotationsInFile searches for annotations in all function comments in the file
func findAnnotationsInFile(file *ast.File) []basicAnnotation {
	var annotations []basicAnnotation

	// Walk through all declarations looking for function comments
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Doc == nil {
			continue
		}

		// Combine all comment lines
		var commentText strings.Builder
		for _, comment := range funcDecl.Doc.List {
			commentText.WriteString(comment.Text)
			commentText.WriteString("\n")
		}

		// Search for annotation patterns
		matches := annotationRegex.FindStringSubmatch(commentText.String())
		if len(matches) < 3 {
			continue
		}

		annotationType := matches[1]
		terraformType := matches[2]
		name := ""
		if len(matches) > 3 && matches[3] != "" {
			name = matches[3]
		}

		// Convert to enum type
		var annoType AnnotationType
		switch annotationType {
		case "SDKResource":
			annoType = AnnotationSDKResource
		case "SDKDataSource":
			annoType = AnnotationSDKDataSource
		case "FrameworkResource":
			annoType = AnnotationFrameworkResource
		case "FrameworkDataSource":
			annoType = AnnotationFrameworkDataSource
		case "EphemeralResource":
			annoType = AnnotationEphemeralResource
		default:
			continue // Skip unknown annotations
		}

		annotations = append(annotations, basicAnnotation{
			Type:          annoType,
			TerraformType: terraformType,
			Name:          name,
			RawAnnotation: matches[0],
			FunctionName:  funcDecl.Name.Name, // Capture the function name
		})
	}

	return annotations
}

// extractSDKResourceCRUDFromFile extracts CRUD method names from SDK resource files
func extractSDKResourceCRUDFromFile(file *ast.File) map[string]string {
	methods := make(map[string]string)

	// Look for function that returns *schema.Resource and extract CRUD methods
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		// Look for return statements that return &schema.Resource{...}
		ast.Inspect(funcDecl.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			for _, result := range returnStmt.Results {
				unaryExpr, ok := result.(*ast.UnaryExpr)
				if !ok || unaryExpr.Op != token.AND {
					continue
				}
				if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
					extractCRUDFromCompositeLit(compositeLit, methods)
				}
			}
			return true
		})

		return true
	})

	return methods
}

// extractCRUDFromCompositeLit extracts CRUD methods from &schema.Resource{...} composite literal
func extractCRUDFromCompositeLit(compositeLit *ast.CompositeLit, methods map[string]string) {
	for _, elt := range compositeLit.Elts {
		keyValue, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		ident, ok := keyValue.Key.(*ast.Ident)
		if !ok {
			continue
		}

		var methodType string
		if strings.HasPrefix(ident.Name, "Create") {
			methodType = "create"
		} else if strings.HasPrefix(ident.Name, "Read") {
			methodType = "read"
		} else if strings.HasPrefix(ident.Name, "Update") {
			methodType = "update"
		} else if strings.HasPrefix(ident.Name, "Delete") {
			methodType = "delete"
		} else {
			continue
		}

		// Extract function name - handle both identifiers and selector expressions
		switch value := keyValue.Value.(type) {
		case *ast.Ident:
			methods[methodType] = value.Name
		case *ast.SelectorExpr:
			// Handle cases like schema.NoopContext
			pkgIdent, ok := value.X.(*ast.Ident)
			if !ok {
				continue
			}

			fullName := pkgIdent.Name + "." + value.Sel.Name
			// Skip special schema functions as requested - leave these empty
			if !strings.HasPrefix(fullName, "schema.") {
				methods[methodType] = fullName
			}
			// For schema.* cases, we intentionally don't add anything to methods
		}
	}
}

// extractSDKDataSourceMethodsFromFile extracts read method from SDK data source files
func extractSDKDataSourceMethodsFromFile(file *ast.File) map[string]string {
	methods := make(map[string]string)

	// Similar to SDK resources but only looking for Read method
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			return true
		}

		ast.Inspect(funcDecl.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			for _, result := range returnStmt.Results {
				unaryExpr, ok := result.(*ast.UnaryExpr)
				if !ok || unaryExpr.Op != token.AND {
					continue
				}

				compositeLit, ok := unaryExpr.X.(*ast.CompositeLit)
				if !ok {
					continue
				}

				for _, elt := range compositeLit.Elts {
					keyValue, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}

					ident, ok := keyValue.Key.(*ast.Ident)
					if !ok || !strings.HasPrefix(ident.Name, "Read") {
						continue
					}

					if valueIdent, ok := keyValue.Value.(*ast.Ident); ok {
						methods["read"] = valueIdent.Name
					}
				}
			}
			return true
		})

		return true
	})

	return methods
}

// extractFrameworkStructTypeBySchemaMethod finds the struct type that has a Schema method
// This is the actual resource/datasource/ephemeral struct that implements the framework interfaces
func extractFrameworkStructTypeBySchemaMethod(file *ast.File) string {
	// Find all structs with Schema methods
	structsWithSchema := findStructsWithSchemaMethod(file)

	// If we found exactly one struct with a Schema method, return it
	if len(structsWithSchema) == 1 {
		return structsWithSchema[0]
	}

	// If multiple structs have Schema methods, prefer the one that embeds framework types
	for _, structName := range structsWithSchema {
		if structEmbedsFramework(file, structName) {
			return structName
		}
	}

	// Fallback: return the first one found (shouldn't happen in well-formed code)
	if len(structsWithSchema) > 0 {
		return structsWithSchema[0]
	}

	return ""
}

// findStructsWithSchemaMethod finds all struct types that have a Schema method
func findStructsWithSchemaMethod(file *ast.File) []string {
	var structs []string
	structSet := make(map[string]bool)

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || funcDecl.Name.Name != "Schema" {
			return true
		}

		// Check if this is a method on a struct (receiver should be a pointer to a struct)
		if len(funcDecl.Recv.List) == 0 {
			return true
		}

		field := funcDecl.Recv.List[0]
		starExpr, ok := field.Type.(*ast.StarExpr)
		if !ok {
			return true
		}

		ident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		if !structSet[ident.Name] {
			structSet[ident.Name] = true
			structs = append(structs, ident.Name)
		}

		return true
	})

	return structs
}

// structEmbedsFramework checks if a struct embeds framework types
func structEmbedsFramework(file *ast.File, structName string) bool {
	var embedsFramework bool

	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != structName {
				continue
			}

			structTypeDecl, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Check if this struct embeds framework types
			for _, field := range structTypeDecl.Fields.List {
				if len(field.Names) != 0 { // Not an embedded field
					continue
				}

				selectorExpr, ok := field.Type.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				ident, ok := selectorExpr.X.(*ast.Ident)
				if !ok || ident.Name != "framework" {
					continue
				}

				embedsFramework = true
				return false
			}
		}

		return true
	})

	return embedsFramework
}

// inferFrameworkMethods returns expected methods based on annotation type
func inferFrameworkMethods(annoType AnnotationType) []string {
	switch annoType {
	case AnnotationFrameworkResource:
		return []string{"Create", "Read", "Update", "Delete", "Metadata", "Schema"}
	case AnnotationFrameworkDataSource:
		return []string{"Read", "Metadata", "Schema"}
	case AnnotationEphemeralResource:
		return []string{"Open", "Close", "Renew", "Metadata", "Schema"}
	default:
		return []string{}
	}
}

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
			result.StructType = extractFrameworkStructTypeFromFile(fileInfo.File)
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
		if len(matches) >= 3 {
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
			})
		}
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
				if unaryExpr, ok := result.(*ast.UnaryExpr); ok && unaryExpr.Op == token.AND {
					if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
						extractCRUDFromCompositeLit(compositeLit, methods)
					}
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
		if keyValue, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := keyValue.Key.(*ast.Ident); ok {
				var methodType string
				switch ident.Name {
				case "CreateWithoutTimeout", "Create":
					methodType = "create"
				case "ReadWithoutTimeout", "Read":
					methodType = "read"
				case "UpdateWithoutTimeout", "Update":
					methodType = "update"
				case "DeleteWithoutTimeout", "Delete":
					methodType = "delete"
				default:
					continue
				}

				// Extract function name
				if valueIdent, ok := keyValue.Value.(*ast.Ident); ok {
					methods[methodType] = valueIdent.Name
				}
			}
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
				if unaryExpr, ok := result.(*ast.UnaryExpr); ok && unaryExpr.Op == token.AND {
					if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
						for _, elt := range compositeLit.Elts {
							if keyValue, ok := elt.(*ast.KeyValueExpr); ok {
								if ident, ok := keyValue.Key.(*ast.Ident); ok {
									if ident.Name == "ReadWithoutTimeout" || ident.Name == "Read" {
										if valueIdent, ok := keyValue.Value.(*ast.Ident); ok {
											methods["read"] = valueIdent.Name
										}
									}
								}
							}
						}
					}
				}
			}
			return true
		})

		return true
	})

	return methods
}

// extractFrameworkStructTypeFromFile finds struct types that embed framework interfaces
func extractFrameworkStructTypeFromFile(file *ast.File) string {
	var structType string

	// Look for struct declarations that embed framework types
	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structTypeDecl, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Check if this struct embeds framework types
			for _, field := range structTypeDecl.Fields.List {
				if len(field.Names) == 0 { // Embedded field
					if selectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
						if ident, ok := selectorExpr.X.(*ast.Ident); ok && ident.Name == "framework" {
							structType = typeSpec.Name.Name
							return false // Found it, stop searching
						}
					}
				}
			}
		}

		return true
	})

	return structType
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

// validateAnnotationResult performs basic validation on an annotation result
func validateAnnotationResult(result *AnnotationResult) error {
	if result.Type == "" {
		return fmt.Errorf("annotation type is required")
	}

	if result.TerraformType == "" {
		return fmt.Errorf("terraform type is required")
	}

	// Validate terraform type format (should start with aws_)
	if !strings.HasPrefix(result.TerraformType, "aws_") {
		return fmt.Errorf("terraform type should start with 'aws_': %s", result.TerraformType)
	}

	return nil
}

// getAnnotationTypeName returns a human-readable name for the annotation type
func getAnnotationTypeName(annoType AnnotationType) string {
	switch annoType {
	case AnnotationSDKResource:
		return "SDK Resource"
	case AnnotationSDKDataSource:
		return "SDK Data Source"
	case AnnotationFrameworkResource:
		return "Framework Resource"
	case AnnotationFrameworkDataSource:
		return "Framework Data Source"
	case AnnotationEphemeralResource:
		return "Ephemeral Resource"
	default:
		return "Unknown"
	}
}

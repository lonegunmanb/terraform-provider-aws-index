package pkg

import (
	"go/parser"
	"go/token"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

func TestFileLevelAnnotationScanner(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedCount  int
		expectedType   AnnotationType
		expectedTF     string
		expectedName   string
		expectedStruct string
		expectedCRUD   map[string]string
	}{
		{
			name: "SDK Resource with CRUD methods",
			source: `package test

// @SDKResource("aws_lambda_function", name="Function")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
		UpdateWithoutTimeout: resourceFunctionUpdate,
		DeleteWithoutTimeout: resourceFunctionDelete,
	}
}`,
			expectedCount: 1,
			expectedType:  AnnotationSDKResource,
			expectedTF:    "aws_lambda_function",
			expectedName:  "Function",
			expectedCRUD: map[string]string{
				"create": "resourceFunctionCreate",
				"read":   "resourceFunctionRead",
				"update": "resourceFunctionUpdate",
				"delete": "resourceFunctionDelete",
			},
		},
		{
			name: "Framework Resource with struct type",
			source: `package test

// @FrameworkResource("aws_bedrock_guardrail", name="Guardrail")
func newGuardrailResource() (resource.ResourceWithConfigure, error) {
	return &guardrailResource{}, nil
}

type guardrailResource struct {
	framework.ResourceWithModel[guardrailResourceModel]
}`,
			expectedCount:  1,
			expectedType:   AnnotationFrameworkResource,
			expectedTF:     "aws_bedrock_guardrail",
			expectedName:   "Guardrail",
			expectedStruct: "guardrailResource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source code
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Create a mock FileInfo
			fileInfo := &gophon.FileInfo{
				File:     astFile,
				FilePath: "test.go",
			}

			// Scan for annotations
			results, err := scanFileForAnnotations(fileInfo)
			if err != nil {
				t.Fatalf("Failed to scan annotations: %v", err)
			}

			// Check expected count
			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d annotations, got %d", tt.expectedCount, len(results))
				return
			}

			// If we expect results, verify the first one
			if tt.expectedCount > 0 {
				result := results[0]

				if result.Type != tt.expectedType {
					t.Errorf("Expected type %v, got %v", tt.expectedType, result.Type)
				}

				if result.TerraformType != tt.expectedTF {
					t.Errorf("Expected terraform type %s, got %s", tt.expectedTF, result.TerraformType)
				}

				if result.Name != tt.expectedName {
					t.Errorf("Expected name %s, got %s", tt.expectedName, result.Name)
				}

				if tt.expectedStruct != "" && result.StructType != tt.expectedStruct {
					t.Errorf("Expected struct type %s, got %s", tt.expectedStruct, result.StructType)
				}

				if tt.expectedCRUD != nil {
					if len(result.CRUDMethods) != len(tt.expectedCRUD) {
						t.Errorf("Expected %d CRUD methods, got %d", len(tt.expectedCRUD), len(result.CRUDMethods))
					}
					for expectedKey, expectedValue := range tt.expectedCRUD {
						if actualValue, exists := result.CRUDMethods[expectedKey]; !exists {
							t.Errorf("Expected CRUD method %s not found", expectedKey)
						} else if actualValue != expectedValue {
							t.Errorf("Expected CRUD method %s to be %s, got %s", expectedKey, expectedValue, actualValue)
						}
					}
				}

				// Validate annotation
				if err := validateAnnotationResult(&result); err != nil {
					t.Errorf("Annotation validation failed: %v", err)
				}
			}
		})
	}
}

func TestRealWorldAWSFile(t *testing.T) {
	// Test with a real AWS provider file pattern
	source := `package lambda

import (
	"context"
	"time"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// @SDKResource("aws_lambda_function", name="Function")
// @Tags(identifierAttribute="arn")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
		UpdateWithoutTimeout: resourceFunctionUpdate,
		DeleteWithoutTimeout: resourceFunctionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation
	return nil
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation  
	return nil
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation
	return nil
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation
	return nil
}`

	// Parse the source code
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "function.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	// Create a mock FileInfo
	fileInfo := &gophon.FileInfo{
		File:     astFile,
		FilePath: "function.go",
	}

	// Scan for annotations
	results, err := scanFileForAnnotations(fileInfo)
	if err != nil {
		t.Fatalf("Failed to scan annotations: %v", err)
	}

	// Should find exactly one SDK resource
	if len(results) != 1 {
		t.Fatalf("Expected 1 annotation, got %d", len(results))
	}

	result := results[0]

	// Verify the annotation details
	if result.Type != AnnotationSDKResource {
		t.Errorf("Expected SDK Resource, got %v", result.Type)
	}

	if result.TerraformType != "aws_lambda_function" {
		t.Errorf("Expected aws_lambda_function, got %s", result.TerraformType)
	}

	if result.Name != "Function" {
		t.Errorf("Expected Function, got %s", result.Name)
	}

	// Verify CRUD methods were extracted
	expectedCRUD := map[string]string{
		"create": "resourceFunctionCreate",
		"read":   "resourceFunctionRead",
		"update": "resourceFunctionUpdate",
		"delete": "resourceFunctionDelete",
	}

	if len(result.CRUDMethods) != len(expectedCRUD) {
		t.Errorf("Expected %d CRUD methods, got %d", len(expectedCRUD), len(result.CRUDMethods))
	}

	for expectedKey, expectedValue := range expectedCRUD {
		if actualValue, exists := result.CRUDMethods[expectedKey]; !exists {
			t.Errorf("Expected CRUD method %s not found", expectedKey)
		} else if actualValue != expectedValue {
			t.Errorf("Expected CRUD method %s to be %s, got %s", expectedKey, expectedValue, actualValue)
		}
	}
}

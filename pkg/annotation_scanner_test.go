package pkg

import (
	"embed"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

//go:embed testharness/*.gocode
var testHarnessFS embed.FS

// TestRealWorldAnnotationScanning tests annotation scanning against real AWS provider code examples
func TestRealWorldAnnotationScanning(t *testing.T) {
	testCases := []struct {
		name                string
		filename            string
		expectedType        AnnotationType
		expectedTF          string
		expectedName        string
		expectedStruct      string
		expectedCRUDMethods map[string]string
	}{
		{
			name:         "SDK Resource - aws_lambda_invocation",
			filename:     "sdk_resource_aws_lambda_invocation.gocode",
			expectedType: AnnotationSDKResource,
			expectedTF:   "aws_lambda_invocation",
			expectedName: "Invocation",
			expectedCRUDMethods: map[string]string{
				"create": "resourceInvocationCreate",
				// Note: "read" is schema.NoopContext which we correctly skip
				"update": "resourceInvocationUpdate",
				"delete": "resourceInvocationDelete",
			},
		},
		{
			name:         "SDK DataSource - aws_ebs_snapshot",
			filename:     "sdk_data_aws_ebs_snapshot.gocode",
			expectedType: AnnotationSDKDataSource,
			expectedTF:   "aws_ebs_snapshot",
			expectedName: "EBS Snapshot",
			expectedCRUDMethods: map[string]string{
				"read": "dataSourceEBSSnapshotRead",
			},
		},
		{
			name:           "Framework Resource - aws_bedrock_guardrail",
			filename:       "framework_resource_aws_bedrock_guardrail.gocode",
			expectedType:   AnnotationFrameworkResource,
			expectedTF:     "aws_bedrock_guardrail",
			expectedName:   "Guardrail",
			expectedStruct: "guardrailResource",
		},
		{
			name:           "Framework DataSource - aws_bedrock_foundation_model",
			filename:       "framework_data_aws_bedrock_foundation_model.gocode",
			expectedType:   AnnotationFrameworkDataSource,
			expectedTF:     "aws_bedrock_foundation_model",
			expectedName:   "Foundation Model",
			expectedStruct: "foundationModelDataSource",
		},
		{
			name:           "Ephemeral Resource - aws_lambda_invocation",
			filename:       "framework_ephemeral_aws_lambda_invocation.gocode",
			expectedType:   AnnotationEphemeralResource,
			expectedTF:     "aws_lambda_invocation",
			expectedName:   "Invocation",
			expectedStruct: "invocationEphemeralResource",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the test file
			content, err := testHarnessFS.ReadFile("testharness/" + tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Parse the source code
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, tc.filename, string(content), parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Create a mock FileInfo
			fileInfo := &gophon.FileInfo{
				File:     astFile,
				FilePath: tc.filename,
			}

			// Scan for annotations
			results, err := scanFileForAnnotations(fileInfo)
			if err != nil {
				t.Fatalf("Failed to scan annotations: %v", err)
			}

			// Should find exactly one annotation
			if len(results) != 1 {
				t.Fatalf("Expected 1 annotation, got %d", len(results))
			}

			result := results[0]

			// Verify annotation type
			if result.Type != tc.expectedType {
				t.Errorf("Expected type %v, got %v", tc.expectedType, result.Type)
			}

			// Verify terraform type
			if result.TerraformType != tc.expectedTF {
				t.Errorf("Expected terraform type %s, got %s", tc.expectedTF, result.TerraformType)
			}

			// Verify name
			if result.Name != tc.expectedName {
				t.Errorf("Expected name %s, got %s", tc.expectedName, result.Name)
			}

			// Verify struct type for framework/ephemeral resources
			if tc.expectedStruct != "" {
				if result.StructType != tc.expectedStruct {
					t.Errorf("Expected struct type %s, got %s", tc.expectedStruct, result.StructType)
				}
			}

			// Verify CRUD methods for SDK resources/datasources
			if tc.expectedCRUDMethods != nil {
				if len(result.CRUDMethods) != len(tc.expectedCRUDMethods) {
					t.Errorf("Expected %d CRUD methods, got %d", len(tc.expectedCRUDMethods), len(result.CRUDMethods))
				}

				for expectedKey, expectedValue := range tc.expectedCRUDMethods {
					if actualValue, exists := result.CRUDMethods[expectedKey]; !exists {
						t.Errorf("Expected CRUD method %s not found", expectedKey)
					} else if actualValue != expectedValue {
						t.Errorf("Expected CRUD method %s = %s, got %s", expectedKey, expectedValue, actualValue)
					}
				}
			}

			// Verify framework methods are inferred correctly
			if tc.expectedType == AnnotationFrameworkResource {
				expectedFrameworkMethods := []string{"Create", "Read", "Update", "Delete", "Metadata", "Schema"}
				if len(result.FrameworkMethods) != len(expectedFrameworkMethods) {
					t.Errorf("Expected %d framework methods, got %d", len(expectedFrameworkMethods), len(result.FrameworkMethods))
				}
			} else if tc.expectedType == AnnotationFrameworkDataSource {
				expectedFrameworkMethods := []string{"Read", "Metadata", "Schema"}
				if len(result.FrameworkMethods) != len(expectedFrameworkMethods) {
					t.Errorf("Expected %d framework methods, got %d", len(expectedFrameworkMethods), len(result.FrameworkMethods))
				}
			} else if tc.expectedType == AnnotationEphemeralResource {
				expectedFrameworkMethods := []string{"Open", "Close", "Renew", "Metadata", "Schema"}
				if len(result.FrameworkMethods) != len(expectedFrameworkMethods) {
					t.Errorf("Expected %d framework methods, got %d", len(expectedFrameworkMethods), len(result.FrameworkMethods))
				}
			}
		})
	}
}

// TestFrameworkStructTypeExtraction specifically tests the problematic struct type extraction
func TestFrameworkStructTypeExtraction(t *testing.T) {
	testCases := []struct {
		name           string
		filename       string
		expectedStruct string
		annotatedFunc  string
	}{
		{
			name:           "Guardrail Resource",
			filename:       "framework_resource_aws_bedrock_guardrail.gocode",
			expectedStruct: "guardrailResource",
			annotatedFunc:  "newGuardrailResource",
		},
		{
			name:           "Foundation Model DataSource",
			filename:       "framework_data_aws_bedrock_foundation_model.gocode",
			expectedStruct: "foundationModelDataSource",
			annotatedFunc:  "newFoundationModelDataSource",
		},
		{
			name:           "Invocation Ephemeral Resource",
			filename:       "framework_ephemeral_aws_lambda_invocation.gocode",
			expectedStruct: "invocationEphemeralResource",
			annotatedFunc:  "newInvocationEphemeralResource",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the test file
			content, err := testHarnessFS.ReadFile("testharness/" + tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Parse the source code
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, tc.filename, string(content), parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Test the specific function we need to fix
			structType := extractFrameworkStructTypeBySchemaMethod(astFile)
			if structType != tc.expectedStruct {
				t.Errorf("Expected struct type %s, got %s", tc.expectedStruct, structType)
			}
		})
	}
}

// TestSDKCRUDExtraction tests CRUD method extraction from real AWS SDK resources
func TestSDKCRUDExtraction(t *testing.T) {
	testCases := []struct {
		name                string
		filename            string
		expectedCRUDMethods map[string]string
	}{
		{
			name:     "Lambda Invocation Resource",
			filename: "sdk_resource_aws_lambda_invocation.gocode",
			expectedCRUDMethods: map[string]string{
				"create": "resourceInvocationCreate",
				"update": "resourceInvocationUpdate",
				"delete": "resourceInvocationDelete",
				// Note: "read" might be "schema.NoopContext" which we should handle
			},
		},
		{
			name:     "EBS Snapshot DataSource",
			filename: "sdk_data_aws_ebs_snapshot.gocode",
			expectedCRUDMethods: map[string]string{
				"read": "dataSourceEBSSnapshotRead",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the test file
			content, err := testHarnessFS.ReadFile("testharness/" + tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Parse the source code
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, tc.filename, string(content), parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Test CRUD extraction
			var methods map[string]string
			if filepath.Base(tc.filename) == "sdk_resource_aws_lambda_invocation.gocode" {
				methods = extractSDKResourceCRUDFromFile(astFile)
			} else {
				methods = extractSDKDataSourceMethodsFromFile(astFile)
			}

			// Verify expected methods are found
			for expectedKey, expectedValue := range tc.expectedCRUDMethods {
				if actualValue, exists := methods[expectedKey]; !exists {
					t.Errorf("Expected CRUD method %s not found", expectedKey)
				} else if actualValue != expectedValue {
					// For now, just log special cases like schema.NoopContext
					if expectedValue == "schema.NoopContext" && actualValue != expectedValue {
						t.Logf("Special case: expected %s, got %s (this might be acceptable)", expectedValue, actualValue)
					} else if actualValue != expectedValue {
						t.Errorf("Expected CRUD method %s = %s, got %s", expectedKey, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

// TestAnnotationRegexAgainstRealWorld tests the annotation regex against real comment patterns
func TestAnnotationRegexAgainstRealWorld(t *testing.T) {
	testCases := []struct {
		name         string
		filename     string
		expectedMatches []struct {
			annotationType string
			terraformType  string
			name           string
		}
	}{
		{
			name:     "SDK Resource with single annotation",
			filename: "sdk_resource_aws_lambda_invocation.gocode",
			expectedMatches: []struct {
				annotationType string
				terraformType  string
				name           string
			}{
				{
					annotationType: "SDKResource",
					terraformType:  "aws_lambda_invocation",
					name:           "Invocation",
				},
			},
		},
		{
			name:     "SDK DataSource with multiple comment lines",
			filename: "sdk_data_aws_ebs_snapshot.gocode",
			expectedMatches: []struct {
				annotationType string
				terraformType  string
				name           string
			}{
				{
					annotationType: "SDKDataSource",
					terraformType:  "aws_ebs_snapshot",
					name:           "EBS Snapshot",
				},
			},
		},
		{
			name:     "Framework Resource with complex annotations",
			filename: "framework_resource_aws_bedrock_guardrail.gocode",
			expectedMatches: []struct {
				annotationType string
				terraformType  string
				name           string
			}{
				{
					annotationType: "FrameworkResource",
					terraformType:  "aws_bedrock_guardrail",
					name:           "Guardrail",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Read the test file
			content, err := testHarnessFS.ReadFile("testharness/" + tc.filename)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tc.filename, err)
			}

			// Parse the source code
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, tc.filename, string(content), parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse source: %v", err)
			}

			// Find annotations
			annotations := findAnnotationsInFile(astFile)

			// Verify expected matches
			if len(annotations) != len(tc.expectedMatches) {
				t.Errorf("Expected %d annotations, found %d", len(tc.expectedMatches), len(annotations))
			}

			for i, expected := range tc.expectedMatches {
				if i >= len(annotations) {
					t.Errorf("Missing annotation %d", i)
					continue
				}

				annotation := annotations[i]
				expectedType := stringToAnnotationType(expected.annotationType)

				if annotation.Type != expectedType {
					t.Errorf("Expected annotation type %v, got %v", expectedType, annotation.Type)
				}

				if annotation.TerraformType != expected.terraformType {
					t.Errorf("Expected terraform type %s, got %s", expected.terraformType, annotation.TerraformType)
				}

				if annotation.Name != expected.name {
					t.Errorf("Expected name %s, got %s", expected.name, annotation.Name)
				}
			}
		})
	}
}

// Helper function to convert string to AnnotationType
func stringToAnnotationType(s string) AnnotationType {
	switch s {
	case "SDKResource":
		return AnnotationSDKResource
	case "SDKDataSource":
		return AnnotationSDKDataSource
	case "FrameworkResource":
		return AnnotationFrameworkResource
	case "FrameworkDataSource":
		return AnnotationFrameworkDataSource
	case "EphemeralResource":
		return AnnotationEphemeralResource
	default:
		return AnnotationType("") // Invalid
	}
}

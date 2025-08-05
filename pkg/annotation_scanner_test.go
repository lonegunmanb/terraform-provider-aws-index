package pkg

import (
	"go/parser"
	"go/token"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

func TestAnnotationScanner(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedCount  int
		expectedType   AnnotationType
		expectedTF     string
		expectedName   string
		expectedStruct string
	}{
		{
			name: "SDK Resource annotation",
			source: `package test

// @SDKResource("aws_lambda_function", name="Function")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
	}
}`,
			expectedCount: 1,
			expectedType:  AnnotationSDKResource,
			expectedTF:    "aws_lambda_function",
			expectedName:  "Function",
		},
		{
			name: "SDK DataSource annotation",
			source: `package test

// @SDKDataSource("aws_lambda_alias", name="Alias")
func dataSourceAlias() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAliasRead,
	}
}`,
			expectedCount: 1,
			expectedType:  AnnotationSDKDataSource,
			expectedTF:    "aws_lambda_alias",
			expectedName:  "Alias",
		},
		{
			name: "Framework Resource annotation with struct type",
			source: `package test

// @FrameworkResource("aws_bedrock_guardrail", name="Guardrail")
func newGuardrailResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &guardrailResource{
		flexOpt: fwflex.WithFieldNameSuffix("Config"),
	}
	return r, nil
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
		{
			name: "Framework DataSource annotation",
			source: `package test

// @FrameworkDataSource("aws_bedrock_inference_profile", name="Inference Profile")
func newInferenceProfileDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfileDataSource{}, nil
}

type inferenceProfileDataSource struct {
	framework.DataSourceWithModel[inferenceProfileDataSourceModel]
}`,
			expectedCount:  1,
			expectedType:   AnnotationFrameworkDataSource,
			expectedTF:     "aws_bedrock_inference_profile",
			expectedName:   "Inference Profile",
			expectedStruct: "inferenceProfileDataSource",
		},
		{
			name: "Ephemeral Resource annotation",
			source: `package test

// @EphemeralResource("aws_lambda_invocation", name="Invocation")
func newInvocationEphemeralResource(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &invocationEphemeralResource{}, nil
}

type invocationEphemeralResource struct {
	framework.EphemeralResourceWithModel[invocationEphemeralResourceModel]
}`,
			expectedCount:  1,
			expectedType:   AnnotationEphemeralResource,
			expectedTF:     "aws_lambda_invocation",
			expectedName:   "Invocation",
			expectedStruct: "invocationEphemeralResource",
		},
		{
			name: "No annotation",
			source: `package test

func regularFunction() *schema.Resource {
	return &schema.Resource{}
}`,
			expectedCount: 0,
		},
		{
			name: "Multiple annotations in same file",
			source: `package test

// @SDKResource("aws_s3_bucket", name="Bucket")
func resourceBucket() *schema.Resource {
	return &schema.Resource{}
}

// @SDKDataSource("aws_s3_bucket", name="Bucket")
func dataSourceBucket() *schema.Resource {
	return &schema.Resource{}
}`,
			expectedCount: 2,
			expectedType:  AnnotationSDKResource,
			expectedTF:    "aws_s3_bucket",
			expectedName:  "Bucket",
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

				// Validate annotation
				if err := validateAnnotationResult(&result); err != nil {
					t.Errorf("Annotation validation failed: %v", err)
				}
			}
		})
	}
}

func TestScanPackageForAnnotations(t *testing.T) {
	// Create a mock package with multiple files
	fset := token.NewFileSet()

	// File 1: SDK resources and data sources
	source1 := `package test

// @SDKResource("aws_s3_bucket", name="Bucket")
func resourceBucket() *schema.Resource {
	return &schema.Resource{}
}

// @SDKDataSource("aws_s3_bucket", name="Bucket")
func dataSourceBucket() *schema.Resource {
	return &schema.Resource{}
}`

	astFile1, err := parser.ParseFile(fset, "file1.go", source1, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source1: %v", err)
	}

	// File 2: Framework resources
	source2 := `package test

// @FrameworkResource("aws_bedrock_model", name="Model")
func newModelResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &modelResource{}, nil
}

type modelResource struct {
	framework.ResourceWithModel[modelResourceModel]
}`

	astFile2, err := parser.ParseFile(fset, "file2.go", source2, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source2: %v", err)
	}

	// Create mock PackageInfo
	packageInfo := &gophon.PackageInfo{
		Files: []*gophon.FileInfo{
			{
				File:     astFile1,
				FilePath: "file1.go",
			},
			{
				File:     astFile2,
				FilePath: "file2.go",
			},
		},
	}

	// Scan the package
	results, err := scanPackageForAnnotations(packageInfo)
	if err != nil {
		t.Fatalf("Failed to scan package: %v", err)
	}

	// Verify results
	if results.TotalAnnotations != 3 {
		t.Errorf("Expected 3 total annotations, got %d", results.TotalAnnotations)
	}

	if len(results.SDKResources) != 1 {
		t.Errorf("Expected 1 SDK resource, got %d", len(results.SDKResources))
	}

	if len(results.SDKDataSources) != 1 {
		t.Errorf("Expected 1 SDK data source, got %d", len(results.SDKDataSources))
	}

	if len(results.FrameworkResources) != 1 {
		t.Errorf("Expected 1 framework resource, got %d", len(results.FrameworkResources))
	}

	// Test GetAll method
	allResults := results.GetAll()
	if len(allResults) != 3 {
		t.Errorf("Expected 3 results from GetAll(), got %d", len(allResults))
	}
}

func TestAnnotationRegex(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		shouldMatch bool
		expectedType string
		expectedTF   string
		expectedName string
	}{
		{
			name:         "Simple SDK Resource",
			comment:      `@SDKResource("aws_lambda_function", name="Function")`,
			shouldMatch:  true,
			expectedType: "SDKResource",
			expectedTF:   "aws_lambda_function",
			expectedName: "Function",
		},
		{
			name:         "Framework Resource with complex name",
			comment:      `@FrameworkResource("aws_bedrock_guardrail", name="Guardrail")`,
			shouldMatch:  true,
			expectedType: "FrameworkResource",
			expectedTF:   "aws_bedrock_guardrail",
			expectedName: "Guardrail",
		},
		{
			name:         "Annotation without name parameter",
			comment:      `@EphemeralResource("aws_lambda_invocation")`,
			shouldMatch:  true,
			expectedType: "EphemeralResource",
			expectedTF:   "aws_lambda_invocation",
			expectedName: "",
		},
		{
			name:        "Not an annotation",
			comment:     `// This is just a regular comment`,
			shouldMatch: false,
		},
		{
			name:        "Invalid annotation format",
			comment:     `@InvalidAnnotation("aws_test")`,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := annotationRegex.FindStringSubmatch(tt.comment)

			if tt.shouldMatch {
				if len(matches) < 3 {
					t.Errorf("Expected annotation to match, but it didn't")
					return
				}

				if matches[1] != tt.expectedType {
					t.Errorf("Expected type %s, got %s", tt.expectedType, matches[1])
				}

				if matches[2] != tt.expectedTF {
					t.Errorf("Expected terraform type %s, got %s", tt.expectedTF, matches[2])
				}

				expectedName := tt.expectedName
				actualName := ""
				if len(matches) > 3 {
					actualName = matches[3]
				}

				if actualName != expectedName {
					t.Errorf("Expected name %s, got %s", expectedName, actualName)
				}
			} else {
				if len(matches) >= 3 {
					t.Errorf("Expected annotation not to match, but it did: %v", matches)
				}
			}
		})
	}
}

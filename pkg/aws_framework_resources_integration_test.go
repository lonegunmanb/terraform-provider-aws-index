package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAWSFrameworkResourcesIntegration_WriteResourceFiles tests that AWS Framework resources
// are correctly integrated into the WriteResourceFiles pipeline
func TestAWSFrameworkResourcesIntegration_WriteResourceFiles(t *testing.T) {
	tests := []struct {
		name                    string
		frameworkResources      map[string]AWSResourceInfo
		expectedFiles           []string
		expectedTerraformTypes  []string
	}{
		{
			name: "single_framework_resource",
			frameworkResources: map[string]AWSResourceInfo{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					Name:           "Bucket",
					FactoryFunction: "newBucketResource",
					SDKType:        "framework",
					StructType:     "bucketResource",
				},
			},
			expectedFiles:          []string{"aws_s3_bucket.json"},
			expectedTerraformTypes: []string{"aws_s3_bucket"},
		},
		{
			name: "multiple_framework_resources",
			frameworkResources: map[string]AWSResourceInfo{
				"aws_dynamodb_table": {
					TerraformType:   "aws_dynamodb_table",
					Name:           "Table",
					FactoryFunction: "newTableResource",
					SDKType:        "framework",
					StructType:     "tableResource",
				},
				"aws_lambda_function": {
					TerraformType:   "aws_lambda_function",
					Name:           "Function",
					FactoryFunction: "newFunctionResource",
					SDKType:        "framework",
					StructType:     "functionResource",
				},
			},
			expectedFiles:          []string{"aws_dynamodb_table.json", "aws_lambda_function.json"},
			expectedTerraformTypes: []string{"aws_dynamodb_table", "aws_lambda_function"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a temporary directory
			outputDir, err := os.MkdirTemp("", "framework_resources_test")
			require.NoError(t, err)
			defer os.RemoveAll(outputDir)

			// Setup: Create test service with Framework resources
			service := ServiceRegistration{
				ServiceName:           "s3",
				PackagePath:          "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
				AWSSDKResources:      make(map[string]AWSResourceInfo),
				AWSSDKDataSources:    make(map[string]AWSResourceInfo),
				AWSFrameworkResources: tt.frameworkResources,
				AWSFrameworkDataSources: make(map[string]AWSResourceInfo),
				AWSEphemeralResources: make(map[string]AWSResourceInfo),
				SupportedResources:   make(map[string]string),
				SupportedDataSources: make(map[string]string),
				Resources:            []string{},
				DataSources:          []string{},
				EphemeralFunctions:   []string{},
				ResourceCRUDMethods:  make(map[string]*LegacyResourceCRUDFunctions),
				DataSourceMethods:    make(map[string]*LegacyDataSourceMethods),
				ResourceTerraformTypes:   make(map[string]string),
				DataSourceTerraformTypes: make(map[string]string),
				EphemeralTerraformTypes:  make(map[string]string),
			}

			// Setup: Create TerraformProviderIndex
			index := &TerraformProviderIndex{
				Version:  "5.0.0",
				Services: []ServiceRegistration{service},
			}

			// Execute: WriteResourceFiles
			err = index.WriteResourceFiles(outputDir, nil)

			// Verify: No errors occurred
			require.NoError(t, err)

			// Verify: Expected resource files were created
			resourcesDir := filepath.Join(outputDir, "resources")
			for _, expectedFile := range tt.expectedFiles {
				filePath := filepath.Join(resourcesDir, expectedFile)
				assert.FileExists(t, filePath, "Expected resource file %s should exist", expectedFile)

				// Verify: File contains valid JSON
				fileBytes, err := os.ReadFile(filePath)
				require.NoError(t, err)

				var resourceData map[string]interface{}
				err = json.Unmarshal(fileBytes, &resourceData)
				require.NoError(t, err, "Resource file %s should contain valid JSON", expectedFile)

				// Verify: Contains required TerraformResource fields
				assert.Contains(t, resourceData, "terraform_type")
				assert.Contains(t, resourceData, "namespace")
				assert.Contains(t, resourceData, "registration_method")
				assert.Contains(t, resourceData, "sdk_type")

				// Verify: SDK type is framework
				assert.Equal(t, "aws_framework", resourceData["sdk_type"])

				// Verify: Registration method is correct
				assert.Equal(t, "FrameworkResources", resourceData["registration_method"])
			}

			// Verify: File count matches expectation
			files, err := os.ReadDir(resourcesDir)
			require.NoError(t, err)
			assert.Len(t, files, len(tt.expectedFiles), "Should have exactly %d resource files", len(tt.expectedFiles))
		})
	}
}

// TestNewTerraformResourceFromAWSFramework_BasicMapping tests the conversion function
// for mapping AWS Framework resource info to TerraformResource API
func TestNewTerraformResourceFromAWSFramework_BasicMapping(t *testing.T) {
	// Setup: Create test AWS Framework resource with proper StructType
	awsResource := AWSResourceInfo{
		TerraformType:   "aws_s3_bucket",
		Name:           "Bucket",
		FactoryFunction: "newBucketResource",
		SDKType:        "framework",
		StructType:     "bucketResource", // Framework resources have struct types
	}

	// Setup: Create test service
	service := ServiceRegistration{
		ServiceName: "s3",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
	}

	// Execute: Convert AWS Framework resource to TerraformResource
	result := NewTerraformResourceFromAWSFramework(awsResource, service)

	// Verify: Basic fields are correctly mapped
	assert.Equal(t, "aws_s3_bucket", result.TerraformType)
	assert.Equal(t, "github.com/hashicorp/terraform-provider-aws/internal/service/s3", result.Namespace)
	assert.Equal(t, "FrameworkResources", result.RegistrationMethod)
	assert.Equal(t, "aws_framework", result.SDKType)

	// Verify: Schema and Attribute indexes use method-based pattern
	assert.Equal(t, "method.bucketResource.Schema.goindex", result.SchemaIndex)
	assert.Equal(t, "method.bucketResource.Schema.goindex", result.AttributeIndex)

	// Verify: Struct type is properly populated for Framework resources
	assert.Equal(t, "bucketResource", result.StructType)
}

// TestNewTerraformResourceFromAWSFramework_WithCRUDMethods tests Framework resource with CRUD methods
func TestNewTerraformResourceFromAWSFramework_WithCRUDMethods(t *testing.T) {
	// Setup: Create test AWS Framework resource
	awsResource := AWSResourceInfo{
		TerraformType:   "aws_bedrock_guardrail",
		Name:           "Guardrail",
		FactoryFunction: "newGuardrailResource",
		SDKType:        "framework",
		StructType:     "guardrailResource",
	}

	// Setup: Create test service with CRUD methods
	service := ServiceRegistration{
		ServiceName: "bedrock",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock",
		ResourceCRUDMethods: map[string]*LegacyResourceCRUDFunctions{
			"aws_bedrock_guardrail": {
				CreateMethod: "Create",  // Framework methods are just the interface method names
				ReadMethod:   "Read",
				UpdateMethod: "Update",
				DeleteMethod: "Delete",
			},
		},
	}

	// Execute: Convert AWS Framework resource to TerraformResource
	result := NewTerraformResourceFromAWSFramework(awsResource, service)

	// Verify: Basic fields are correctly mapped
	assert.Equal(t, "aws_bedrock_guardrail", result.TerraformType)
	assert.Equal(t, "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock", result.Namespace)
	assert.Equal(t, "FrameworkResources", result.RegistrationMethod)
	assert.Equal(t, "aws_framework", result.SDKType)
	assert.Equal(t, "guardrailResource", result.StructType)

	// Verify: Schema and Attribute indexes use method-based pattern
	assert.Equal(t, "method.guardrailResource.Schema.goindex", result.SchemaIndex)
	assert.Equal(t, "method.guardrailResource.Schema.goindex", result.AttributeIndex)

	// Verify: CRUD methods use method-based pattern with struct type
	assert.Equal(t, "method.guardrailResource.Create.goindex", result.CreateIndex)
	assert.Equal(t, "method.guardrailResource.Read.goindex", result.ReadIndex)
	assert.Equal(t, "method.guardrailResource.Update.goindex", result.UpdateIndex)
	assert.Equal(t, "method.guardrailResource.Delete.goindex", result.DeleteIndex)
}

// TestAWSFrameworkResourcesIntegration_MixedWithSDK tests that Framework and SDK resources
// can coexist in the same service package
func TestAWSFrameworkResourcesIntegration_MixedWithSDK(t *testing.T) {
	// Setup: Create temporary directory
	outputDir, err := os.MkdirTemp("", "mixed_resources_test")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Setup: Create service with both SDK and Framework resources
	service := ServiceRegistration{
		ServiceName: "s3",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		AWSSDKResources: map[string]AWSResourceInfo{
			"aws_s3_bucket_policy": {
				TerraformType:   "aws_s3_bucket_policy",
				Name:           "BucketPolicy",
				FactoryFunction: "resourceBucketPolicy",
				SDKType:        "sdk",
			},
		},
		AWSFrameworkResources: map[string]AWSResourceInfo{
			"aws_s3_bucket": {
				TerraformType:   "aws_s3_bucket",
				Name:           "Bucket",
				FactoryFunction: "newBucketResource",
				SDKType:        "framework",
				StructType:     "bucketResource", // Framework resources need StructType
			},
		},
		AWSSDKDataSources:       make(map[string]AWSResourceInfo),
		AWSFrameworkDataSources: make(map[string]AWSResourceInfo),
		AWSEphemeralResources:   make(map[string]AWSResourceInfo),
		SupportedResources:      make(map[string]string),
		SupportedDataSources:    make(map[string]string),
		Resources:               []string{},
		DataSources:             []string{},
		EphemeralFunctions:      []string{},
		ResourceCRUDMethods:     make(map[string]*LegacyResourceCRUDFunctions),
		DataSourceMethods:       make(map[string]*LegacyDataSourceMethods),
		ResourceTerraformTypes:  make(map[string]string),
		DataSourceTerraformTypes: make(map[string]string),
		EphemeralTerraformTypes: make(map[string]string),
	}

	// Setup: Create TerraformProviderIndex
	index := &TerraformProviderIndex{
		Version:  "5.0.0",
		Services: []ServiceRegistration{service},
	}

	// Execute: WriteResourceFiles
	err = index.WriteResourceFiles(outputDir, nil)
	require.NoError(t, err)

	// Verify: Both SDK and Framework resource files were created
	resourcesDir := filepath.Join(outputDir, "resources")
	
	// Check SDK resource
	sdkFilePath := filepath.Join(resourcesDir, "aws_s3_bucket_policy.json")
	assert.FileExists(t, sdkFilePath)
	
	// Check Framework resource
	frameworkFilePath := filepath.Join(resourcesDir, "aws_s3_bucket.json")
	assert.FileExists(t, frameworkFilePath)

	// Verify: Framework resource has correct SDK type
	frameworkBytes, err := os.ReadFile(frameworkFilePath)
	require.NoError(t, err)
	
	var frameworkData map[string]interface{}
	err = json.Unmarshal(frameworkBytes, &frameworkData)
	require.NoError(t, err)
	
	assert.Equal(t, "aws_framework", frameworkData["sdk_type"])
	assert.Equal(t, "FrameworkResources", frameworkData["registration_method"])

	// Verify: SDK resource still works correctly
	sdkBytes, err := os.ReadFile(sdkFilePath)
	require.NoError(t, err)
	
	var sdkData map[string]interface{}
	err = json.Unmarshal(sdkBytes, &sdkData)
	require.NoError(t, err)
	
	assert.Equal(t, "aws_sdk", sdkData["sdk_type"])
	assert.Equal(t, "SDKResources", sdkData["registration_method"])
}

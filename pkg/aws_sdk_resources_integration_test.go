package pkg

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAWSSDKResourcesIntegration_WriteResourceFiles tests that AWS SDK resources
// are properly integrated into the resource file generation pipeline
func TestAWSSDKResourcesIntegration_WriteResourceFiles(t *testing.T) {
	tests := []struct {
		name                string
		awsSDKResources     map[string]AWSResourceInfo
		expectedFiles       []string
		expectedContentKeys []string
	}{
		{
			name: "Single AWS SDK resource creates individual JSON file",
			awsSDKResources: map[string]AWSResourceInfo{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					FactoryFunction: "resourceBucket",
					Name:            "Bucket",
					SDKType:         "sdk",
				},
			},
			expectedFiles:       []string{"aws_s3_bucket.json"},
			expectedContentKeys: []string{"terraform_type", "factory_function", "name", "sdk_type"},
		},
		{
			name: "Multiple AWS SDK resources create multiple JSON files",
			awsSDKResources: map[string]AWSResourceInfo{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					FactoryFunction: "resourceBucket",
					Name:            "Bucket",
					SDKType:         "sdk",
				},
				"aws_ec2_instance": {
					TerraformType:   "aws_ec2_instance",
					FactoryFunction: "resourceInstance",
					Name:            "Instance",
					SDKType:         "sdk",
				},
			},
			expectedFiles:       []string{"aws_s3_bucket.json", "aws_ec2_instance.json"},
			expectedContentKeys: []string{"terraform_type", "factory_function", "name", "sdk_type"},
		},
		{
			name:                "Empty AWS SDK resources creates no files",
			awsSDKResources:     map[string]AWSResourceInfo{},
			expectedFiles:       []string{},
			expectedContentKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup filesystem
			fs := afero.NewMemMapFs()
			stub := gostub.Stub(&outputFs, fs)
			defer stub.Reset()

			outputDir := "/tmp/output"
			resourcesDir := filepath.Join(outputDir, "resources")

			// Create TerraformProviderIndex with AWS SDK resources
			index := &TerraformProviderIndex{
				Version: "v5.0.0",
				Services: []ServiceRegistration{
					{
						ServiceName:     "s3",
						PackagePath:     "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
						AWSSDKResources: tt.awsSDKResources,
					},
				},
				Statistics: ProviderStatistics{},
			}

			// Execute WriteResourceFiles
			err := index.WriteResourceFiles(outputDir, nil)
			require.NoError(t, err)

			// Verify expected files were created
			if len(tt.expectedFiles) == 0 {
				// If no resources expected, verify resources directory is either empty or doesn't exist
				exists, err := afero.DirExists(fs, resourcesDir)
				if err == nil && exists {
					files, err := afero.ReadDir(fs, resourcesDir)
					require.NoError(t, err)
					assert.Empty(t, files, "Resources directory should be empty when no AWS SDK resources")
				}
			} else {
				// Verify each expected file exists
				for _, expectedFile := range tt.expectedFiles {
					filePath := filepath.Join(resourcesDir, expectedFile)
					exists, err := afero.Exists(fs, filePath)
					require.NoError(t, err)
					assert.True(t, exists, "Expected resource file %s should exist", expectedFile)

					// Read and verify file content
					fileContent, err := afero.ReadFile(fs, filePath)
					require.NoError(t, err)

					var resourceData map[string]interface{}
					err = json.Unmarshal(fileContent, &resourceData)
					require.NoError(t, err)

					// Verify expected keys are present
					for _, key := range tt.expectedContentKeys {
						assert.Contains(t, resourceData, key, "Resource file should contain key: %s", key)
					}

					// Verify specific values for AWS SDK resources
					assert.Equal(t, "aws_sdk", resourceData["sdk_type"], "SDK type should be aws_sdk")
					assert.Contains(t, resourceData["terraform_type"], "aws_", "Terraform type should start with aws_")
				}
			}
		})
	}
}

// TestAWSSDKResourcesIntegration_TerraformResourceConversion tests that AWS SDK resources
// are properly converted to TerraformResource structs with correct metadata
func TestAWSSDKResourcesIntegration_TerraformResourceConversion(t *testing.T) {
	// Setup test data
	awsResource := AWSResourceInfo{
		TerraformType:   "aws_s3_bucket",
		FactoryFunction: "resourceBucket",
		Name:            "Bucket",
		SDKType:         "sdk",
	}

	service := ServiceRegistration{
		ServiceName:     "s3",
		PackagePath:     "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		AWSSDKResources: map[string]AWSResourceInfo{"aws_s3_bucket": awsResource},
	}

	// Execute conversion function (this will be implemented)
	terraformResource := NewTerraformResourceFromAWSSDK(awsResource, service)

	// Verify conversion results
	assert.Equal(t, "aws_s3_bucket", terraformResource.TerraformType)
	assert.Equal(t, "aws_sdk", terraformResource.SDKType)
	assert.Equal(t, "SDKResources", terraformResource.RegistrationMethod)
	assert.Equal(t, service.PackagePath, terraformResource.Namespace)
	assert.Equal(t, "", terraformResource.StructType) // SDK resources don't have struct types

	// Verify that AWS-specific metadata is preserved in indexes
	assert.Contains(t, terraformResource.SchemaIndex, "resourceBucket")
	assert.NotEmpty(t, terraformResource.CreateIndex)
	assert.NotEmpty(t, terraformResource.ReadIndex)
}

// TestNewTerraformResourceFromAWSSDK_CRUDMethodExtraction tests that the function
// properly uses extracted CRUD methods when available in serviceReg.ResourceCRUDMethods
func TestNewTerraformResourceFromAWSSDK_CRUDMethodExtraction(t *testing.T) {
	tests := []struct {
		name            string
		awsResource     AWSResourceInfo
		crudMethods     *LegacyResourceCRUDFunctions
		expectedIndexes map[string]string
		description     string
	}{
		{
			name: "Uses extracted CRUD methods when available",
			awsResource: AWSResourceInfo{
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "resourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
			},
			crudMethods: &LegacyResourceCRUDFunctions{
				CreateMethod: "resourceBucketCreate",
				ReadMethod:   "resourceBucketRead",
				UpdateMethod: "resourceBucketUpdate",
				DeleteMethod: "resourceBucketDelete",
			},
			expectedIndexes: map[string]string{
				"SchemaIndex":    "func.resourceBucket.goindex",
				"CreateIndex":    "func.resourceBucketCreate.goindex",
				"ReadIndex":      "func.resourceBucketRead.goindex",
				"UpdateIndex":    "func.resourceBucketUpdate.goindex",
				"DeleteIndex":    "func.resourceBucketDelete.goindex",
				"AttributeIndex": "func.resourceBucket.goindex",
			},
			description: "When CRUD methods are extracted, should use specific method names for each operation",
		},
		{
			name: "Falls back to factory function when CRUD methods not available",
			awsResource: AWSResourceInfo{
				TerraformType:   "aws_ec2_instance",
				FactoryFunction: "resourceInstance",
				Name:            "Instance",
				SDKType:         "sdk",
			},
			crudMethods: nil, // No extracted CRUD methods
			expectedIndexes: map[string]string{
				"SchemaIndex":    "func.resourceInstance.goindex",
				"CreateIndex":    "func.resourceInstance.goindex", // Should fallback to factory function, not ".create"
				"ReadIndex":      "func.resourceInstance.goindex",
				"UpdateIndex":    "func.resourceInstance.goindex",
				"DeleteIndex":    "func.resourceInstance.goindex",
				"AttributeIndex": "func.resourceInstance.goindex",
			},
			description: "When no CRUD methods available, should use factory function name for all operations",
		},
		{
			name: "Handles partial CRUD methods",
			awsResource: AWSResourceInfo{
				TerraformType:   "aws_dynamodb_table",
				FactoryFunction: "resourceTable",
				Name:            "Table",
				SDKType:         "sdk",
			},
			crudMethods: &LegacyResourceCRUDFunctions{
				CreateMethod: "resourceTableCreate",
				ReadMethod:   "resourceTableRead",
				// UpdateMethod and DeleteMethod are empty
			},
			expectedIndexes: map[string]string{
				"SchemaIndex":    "func.resourceTable.goindex",
				"CreateIndex":    "func.resourceTableCreate.goindex",
				"ReadIndex":      "func.resourceTableRead.goindex",
				"UpdateIndex":    "func.resourceTable.goindex", // Should fallback to factory function
				"DeleteIndex":    "func.resourceTable.goindex", // Should fallback to factory function
				"AttributeIndex": "func.resourceTable.goindex",
			},
			description: "When partial CRUD methods available, should use specific methods when available and fallback for missing ones",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup service registration with or without CRUD methods
			service := ServiceRegistration{
				ServiceName:         "s3",
				PackagePath:         "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
				AWSSDKResources:     map[string]AWSResourceInfo{tt.awsResource.TerraformType: tt.awsResource},
				ResourceCRUDMethods: make(map[string]*LegacyResourceCRUDFunctions),
			}

			// Add CRUD methods if provided
			if tt.crudMethods != nil {
				service.ResourceCRUDMethods[tt.awsResource.TerraformType] = tt.crudMethods
			}

			// Execute the function
			result := NewTerraformResourceFromAWSSDK(tt.awsResource, service)

			// Verify basic fields
			assert.Equal(t, tt.awsResource.TerraformType, result.TerraformType, "TerraformType should match")
			assert.Equal(t, "aws_sdk", result.SDKType, "SDKType should be aws_sdk")
			assert.Equal(t, "SDKResources", result.RegistrationMethod, "RegistrationMethod should be SDKResources")
			assert.Equal(t, "", result.StructType, "StructType should be empty for SDK resources")

			// Verify exact index values - this is the critical test that will catch the bug
			assert.Equal(t, tt.expectedIndexes["SchemaIndex"], result.SchemaIndex, 
				"SchemaIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["CreateIndex"], result.CreateIndex, 
				"CreateIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["ReadIndex"], result.ReadIndex, 
				"ReadIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["UpdateIndex"], result.UpdateIndex, 
				"UpdateIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["DeleteIndex"], result.DeleteIndex, 
				"DeleteIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["AttributeIndex"], result.AttributeIndex, 
				"AttributeIndex: %s", tt.description)
		})
	}
}

// TestAWSSDKResourcesIntegration_BackwardCompatibility tests that existing
// TerraformResource API remains unchanged while supporting AWS SDK resources
func TestAWSSDKResourcesIntegration_BackwardCompatibility(t *testing.T) {
	// Setup legacy resource for comparison
	legacyService := ServiceRegistration{
		ServiceName:        "keyvault",
		PackagePath:        "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
		SupportedResources: map[string]string{"azurerm_key_vault": "resourceKeyVault"},
	}

	// Setup AWS SDK resource
	awsService := ServiceRegistration{
		ServiceName: "s3",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		AWSSDKResources: map[string]AWSResourceInfo{
			"aws_s3_bucket": {
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "resourceBucket",
				SDKType:         "sdk",
			},
		},
	}

	// Create TerraformResource from legacy (existing functionality)
	legacyResource := NewTerraformResourceInfo("azurerm_key_vault", "", "resourceKeyVault", "legacy_pluginsdk", legacyService)

	// Create TerraformResource from AWS SDK (new functionality)
	awsSDKResource := NewTerraformResourceFromAWSSDK(awsService.AWSSDKResources["aws_s3_bucket"], awsService)

	// Verify both have same TerraformResource structure (backward compatibility)
	assert.IsType(t, TerraformResource{}, legacyResource)
	assert.IsType(t, TerraformResource{}, awsSDKResource)

	// Verify all required fields are present in both
	requiredFields := []string{"terraform_type", "namespace", "registration_method", "sdk_type"}

	legacyJSON, err := json.Marshal(legacyResource)
	require.NoError(t, err)
	var legacyData map[string]interface{}
	err = json.Unmarshal(legacyJSON, &legacyData)
	require.NoError(t, err)

	awsJSON, err := json.Marshal(awsSDKResource)
	require.NoError(t, err)
	var awsData map[string]interface{}
	err = json.Unmarshal(awsJSON, &awsData)
	require.NoError(t, err)

	for _, field := range requiredFields {
		assert.Contains(t, legacyData, field, "Legacy resource should have field: %s", field)
		assert.Contains(t, awsData, field, "AWS SDK resource should have field: %s", field)
	}
}

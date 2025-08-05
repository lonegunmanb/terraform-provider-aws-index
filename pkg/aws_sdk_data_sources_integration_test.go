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

// TestAWSSDKDataSourcesIntegration_WriteDataSourceFiles tests that AWS SDK data sources
// are properly integrated into the data source file generation pipeline
func TestAWSSDKDataSourcesIntegration_WriteDataSourceFiles(t *testing.T) {
	tests := []struct {
		name                string
		awsSDKDataSources   map[string]AWSResourceInfo
		expectedFiles       []string
		expectedContentKeys []string
	}{
		{
			name: "Single AWS SDK data source creates individual JSON file",
			awsSDKDataSources: map[string]AWSResourceInfo{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					FactoryFunction: "dataSourceBucket",
					Name:            "Bucket",
					SDKType:         "sdk",
				},
			},
			expectedFiles:       []string{"aws_s3_bucket.json"},
			expectedContentKeys: []string{"terraform_type", "factory_function", "name", "sdk_type"},
		},
		{
			name: "Multiple AWS SDK data sources create multiple JSON files",
			awsSDKDataSources: map[string]AWSResourceInfo{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					FactoryFunction: "dataSourceBucket",
					Name:            "Bucket",
					SDKType:         "sdk",
				},
				"aws_ec2_instance": {
					TerraformType:   "aws_ec2_instance",
					FactoryFunction: "dataSourceInstance",
					Name:            "Instance",
					SDKType:         "sdk",
				},
			},
			expectedFiles:       []string{"aws_s3_bucket.json", "aws_ec2_instance.json"},
			expectedContentKeys: []string{"terraform_type", "factory_function", "name", "sdk_type"},
		},
		{
			name:                "Empty AWS SDK data sources creates no files",
			awsSDKDataSources:   map[string]AWSResourceInfo{},
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
			dataSourcesDir := filepath.Join(outputDir, "datasources")

			// Create TerraformProviderIndex with AWS SDK data sources
			index := &TerraformProviderIndex{
				Version: "v5.0.0",
				Services: []ServiceRegistration{
					{
						ServiceName:       "s3",
						PackagePath:       "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
						AWSSDKDataSources: tt.awsSDKDataSources,
					},
				},
				Statistics: ProviderStatistics{},
			}

			// Execute WriteDataSourceFiles
			err := index.WriteDataSourceFiles(outputDir, nil)
			require.NoError(t, err)

			// Verify expected files were created
			if len(tt.expectedFiles) == 0 {
				// If no data sources expected, verify datasources directory is either empty or doesn't exist
				exists, err := afero.DirExists(fs, dataSourcesDir)
				if err == nil && exists {
					files, err := afero.ReadDir(fs, dataSourcesDir)
					require.NoError(t, err)
					assert.Empty(t, files, "DataSources directory should be empty when no AWS SDK data sources")
				}
			} else {
				// Verify each expected file exists
				for _, expectedFile := range tt.expectedFiles {
					filePath := filepath.Join(dataSourcesDir, expectedFile)
					exists, err := afero.Exists(fs, filePath)
					require.NoError(t, err)
					assert.True(t, exists, "Expected data source file %s should exist", expectedFile)

					// Read and verify file content
					fileContent, err := afero.ReadFile(fs, filePath)
					require.NoError(t, err)

					var dataSourceData map[string]interface{}
					err = json.Unmarshal(fileContent, &dataSourceData)
					require.NoError(t, err)

					// Verify expected keys are present
					for _, key := range tt.expectedContentKeys {
						assert.Contains(t, dataSourceData, key, "Data source file should contain key: %s", key)
					}

					// Verify specific values for AWS SDK data sources
					assert.Equal(t, "aws_sdk", dataSourceData["sdk_type"], "SDK type should be aws_sdk")
					assert.Contains(t, dataSourceData["terraform_type"], "aws_", "Terraform type should start with aws_")
				}
			}
		})
	}
}

// TestAWSSDKDataSourcesIntegration_TerraformDataSourceConversion tests that AWS SDK data sources
// are properly converted to TerraformDataSource structs with correct metadata
func TestAWSSDKDataSourcesIntegration_TerraformDataSourceConversion(t *testing.T) {
	// Setup test data
	awsDataSource := AWSResourceInfo{
		TerraformType:   "aws_s3_bucket",
		FactoryFunction: "dataSourceBucket",
		Name:            "Bucket",
		SDKType:         "sdk",
	}

	service := ServiceRegistration{
		ServiceName:       "s3",
		PackagePath:       "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		AWSSDKDataSources: map[string]AWSResourceInfo{"aws_s3_bucket": awsDataSource},
	}

	// Execute conversion function (this will be implemented)
	terraformDataSource := NewTerraformDataSourceFromAWSSDK(awsDataSource, service)

	// Verify conversion results
	assert.Equal(t, "aws_s3_bucket", terraformDataSource.TerraformType)
	assert.Equal(t, "aws_sdk", terraformDataSource.SDKType)
	assert.Equal(t, "SDKDataSources", terraformDataSource.RegistrationMethod)
	assert.Equal(t, service.PackagePath, terraformDataSource.Namespace)
	assert.Equal(t, "", terraformDataSource.StructType) // SDK data sources don't have struct types

	// Verify that AWS-specific metadata is preserved in indexes
	assert.Contains(t, terraformDataSource.SchemaIndex, "dataSourceBucket")
}

// TestNewTerraformDataSourceFromAWSSDK_DataSourceMethodExtraction tests that the function
// properly uses extracted data source methods when available in serviceReg.DataSourceMethods
func TestNewTerraformDataSourceFromAWSSDK_DataSourceMethodExtraction(t *testing.T) {
	tests := []struct {
		name              string
		awsDataSource     AWSResourceInfo
		dataSourceMethods *LegacyDataSourceMethods
		expectedIndexes   map[string]string
		description       string
	}{
		{
			name: "Uses extracted data source methods when available",
			awsDataSource: AWSResourceInfo{
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "dataSourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
			},
			dataSourceMethods: &LegacyDataSourceMethods{
				ReadMethod: "dataSourceBucketRead",
			},
			expectedIndexes: map[string]string{
				"SchemaIndex":    "func.dataSourceBucket.goindex",
				"ReadIndex":      "func.dataSourceBucketRead.goindex",
				"AttributeIndex": "func.dataSourceBucket.goindex",
			},
			description: "When data source methods are extracted, should use specific method names for read operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup service registration with or without data source methods
			service := ServiceRegistration{
				ServiceName:       "s3",
				PackagePath:       "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
				AWSSDKDataSources: map[string]AWSResourceInfo{tt.awsDataSource.TerraformType: tt.awsDataSource},
				DataSourceMethods: make(map[string]*LegacyDataSourceMethods),
			}

			// Add data source methods if provided
			if tt.dataSourceMethods != nil {
				service.DataSourceMethods[tt.awsDataSource.TerraformType] = tt.dataSourceMethods
			}

			// Execute the function
			result := NewTerraformDataSourceFromAWSSDK(tt.awsDataSource, service)

			// Verify basic fields
			assert.Equal(t, tt.awsDataSource.TerraformType, result.TerraformType, "TerraformType should match")
			assert.Equal(t, "aws_sdk", result.SDKType, "SDKType should be aws_sdk")
			assert.Equal(t, "SDKDataSources", result.RegistrationMethod, "RegistrationMethod should be SDKDataSources")
			assert.Equal(t, "", result.StructType, "StructType should be empty for SDK data sources")

			// Verify exact index values - this is the critical test that will catch the bug
			assert.Equal(t, tt.expectedIndexes["SchemaIndex"], result.SchemaIndex,
				"SchemaIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["ReadIndex"], result.ReadIndex,
				"ReadIndex: %s", tt.description)
			assert.Equal(t, tt.expectedIndexes["AttributeIndex"], result.AttributeIndex,
				"AttributeIndex: %s", tt.description)
		})
	}
}

// TestAWSSDKDataSourcesIntegration_BackwardCompatibility tests that existing
// TerraformDataSource API remains unchanged while supporting AWS SDK data sources
func TestAWSSDKDataSourcesIntegration_BackwardCompatibility(t *testing.T) {
	// Setup AWS SDK data source
	awsService := ServiceRegistration{
		ServiceName: "s3",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
		AWSSDKDataSources: map[string]AWSResourceInfo{
			"aws_s3_bucket": {
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "dataSourceBucket",
				SDKType:         "sdk",
			},
		},
	}

	// Create TerraformDataSource from AWS SDK
	awsSDKDataSource := NewTerraformDataSourceFromAWSSDK(awsService.AWSSDKDataSources["aws_s3_bucket"], awsService)

	// Verify TerraformDataSource structure (backward compatibility)
	assert.IsType(t, TerraformDataSource{}, awsSDKDataSource)

	// Verify all required fields are present
	requiredFields := []string{"terraform_type", "namespace", "registration_method", "sdk_type"}

	awsJSON, err := json.Marshal(awsSDKDataSource)
	require.NoError(t, err)
	var awsData map[string]interface{}
	err = json.Unmarshal(awsJSON, &awsData)
	require.NoError(t, err)

	for _, field := range requiredFields {
		assert.Contains(t, awsData, field, "AWS SDK data source should have field: %s", field)
	}
}

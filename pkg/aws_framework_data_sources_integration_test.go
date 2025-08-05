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

// TestAWSFrameworkDataSourcesIntegration_WriteDataSourceFiles tests that AWS Framework data sources
// are correctly integrated into the main pipeline and written to files
func TestAWSFrameworkDataSourcesIntegration_WriteDataSourceFiles(t *testing.T) {
	tests := []struct {
		name                    string
		awsFrameworkDataSources map[string]AWSResourceInfo
		expectedFiles           []string
		expectedContentKeys     []string
	}{
		{
			name: "single_framework_data_source",
			awsFrameworkDataSources: map[string]AWSResourceInfo{
				"aws_test_data_source": {
					TerraformType:   "aws_test_data_source",
					FactoryFunction: "newTestDataSource",
					Name:            "Test Data Source",
					StructType:      "testDataSource",
				},
			},
			expectedFiles:       []string{"aws_test_data_source.json"},
			expectedContentKeys: []string{"terraform_type", "struct_type", "sdk_type", "registration_method"},
		},
		{
			name: "multiple_framework_data_sources",
			awsFrameworkDataSources: map[string]AWSResourceInfo{
				"aws_first_data_source": {
					TerraformType:   "aws_first_data_source",
					FactoryFunction: "newFirstDataSource",
					Name:            "First Data Source",
					StructType:      "firstDataSource",
				},
				"aws_second_data_source": {
					TerraformType:   "aws_second_data_source",
					FactoryFunction: "newSecondDataSource",
					Name:            "Second Data Source",
					StructType:      "secondDataSource",
				},
			},
			expectedFiles:       []string{"aws_first_data_source.json", "aws_second_data_source.json"},
			expectedContentKeys: []string{"terraform_type", "struct_type", "sdk_type", "registration_method"},
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

			// Create TerraformProviderIndex with AWS Framework data sources
			index := &TerraformProviderIndex{
				Version: "v5.0.0",
				Services: []ServiceRegistration{
					{
						ServiceName:             "testservice",
						PackagePath:             "github.com/hashicorp/terraform-provider-aws/internal/service/testservice",
						AWSFrameworkDataSources: tt.awsFrameworkDataSources,
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
					assert.Empty(t, files, "DataSources directory should be empty when no AWS Framework data sources")
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

					// Verify specific values for AWS Framework data sources
					assert.Equal(t, "aws_framework", dataSourceData["sdk_type"], "SDK type should be aws_framework")
					assert.Equal(t, "FrameworkDataSources", dataSourceData["registration_method"], "Registration method should be FrameworkDataSources")
					assert.Contains(t, dataSourceData["terraform_type"], "aws_", "Terraform type should start with aws_")
					assert.NotEmpty(t, dataSourceData["struct_type"], "Struct type should not be empty for Framework data sources")
				}
			}
		})
	}
}

// TestNewTerraformDataSourceFromAWSFramework_BasicMapping tests the conversion function
// that maps AWS Framework data source info to TerraformDataSource
func TestNewTerraformDataSourceFromAWSFramework_BasicMapping(t *testing.T) {
	awsDataSource := AWSResourceInfo{
		TerraformType:   "aws_test_data_source",
		FactoryFunction: "newTestDataSource",
		Name:            "Test Data Source",
		StructType:      "testDataSource",
	}

	service := ServiceRegistration{
		ServiceName: "testservice",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/testservice",
	}

	result := NewTerraformDataSourceFromAWSFramework(awsDataSource, service)

	expected := TerraformDataSource{
		TerraformType:      "aws_test_data_source",
		StructType:         "testDataSource",
		Namespace:          "github.com/hashicorp/terraform-provider-aws/internal/service/testservice",
		RegistrationMethod: "FrameworkDataSources",
		SDKType:            "aws_framework",
		SchemaIndex:        "method.testDataSource.Schema.goindex",
		ReadIndex:          "method.testDataSource.Read.goindex",
		AttributeIndex:     "method.testDataSource.Schema.goindex",
	}

	assert.Equal(t, expected, result)
}

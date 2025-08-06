package pkg

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAWSResourcesIntegration_WriteFiles tests that all AWS resource types
// are properly integrated into the file generation pipeline
func TestAWSResourcesIntegration_WriteFiles(t *testing.T) {
	tests := []struct {
		name            string
		resourceType    string
		resources       map[string]AWSResource
		expectedFiles   []string
		expectedSDKType string
		outputDir       string // "resources", "datasources", or "ephemeral"
	}{
		// SDK Resources
		{
			name:         "SDK Resources - Single resource creates JSON file",
			resourceType: "sdk_resources",
			resources: map[string]AWSResource{
				"aws_s3_bucket": TestSDKResourceS3Bucket,
			},
			expectedFiles:   []string{"aws_s3_bucket.json"},
			expectedSDKType: "aws_sdk",
			outputDir:       "resources",
		},
		{
			name:         "SDK Resources - Multiple resources create multiple JSON files",
			resourceType: "sdk_resources",
			resources: map[string]AWSResource{
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
			expectedFiles:   []string{"aws_s3_bucket.json", "aws_ec2_instance.json"},
			expectedSDKType: "aws_sdk",
			outputDir:       "resources",
		},
		// SDK DataSources
		{
			name:         "SDK DataSources - Single data source creates JSON file",
			resourceType: "sdk_datasources",
			resources: map[string]AWSResource{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					FactoryFunction: "dataSourceBucket",
					Name:            "Bucket",
					SDKType:         "sdk",
				},
			},
			expectedFiles:   []string{"aws_s3_bucket.json"},
			expectedSDKType: "aws_sdk",
			outputDir:       "datasources",
		},
		{
			name:         "SDK DataSources - Multiple data sources create multiple JSON files",
			resourceType: "sdk_datasources",
			resources: map[string]AWSResource{
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
			expectedFiles:   []string{"aws_s3_bucket.json", "aws_ec2_instance.json"},
			expectedSDKType: "aws_sdk",
			outputDir:       "datasources",
		},
		// Framework Resources
		{
			name:         "Framework Resources - Single resource creates JSON file",
			resourceType: "framework_resources",
			resources: map[string]AWSResource{
				"aws_s3_bucket": {
					TerraformType:   "aws_s3_bucket",
					Name:            "Bucket",
					FactoryFunction: "newBucketResource",
					SDKType:         "framework",
					StructType:      "bucketResource",
				},
			},
			expectedFiles:   []string{"aws_s3_bucket.json"},
			expectedSDKType: "aws_framework",
			outputDir:       "resources",
		},
		{
			name:         "Framework Resources - Multiple resources create multiple JSON files",
			resourceType: "framework_resources",
			resources: map[string]AWSResource{
				"aws_dynamodb_table": {
					TerraformType:   "aws_dynamodb_table",
					Name:            "Table",
					FactoryFunction: "newTableResource",
					SDKType:         "framework",
					StructType:      "tableResource",
				},
				"aws_lambda_function": {
					TerraformType:   "aws_lambda_function",
					Name:            "Function",
					FactoryFunction: "newFunctionResource",
					SDKType:         "framework",
					StructType:      "functionResource",
				},
			},
			expectedFiles:   []string{"aws_dynamodb_table.json", "aws_lambda_function.json"},
			expectedSDKType: "aws_framework",
			outputDir:       "resources",
		},
		// Framework DataSources
		{
			name:         "Framework DataSources - Single data source creates JSON file",
			resourceType: "framework_datasources",
			resources: map[string]AWSResource{
				"aws_bedrock_foundation_model": {
					TerraformType:   "aws_bedrock_foundation_model",
					Name:            "Foundation Model",
					FactoryFunction: "newFoundationModelDataSource",
					SDKType:         "framework",
					StructType:      "foundationModelDataSource",
				},
			},
			expectedFiles:   []string{"aws_bedrock_foundation_model.json"},
			expectedSDKType: "aws_framework",
			outputDir:       "datasources",
		},
		// Ephemeral Resources
		{
			name:         "Ephemeral Resources - Single ephemeral resource creates JSON file",
			resourceType: "ephemeral_resources",
			resources: map[string]AWSResource{
				"aws_secretsmanager_secret_value": {
					TerraformType:   "aws_secretsmanager_secret_value",
					Name:            "Secret Value",
					FactoryFunction: "newSecretValueEphemeralResource",
					SDKType:         "ephemeral",
					StructType:      "secretValueEphemeralResource",
				},
			},
			expectedFiles:   []string{"aws_secretsmanager_secret_value.json"},
			expectedSDKType: "ephemeral",
			outputDir:       "ephemeral",
		},
		{
			name:         "Ephemeral Resources - Multiple ephemeral resources create multiple JSON files",
			resourceType: "ephemeral_resources",
			resources: map[string]AWSResource{
				"aws_lambda_invocation": {
					TerraformType:   "aws_lambda_invocation",
					Name:            "Invocation",
					FactoryFunction: "newInvocationEphemeralResource",
					SDKType:         "ephemeral",
					StructType:      "invocationEphemeralResource",
				},
				"aws_ssm_parameter": {
					TerraformType:   "aws_ssm_parameter",
					Name:            "Parameter",
					FactoryFunction: "newParameterEphemeralResource",
					SDKType:         "ephemeral",
					StructType:      "parameterEphemeralResource",
				},
			},
			expectedFiles:   []string{"aws_lambda_invocation.json", "aws_ssm_parameter.json"},
			expectedSDKType: "ephemeral",
			outputDir:       "ephemeral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup filesystem
			fs := afero.NewMemMapFs()
			stub := gostub.Stub(&outputFs, fs)
			defer stub.Reset()

			outputDir := "/tmp/output"
			targetDir := filepath.Join(outputDir, tt.outputDir)

			// Create TerraformProviderIndex with appropriate resource type
			service := ServiceRegistration{
				ServiceName:              "test-service",
				PackagePath:              "github.com/hashicorp/terraform-provider-aws/internal/service/test",
				AWSSDKResources:          make(map[string]AWSResource),
				AWSSDKDataSources:        make(map[string]AWSResource),
				AWSFrameworkResources:    make(map[string]AWSResource),
				AWSFrameworkDataSources:  make(map[string]AWSResource),
				AWSEphemeralResources:    make(map[string]AWSResource),
				ResourceCRUDMethods:      make(map[string]*LegacyResourceCRUDFunctions),
				DataSourceMethods:        make(map[string]*LegacyDataSourceMethods),
				ResourceTerraformTypes:   make(map[string]string),
				DataSourceTerraformTypes: make(map[string]string),
				EphemeralTerraformTypes:  make(map[string]string),
			}

			// Assign resources to appropriate category
			switch tt.resourceType {
			case "sdk_resources":
				service.AWSSDKResources = tt.resources
			case "sdk_datasources":
				service.AWSSDKDataSources = tt.resources
			case "framework_resources":
				service.AWSFrameworkResources = tt.resources
			case "framework_datasources":
				service.AWSFrameworkDataSources = tt.resources
			case "ephemeral_resources":
				service.AWSEphemeralResources = tt.resources
			}

			index := &TerraformProviderIndex{
				Version:    "v5.0.0",
				Services:   []ServiceRegistration{service},
				Statistics: ProviderStatistics{},
			}

			// Execute appropriate write method
			var err error
			switch tt.outputDir {
			case "resources":
				err = index.WriteResourceFiles(outputDir, nil)
			case "datasources":
				err = index.WriteDataSourceFiles(outputDir, nil)
			case "ephemeral":
				err = index.WriteEphemeralFiles(outputDir, nil)
			}
			require.NoError(t, err)

			// Verify expected files were created
			if len(tt.expectedFiles) == 0 {
				// If no resources expected, verify directory is either empty or doesn't exist
				exists, err := afero.DirExists(fs, targetDir)
				if err == nil && exists {
					files, err := afero.ReadDir(fs, targetDir)
					require.NoError(t, err)
					assert.Empty(t, files, "Directory should be empty when no resources")
				}
			} else {
				// Verify each expected file exists and has correct content
				for _, expectedFile := range tt.expectedFiles {
					filePath := filepath.Join(targetDir, expectedFile)
					exists, err := afero.Exists(fs, filePath)
					require.NoError(t, err)
					assert.True(t, exists, "Expected file %s should exist", expectedFile)

					// Read and verify file content
					fileContent, err := afero.ReadFile(fs, filePath)
					require.NoError(t, err)

					var resourceData map[string]interface{}
					err = json.Unmarshal(fileContent, &resourceData)
					require.NoError(t, err)

					// Verify expected keys are present
					expectedKeys := []string{"terraform_type", "sdk_type", "namespace", "registration_method"}
					for _, key := range expectedKeys {
						assert.Contains(t, resourceData, key, "File should contain key: %s", key)
					}

					// Verify specific values
					assert.Equal(t, tt.expectedSDKType, resourceData["sdk_type"], "SDK type should match expected")
					assert.Contains(t, resourceData["terraform_type"], "aws_", "Terraform type should start with aws_")

					// For framework and ephemeral resources, verify struct_type
					if tt.resourceType == "framework_resources" || tt.resourceType == "framework_datasources" || tt.resourceType == "ephemeral_resources" {
						assert.Contains(t, resourceData, "struct_type", "Framework/Ephemeral resources should have struct_type")
						assert.NotEmpty(t, resourceData["struct_type"], "struct_type should not be empty")
					}
				}
			}
		})
	}
}

// TestAWSResourcesIntegration_TerraformResourceConversion tests that all AWS resource types
// are properly converted to their respective Terraform structs with correct metadata
func TestAWSResourcesIntegration_TerraformResourceConversion(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		awsResource  AWSResource
		expectedType string
		expectedReg  string
	}{
		{
			name:         "SDK Resource conversion",
			resourceType: "sdk_resource",
			awsResource: AWSResource{
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "resourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
			},
			expectedType: "aws_sdk",
			expectedReg:  "SDKResources",
		},
		{
			name:         "SDK DataSource conversion",
			resourceType: "sdk_datasource",
			awsResource: AWSResource{
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "dataSourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
			},
			expectedType: "aws_sdk",
			expectedReg:  "SDKDataSources",
		},
		{
			name:         "Framework Resource conversion",
			resourceType: "framework_resource",
			awsResource: AWSResource{
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "newBucketResource",
				Name:            "Bucket",
				SDKType:         "framework",
				StructType:      "bucketResource",
			},
			expectedType: "aws_framework",
			expectedReg:  "FrameworkResources",
		},
		{
			name:         "Framework DataSource conversion",
			resourceType: "framework_datasource",
			awsResource: AWSResource{
				TerraformType:   "aws_bedrock_foundation_model",
				FactoryFunction: "newFoundationModelDataSource",
				Name:            "Foundation Model",
				SDKType:         "framework",
				StructType:      "foundationModelDataSource",
			},
			expectedType: "aws_framework",
			expectedReg:  "FrameworkDataSources",
		},
		{
			name:         "Ephemeral Resource conversion",
			resourceType: "ephemeral_resource",
			awsResource: AWSResource{
				TerraformType:   "aws_lambda_invocation",
				FactoryFunction: "newInvocationEphemeralResource",
				Name:            "Invocation",
				SDKType:         "ephemeral",
				StructType:      "invocationEphemeralResource",
			},
			expectedType: "ephemeral",
			expectedReg:  "newInvocationEphemeralResource", // Ephemeral uses factory function as registration method
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := ServiceRegistration{
				ServiceName: "test-service",
				PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/test",
			}

			// Execute conversion function based on resource type and verify results
			switch tt.resourceType {
			case "sdk_resource":
				terraformResource := NewTerraformResourceFromAWSSDK(tt.awsResource, service)
				assert.Equal(t, tt.awsResource.TerraformType, terraformResource.TerraformType)
				assert.Equal(t, tt.expectedType, terraformResource.SDKType)
				assert.Equal(t, tt.expectedReg, terraformResource.RegistrationMethod)
				assert.Equal(t, service.PackagePath, terraformResource.Namespace)
				assert.Empty(t, terraformResource.StructType, "SDK resources should not have struct types")
				assert.Contains(t, terraformResource.SchemaIndex, tt.awsResource.FactoryFunction)
			case "sdk_datasource":
				terraformDataSource := NewTerraformDataSourceFromAWSSDK(tt.awsResource, service)
				assert.Equal(t, tt.awsResource.TerraformType, terraformDataSource.TerraformType)
				assert.Equal(t, tt.expectedType, terraformDataSource.SDKType)
				assert.Equal(t, tt.expectedReg, terraformDataSource.RegistrationMethod)
				assert.Equal(t, service.PackagePath, terraformDataSource.Namespace)
				assert.Empty(t, terraformDataSource.StructType, "SDK data sources should not have struct types")
				assert.Contains(t, terraformDataSource.SchemaIndex, tt.awsResource.FactoryFunction)
			case "framework_resource":
				terraformResource := NewTerraformResourceFromAWSFramework(tt.awsResource, service)
				assert.Equal(t, tt.awsResource.TerraformType, terraformResource.TerraformType)
				assert.Equal(t, tt.expectedType, terraformResource.SDKType)
				assert.Equal(t, tt.expectedReg, terraformResource.RegistrationMethod)
				assert.Equal(t, service.PackagePath, terraformResource.Namespace)
				assert.Equal(t, tt.awsResource.StructType, terraformResource.StructType)
				// Framework resources use method-based schema indexes
				assert.Contains(t, terraformResource.SchemaIndex, tt.awsResource.StructType)
			case "framework_datasource":
				terraformDataSource := NewTerraformDataSourceFromAWSFramework(tt.awsResource, service)
				assert.Equal(t, tt.awsResource.TerraformType, terraformDataSource.TerraformType)
				assert.Equal(t, tt.expectedType, terraformDataSource.SDKType)
				assert.Equal(t, tt.expectedReg, terraformDataSource.RegistrationMethod)
				assert.Equal(t, service.PackagePath, terraformDataSource.Namespace)
				assert.Equal(t, tt.awsResource.StructType, terraformDataSource.StructType)
				// Framework data sources use method-based schema indexes
				assert.Contains(t, terraformDataSource.SchemaIndex, tt.awsResource.StructType)
			case "ephemeral_resource":
				terraformEphemeral := NewTerraformEphemeralFromAWS(tt.awsResource, service)
				assert.Equal(t, tt.awsResource.TerraformType, terraformEphemeral.TerraformType)
				assert.Equal(t, tt.expectedType, terraformEphemeral.SDKType)
				// Ephemeral resources use factory function as registration method
				assert.Equal(t, tt.awsResource.FactoryFunction, terraformEphemeral.RegistrationMethod)
				assert.Equal(t, service.PackagePath, terraformEphemeral.Namespace)
				assert.Equal(t, tt.awsResource.StructType, terraformEphemeral.StructType)
				// Ephemeral resources use method-based schema indexes
				assert.Contains(t, terraformEphemeral.SchemaIndex, tt.awsResource.StructType)
			}
		})
	}
}

// TestAWSResourcesIntegration_CRUDMethodExtraction tests that CRUD methods
// are properly extracted and used when available in serviceReg
func TestAWSResourcesIntegration_CRUDMethodExtraction(t *testing.T) {
	tests := []struct {
		name            string
		awsResource     AWSResource
		crudMethods     *LegacyResourceCRUDFunctions
		dsMethod        *LegacyDataSourceMethods
		expectedIndexes map[string]string
		description     string
	}{
		{
			name: "SDK Resource with extracted CRUD methods",
			awsResource: AWSResource{
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
			description: "Should use specific CRUD method names for each operation",
		},
		{
			name: "SDK DataSource with extracted read method",
			awsResource: AWSResource{
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "dataSourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
			},
			dsMethod: &LegacyDataSourceMethods{
				ReadMethod: "dataSourceBucketRead",
			},
			expectedIndexes: map[string]string{
				"SchemaIndex":    "func.dataSourceBucket.goindex",
				"ReadIndex":      "func.dataSourceBucketRead.goindex",
				"AttributeIndex": "func.dataSourceBucket.goindex",
			},
			description: "Should use specific read method name for data source",
		},
		{
			name: "SDK Resource fallback when no CRUD methods available",
			awsResource: AWSResource{
				TerraformType:   "aws_ec2_instance",
				FactoryFunction: "resourceInstance",
				Name:            "Instance",
				SDKType:         "sdk",
			},
			crudMethods: nil, // No CRUD methods extracted
			expectedIndexes: map[string]string{
				"SchemaIndex":    "func.resourceInstance.goindex",
				"AttributeIndex": "func.resourceInstance.goindex",
				// CRUD indexes should be empty when no methods available
				"CreateIndex": "",
				"ReadIndex":   "",
				"UpdateIndex": "",
				"DeleteIndex": "",
			},
			description: "Should use factory function for schema but leave CRUD indexes empty when no methods available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup service registration with CRUD methods
			service := ServiceRegistration{
				ServiceName:         "test-service",
				PackagePath:         "github.com/hashicorp/terraform-provider-aws/internal/service/test",
				ResourceCRUDMethods: make(map[string]*LegacyResourceCRUDFunctions),
				DataSourceMethods:   make(map[string]*LegacyDataSourceMethods),
			}

			// Add CRUD methods if provided
			if tt.crudMethods != nil {
				service.ResourceCRUDMethods[tt.awsResource.TerraformType] = tt.crudMethods
			}
			if tt.dsMethod != nil {
				service.DataSourceMethods[tt.awsResource.TerraformType] = tt.dsMethod
			}

			// Execute conversion
			if tt.dsMethod != nil {
				terraformDataSource := NewTerraformDataSourceFromAWSSDK(tt.awsResource, service)
				// Verify that indexes match expected values
				for indexName, expectedValue := range tt.expectedIndexes {
					var actualValue string
					switch indexName {
					case "SchemaIndex":
						actualValue = terraformDataSource.SchemaIndex
					case "ReadIndex":
						actualValue = terraformDataSource.ReadIndex
					case "AttributeIndex":
						actualValue = terraformDataSource.AttributeIndex
					}
					assert.Equal(t, expectedValue, actualValue, "Index %s should match expected value. %s", indexName, tt.description)
				}
			} else {
				terraformResource := NewTerraformResourceFromAWSSDK(tt.awsResource, service)
				// Verify that indexes match expected values
				for indexName, expectedValue := range tt.expectedIndexes {
					var actualValue string
					switch indexName {
					case "SchemaIndex":
						actualValue = terraformResource.SchemaIndex
					case "CreateIndex":
						actualValue = terraformResource.CreateIndex
					case "ReadIndex":
						actualValue = terraformResource.ReadIndex
					case "UpdateIndex":
						actualValue = terraformResource.UpdateIndex
					case "DeleteIndex":
						actualValue = terraformResource.DeleteIndex
					case "AttributeIndex":
						actualValue = terraformResource.AttributeIndex
					}
					assert.Equal(t, expectedValue, actualValue, "Index %s should match expected value. %s", indexName, tt.description)
				}
			}
		})
	}
}

// TestAWSResourcesIntegration_ProcessServiceFile tests that annotation-based parsing
// correctly extracts all AWS resource types from service files
func TestAWSResourcesIntegration_ProcessServiceFile(t *testing.T) {
	tests := []struct {
		name                   string
		sourceCode             string
		expectedSDKResources   int
		expectedSDKDataSources int
		expectedFrameworkRes   int
		expectedFrameworkDS    int
		expectedEphemeral      int
	}{
		{
			name: "Mixed resource types in single service file",
			sourceCode: `package example

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKResource("aws_example_resource", name="Example Resource")
func resourceExample() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExampleCreate,
		ReadWithoutTimeout:   resourceExampleRead,
		UpdateWithoutTimeout: resourceExampleUpdate,
		DeleteWithoutTimeout: resourceExampleDelete,
	}
}

// @SDKDataSource("aws_example_data", name="Example Data")
func dataSourceExample() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExampleRead,
	}
}

// @FrameworkResource("aws_example_framework", name="Framework Resource")
func newExampleFrameworkResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &exampleFrameworkResource{}, nil
}

type exampleFrameworkResource struct {
	framework.ResourceWithModel[exampleFrameworkResourceModel]
}

func (r *exampleFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {}

// @FrameworkDataSource("aws_example_framework_ds", name="Framework DataSource")
func newExampleFrameworkDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &exampleFrameworkDataSource{}, nil
}

type exampleFrameworkDataSource struct {
	framework.DataSourceWithModel[exampleFrameworkDataSourceModel]
}

func (d *exampleFrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {}

// @EphemeralResource("aws_example_ephemeral", name="Ephemeral Resource")
func newExampleEphemeralResource(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &exampleEphemeralResource{}, nil
}

type exampleEphemeralResource struct {
	framework.EphemeralResourceWithModel[exampleEphemeralResourceModel]
}

func (e *exampleEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {}
`,
			expectedSDKResources:   1,
			expectedSDKDataSources: 1,
			expectedFrameworkRes:   1,
			expectedFrameworkDS:    1,
			expectedEphemeral:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source code into AST
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, "test.go", tt.sourceCode, parser.ParseComments)
			require.NoError(t, err, "Failed to parse source code")

			// Create a service registration
			serviceReg := ServiceRegistration{
				ServiceName:              "test-service",
				PackagePath:              "github.com/hashicorp/terraform-provider-aws/internal/service/test",
				AWSSDKResources:          make(map[string]AWSResource),
				AWSSDKDataSources:        make(map[string]AWSResource),
				AWSFrameworkResources:    make(map[string]AWSResource),
				AWSFrameworkDataSources:  make(map[string]AWSResource),
				AWSEphemeralResources:    make(map[string]AWSResource),
				ResourceCRUDMethods:      make(map[string]*LegacyResourceCRUDFunctions),
				DataSourceMethods:        make(map[string]*LegacyDataSourceMethods),
				ResourceTerraformTypes:   make(map[string]string),
				DataSourceTerraformTypes: make(map[string]string),
				EphemeralTerraformTypes:  make(map[string]string),
			}

			// Create gophon.FileInfo and PackageInfo
			fileInfo := &gophon.FileInfo{
				FilePath: "test.go",
				File:     astFile,
				Package:  "example",
			}

			packageInfo := &gophon.PackageInfo{
				Files: []*gophon.FileInfo{fileInfo},
			}

			// Process the file using annotation-based parsing
			err = parseAWSServiceFileWithAnnotations(packageInfo, &serviceReg)
			require.NoError(t, err)

			// Verify expected counts
			assert.Equal(t, tt.expectedSDKResources, len(serviceReg.AWSSDKResources), "SDK Resources count should match")
			assert.Equal(t, tt.expectedSDKDataSources, len(serviceReg.AWSSDKDataSources), "SDK DataSources count should match")
			assert.Equal(t, tt.expectedFrameworkRes, len(serviceReg.AWSFrameworkResources), "Framework Resources count should match")
			assert.Equal(t, tt.expectedFrameworkDS, len(serviceReg.AWSFrameworkDataSources), "Framework DataSources count should match")
			assert.Equal(t, tt.expectedEphemeral, len(serviceReg.AWSEphemeralResources), "Ephemeral Resources count should match")

			// Log results for debugging
			total := len(serviceReg.AWSSDKResources) + len(serviceReg.AWSSDKDataSources) +
				len(serviceReg.AWSFrameworkResources) + len(serviceReg.AWSFrameworkDataSources) +
				len(serviceReg.AWSEphemeralResources)

			t.Logf("Processed service file and found:")
			t.Logf("  - SDK Resources: %d", len(serviceReg.AWSSDKResources))
			t.Logf("  - SDK DataSources: %d", len(serviceReg.AWSSDKDataSources))
			t.Logf("  - Framework Resources: %d", len(serviceReg.AWSFrameworkResources))
			t.Logf("  - Framework DataSources: %d", len(serviceReg.AWSFrameworkDataSources))
			t.Logf("  - Ephemeral Resources: %d", len(serviceReg.AWSEphemeralResources))
			t.Logf("  - Total: %d", total)
		})
	}
}

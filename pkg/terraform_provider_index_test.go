package pkg

import (
	"encoding/json"
	"go/ast"
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

// Test data setup - now using shared test helpers
// Note: keeping this function for backwards compatibility, but it now delegates to shared helper
func createTestTerraformProviderIndex() *TerraformProviderIndex {
	return CreateTestTerraformProviderIndex()
}

func TestTerraformProviderIndex_WriteIndexFiles(t *testing.T) {
	// Setup
	sut := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := sut.WriteIndexFiles(outputDir, nil)

	// Verify
	require.NoError(t, err)

	// Check main index file
	mainIndexPath := filepath.Join(outputDir, "terraform-provider-aws-index.json")
	exists, err := afero.Exists(fs, mainIndexPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify main index content
	mainIndexData, err := afero.ReadFile(fs, mainIndexPath)
	require.NoError(t, err)

	var readIndex TerraformProviderIndex
	err = json.Unmarshal(mainIndexData, &readIndex)
	require.NoError(t, err)
	assert.Equal(t, sut.Version, readIndex.Version)
	assert.Equal(t, sut.Statistics.ServiceCount, readIndex.Statistics.ServiceCount)
}

func TestTerraformProviderIndex_WriteResourceFiles(t *testing.T) {
	// Setup
	sut := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := sut.WriteResourceFiles(outputDir, nil)

	// Verify
	require.NoError(t, err)

	// Check directory exists
	resourcesDir := filepath.Join(outputDir, "resources")
	exists, err := afero.DirExists(fs, resourcesDir)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check AWS SDK resource files
	awsSDKResourceFile := filepath.Join(resourcesDir, "aws_s3_bucket_policy.json")
	exists, err = afero.Exists(fs, awsSDKResourceFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify AWS SDK resource content
	resourceData, err := afero.ReadFile(fs, awsSDKResourceFile)
	require.NoError(t, err)

	var resourceInfo TerraformResource
	err = json.Unmarshal(resourceData, &resourceInfo)
	require.NoError(t, err)
	assert.Equal(t, "aws_s3_bucket_policy", resourceInfo.TerraformType)
	assert.Equal(t, "aws_sdk", resourceInfo.SDKType)
	assert.Equal(t, "SDKResources", resourceInfo.RegistrationMethod)

	// Check AWS Framework resource files
	awsFrameworkResourceFile := filepath.Join(resourcesDir, "aws_s3_bucket.json")
	exists, err = afero.Exists(fs, awsFrameworkResourceFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify Framework resource content
	frameworkResourceData, err := afero.ReadFile(fs, awsFrameworkResourceFile)
	require.NoError(t, err)

	var frameworkResourceInfo TerraformResource
	err = json.Unmarshal(frameworkResourceData, &frameworkResourceInfo)
	require.NoError(t, err)
	assert.Equal(t, "aws_s3_bucket", frameworkResourceInfo.TerraformType)
	assert.Equal(t, "aws_framework", frameworkResourceInfo.SDKType)
	assert.Equal(t, "bucketResource", frameworkResourceInfo.StructType)
}

func TestTerraformProviderIndex_WriteDataSourceFiles(t *testing.T) {
	// Setup
	index := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.WriteDataSourceFiles(outputDir, nil)

	// Verify
	require.NoError(t, err)

	// Check directory exists
	dataSourcesDir := filepath.Join(outputDir, "datasources")
	exists, err := afero.DirExists(fs, dataSourcesDir)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check AWS SDK data source files
	awsDataSourceFile := filepath.Join(dataSourcesDir, "aws_s3_bucket.json")
	exists, err = afero.Exists(fs, awsDataSourceFile)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify AWS SDK data source content
	dataSourceData, err := afero.ReadFile(fs, awsDataSourceFile)
	require.NoError(t, err)

	var dataSourceInfo TerraformDataSource
	err = json.Unmarshal(dataSourceData, &dataSourceInfo)
	require.NoError(t, err)
	assert.Equal(t, "aws_s3_bucket", dataSourceInfo.TerraformType)
	assert.Equal(t, "aws_sdk", dataSourceInfo.SDKType)
	assert.Equal(t, "SDKDataSources", dataSourceInfo.RegistrationMethod)

}

func TestTerraformProviderIndex_WriteEphemeralFiles(t *testing.T) {
	// Setup
	index := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.WriteEphemeralFiles(outputDir, nil)

	// Verify
	require.NoError(t, err)

	// Check directory exists
	ephemeralDir := filepath.Join(outputDir, "ephemeral")
	exists, err := afero.DirExists(fs, ephemeralDir)
	require.NoError(t, err)
	assert.True(t, exists)

	// Since our test data has no ephemeral resources, verify directory is empty
	files, err := afero.ReadDir(fs, ephemeralDir)
	require.NoError(t, err)
	assert.Empty(t, files, "Ephemeral directory should be empty when no ephemeral resources exist")
}

func TestTerraformProviderIndex_WriteMainIndexFile(t *testing.T) {
	// Setup
	index := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.WriteMainIndexFile(outputDir)

	// Verify
	require.NoError(t, err)

	mainIndexPath := filepath.Join(outputDir, "terraform-provider-aws-index.json")
	exists, err := afero.Exists(fs, mainIndexPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify content
	data, err := afero.ReadFile(fs, mainIndexPath)
	require.NoError(t, err)

	var readIndex TerraformProviderIndex
	err = json.Unmarshal(data, &readIndex)
	require.NoError(t, err)

	assert.Equal(t, index.Version, readIndex.Version)
	assert.Equal(t, len(index.Services), len(readIndex.Services))
	assert.Equal(t, index.Statistics, readIndex.Statistics)
}

func TestTerraformProviderIndex_CreateDirectoryStructure(t *testing.T) {
	// Setup
	index := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.CreateDirectoryStructure(outputDir)

	// Verify
	require.NoError(t, err)

	// Check all directories exist
	expectedDirs := []string{
		outputDir,
		filepath.Join(outputDir, "resources"),
		filepath.Join(outputDir, "datasources"),
		filepath.Join(outputDir, "ephemeral"),
	}

	for _, dir := range expectedDirs {
		exists, err := afero.DirExists(fs, dir)
		require.NoError(t, err, "Directory should exist: %s", dir)
		assert.True(t, exists, "Directory should exist: %s", dir)
	}
}

func TestTerraformProviderIndex_WriteJSONFile(t *testing.T) {
	// Setup
	index := createTestTerraformProviderIndex()
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	filePath := "/test/data.json"
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": []string{"a", "b", "c"},
	}

	// Execute
	err := index.WriteJSONFile(filePath, testData)

	// Verify
	require.NoError(t, err)

	exists, err := afero.Exists(fs, filePath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify content
	data, err := afero.ReadFile(fs, filePath)
	require.NoError(t, err)

	var readData map[string]interface{}
	err = json.Unmarshal(data, &readData)
	require.NoError(t, err)

	assert.Equal(t, "value1", readData["key1"])
	assert.Equal(t, float64(42), readData["key2"]) // JSON numbers are float64
	assert.Len(t, readData["key3"], 3)
}

func TestTerraformProviderIndex_WriteIndexFiles_EmptyIndex(t *testing.T) {
	// Setup - empty index
	index := &TerraformProviderIndex{
		Version:    "v3.0.0",
		Services:   []ServiceRegistration{},
		Statistics: ProviderStatistics{},
	}
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.WriteIndexFiles(outputDir, nil)

	// Verify
	require.NoError(t, err)

	// Check main index file exists
	mainIndexPath := filepath.Join(outputDir, "terraform-provider-aws-index.json")
	exists, err := afero.Exists(fs, mainIndexPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Check directories were created
	expectedDirs := []string{
		filepath.Join(outputDir, "resources"),
		filepath.Join(outputDir, "datasources"),
		filepath.Join(outputDir, "ephemeral"),
	}

	for _, dir := range expectedDirs {
		exists, err := afero.DirExists(fs, dir)
		require.NoError(t, err)
		assert.True(t, exists)
	}
}

func TestTerraformProviderIndex_WriteIndexFiles_FileSystemError(t *testing.T) {
	// Setup - read-only filesystem to trigger errors
	index := createTestTerraformProviderIndex()
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.WriteIndexFiles(outputDir, nil)

	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory structure")
}

func TestTerraformProviderIndex_WriteResourceFiles_NoResources(t *testing.T) {
	// Setup - index with no resources
	index := &TerraformProviderIndex{
		Services: []ServiceRegistration{
			{
				ServiceName:             "empty",
				AWSSDKResources:         map[string]AWSResourceInfo{},
				AWSFrameworkResources:   map[string]AWSResourceInfo{},
				AWSSDKDataSources:       map[string]AWSResourceInfo{},
				AWSFrameworkDataSources: map[string]AWSResourceInfo{},
				AWSEphemeralResources:   map[string]AWSResourceInfo{},
				ResourceTerraformTypes:  map[string]string{},
				ResourceCRUDMethods:     map[string]*LegacyResourceCRUDFunctions{},
				DataSourceMethods:       map[string]*LegacyDataSourceMethods{},
			},
		},
	}
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()
	outputDir := "/test/output"

	// Execute
	err := index.CreateDirectoryStructure(outputDir)
	require.NoError(t, err)
	err = index.WriteResourceFiles(outputDir, nil)

	// Verify - should succeed even with no resources
	require.NoError(t, err)

	// Check directory was created
	resourcesDir := filepath.Join(outputDir, "resources")
	exists, err := afero.DirExists(fs, resourcesDir)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExtractStructTypeFromEphemeralFunction(t *testing.T) {
	tests := []struct {
		name     string
		funcCode string
		expected string
	}{
		{
			name: "Standard ephemeral function with &StructName{} pattern",
			funcCode: `
package test
func NewKeyVaultSecretEphemeralResource() ephemeral.EphemeralResource {
	return &KeyVaultSecretEphemeralResource{}
}`,
			expected: "KeyVaultSecretEphemeralResource",
		},
		{
			name: "Ephemeral function with StructName{} pattern (no pointer)",
			funcCode: `
package test
func NewKeyVaultCertificateEphemeralResource() ephemeral.EphemeralResource {
	return KeyVaultCertificateEphemeralResource{}
}`,
			expected: "KeyVaultCertificateEphemeralResource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the function code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.funcCode, parser.ParseComments)
			require.NoError(t, err)

			// Find the function declaration
			var funcDecl *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					funcDecl = fn
					return false // Stop after finding the first function
				}
				return true
			})

			require.NotNil(t, funcDecl, "Function declaration not found")

			// Test the extraction
			result := extractStructTypeFromEphemeralFunction(funcDecl)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertFunctionNamesToStructNames(t *testing.T) {
	tests := []struct {
		name          string
		functionNames []string
		packageCode   string
		expected      []string
	}{
		{
			name:          "Standard ephemeral functions",
			functionNames: []string{"NewKeyVaultSecretEphemeralResource", "NewKeyVaultCertificateEphemeralResource"},
			packageCode: `
package test
func NewKeyVaultSecretEphemeralResource() ephemeral.EphemeralResource {
	return &KeyVaultSecretEphemeralResource{}
}
func NewKeyVaultCertificateEphemeralResource() ephemeral.EphemeralResource {
	return &KeyVaultCertificateEphemeralResource{}
}`,
			expected: []string{"KeyVaultSecretEphemeralResource", "KeyVaultCertificateEphemeralResource"},
		},
		{
			name:          "Mixed patterns - some functions found, some not",
			functionNames: []string{"NewFoundFunction", "NewNotFoundFunction"},
			packageCode: `
package test
func NewFoundFunction() ephemeral.EphemeralResource {
	return &FoundEphemeralResource{}
}`,
			expected: []string{"FoundEphemeralResource", "NotFoundFunction"}, // Not found falls back to string manipulation
		},
		{
			name:          "Function without New prefix",
			functionNames: []string{"CreateSomething"},
			packageCode: `
package test
func CreateSomething() ephemeral.EphemeralResource {
	return &SomethingEphemeralResource{}
}`,
			expected: []string{"SomethingEphemeralResource"}, // Should extract from AST
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock PackageInfo with the test functions
			packageInfo := createMockPackageInfoWithFunctions(t, tt.packageCode)

			// Test the conversion
			result := convertFunctionNamesToStructNames(tt.functionNames, packageInfo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a mock PackageInfo with functions parsed from code
func createMockPackageInfoWithFunctions(t *testing.T, packageCode string) *gophon.PackageInfo {
	// Parse the package code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", packageCode, parser.ParseComments)
	require.NoError(t, err)

	// Extract all function declarations
	var functions []*gophon.FunctionInfo
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			functions = append(functions, &gophon.FunctionInfo{
				Name:     fn.Name.Name,
				FuncDecl: fn,
			})
		}
		return true
	})

	return &gophon.PackageInfo{
		Functions: functions,
	}
}

package pkg

import (
	"go/parser"
	"go/token"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the integration of dynamic file detection with service registration processing
func TestProcessAWSServiceFile_IntegratesWithDynamicDetection(t *testing.T) {
	tests := []struct {
		name         string
		sourceCode   string
		expectedSDK  int
		expectedFW   int
		description  string
	}{
		{
			name: "Service file with SDK resources",
			sourceCode: `package s3

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
		},
		{
			Factory:  resourceBucketPolicy,
			TypeName: "aws_s3_bucket_policy", 
			Name:     "Bucket Policy",
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
		},
	}
}`,
			expectedSDK: 3, // 2 resources + 1 data source
			expectedFW:  0,
			description: "Should extract SDK resources and data sources",
		},
		{
			name: "Service file with Framework resources",
			sourceCode: `package s3

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceExample,
			TypeName: "aws_s3_example",
			Name:     "Example",
		},
	}
}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceExample,
			TypeName: "aws_s3_example_data",
			Name:     "Example Data",
		},
	}
}`,
			expectedSDK: 0,
			expectedFW:  2, // 1 resource + 1 data source
			description: "Should extract Framework resources and data sources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source code into an AST
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.sourceCode, parser.ParseComments)
			require.NoError(t, err)

			// Create a mock FileInfo
			fileInfo := &gophon.FileInfo{
				FileName: "service_package_gen.go",
				File:     file,
			}

			// Create a mock PackageInfo with this file
			packageInfo := &gophon.PackageInfo{
				Files: []*gophon.FileInfo{
					fileInfo,
				},
			}

			// Test that identifyServicePackageFile correctly identifies this file
			serviceFile, err := identifyServicePackageFile(packageInfo)
			require.NoError(t, err)
			require.NotNil(t, serviceFile)
			assert.Equal(t, fileInfo, serviceFile, "Should identify the correct service file")

			// Test that processAWSServiceFile correctly processes the identified file
			serviceReg := &ServiceRegistration{
				AWSSDKResources:          make(map[string]AWSResourceInfo),
				AWSSDKDataSources:        make(map[string]AWSResourceInfo),
				AWSFrameworkResources:    make(map[string]AWSResourceInfo),
				AWSFrameworkDataSources:  make(map[string]AWSResourceInfo),
				AWSEphemeralResources:    make(map[string]AWSResourceInfo),
			}

			processAWSServiceFile(serviceFile, serviceReg)

			// Verify the results
			totalSDK := len(serviceReg.AWSSDKResources) + len(serviceReg.AWSSDKDataSources)
			totalFW := len(serviceReg.AWSFrameworkResources) + len(serviceReg.AWSFrameworkDataSources)

			assert.Equal(t, tt.expectedSDK, totalSDK, "Should find expected number of SDK resources/data sources")
			assert.Equal(t, tt.expectedFW, totalFW, "Should find expected number of Framework resources/data sources")

			// Verify specific resources if SDK test
			if tt.expectedSDK > 0 {
				assert.Contains(t, serviceReg.AWSSDKResources, "aws_s3_bucket", "Should find aws_s3_bucket resource")
				if tt.expectedSDK >= 3 {
					assert.Contains(t, serviceReg.AWSSDKResources, "aws_s3_bucket_policy", "Should find aws_s3_bucket_policy resource")
					assert.Contains(t, serviceReg.AWSSDKDataSources, "aws_s3_bucket", "Should find aws_s3_bucket data source")
				}
			}

			// Verify specific resources if Framework test
			if tt.expectedFW > 0 {
				assert.Contains(t, serviceReg.AWSFrameworkResources, "aws_s3_example", "Should find aws_s3_example resource")
				assert.Contains(t, serviceReg.AWSFrameworkDataSources, "aws_s3_example_data", "Should find aws_s3_example_data data source")
			}
		})
	}
}

// Test the full integration workflow: identify service file -> process it
func TestDynamicFileDetectionIntegration_WorkflowTest(t *testing.T) {
	// Create a package with multiple files - only one should be selected
	helperCode := `package s3

import "fmt"

func helperFunction() {
	fmt.Println("just a helper")
}
`

	serviceCode := `package s3

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
		},
	}
}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceExample,
			TypeName: "aws_s3_example_data",
			Name:     "Example Data",
		},
	}
}`

	// Parse both files
	fset := token.NewFileSet()
	
	helperFile, err := parser.ParseFile(fset, "helper.go", helperCode, parser.ParseComments)
	require.NoError(t, err)
	
	serviceFile, err := parser.ParseFile(fset, "service_package_gen.go", serviceCode, parser.ParseComments)
	require.NoError(t, err)

	// Create mock FileInfos
	helperFileInfo := &gophon.FileInfo{
		FileName: "helper.go",
		File:     helperFile,
	}
	
	serviceFileInfo := &gophon.FileInfo{
		FileName: "service_package_gen.go",
		File:     serviceFile,
	}

	// Create PackageInfo with both files
	packageInfo := &gophon.PackageInfo{
		Files: []*gophon.FileInfo{
			helperFileInfo,
			serviceFileInfo,
		},
	}

	// Step 1: Dynamic file detection should identify the service file
	identifiedFile, err := identifyServicePackageFile(packageInfo)
	require.NoError(t, err)
	require.NotNil(t, identifiedFile)
	assert.Equal(t, serviceFileInfo, identifiedFile, "Should identify the service file, not the helper file")

	// Step 2: Process the identified file
	serviceReg := &ServiceRegistration{
		AWSSDKResources:          make(map[string]AWSResourceInfo),
		AWSSDKDataSources:        make(map[string]AWSResourceInfo),
		AWSFrameworkResources:    make(map[string]AWSResourceInfo),
		AWSFrameworkDataSources:  make(map[string]AWSResourceInfo),
		AWSEphemeralResources:    make(map[string]AWSResourceInfo),
	}

	processAWSServiceFile(identifiedFile, serviceReg)

	// Step 3: Verify the results
	assert.Len(t, serviceReg.AWSSDKResources, 1, "Should find 1 SDK resource")
	assert.Len(t, serviceReg.AWSFrameworkDataSources, 1, "Should find 1 Framework data source")
	assert.Len(t, serviceReg.AWSSDKDataSources, 0, "Should find 0 SDK data sources")
	assert.Len(t, serviceReg.AWSFrameworkResources, 0, "Should find 0 Framework resources")

	assert.Contains(t, serviceReg.AWSSDKResources, "aws_s3_bucket", "Should find the SDK resource")
	assert.Contains(t, serviceReg.AWSFrameworkDataSources, "aws_s3_example_data", "Should find the Framework data source")
}

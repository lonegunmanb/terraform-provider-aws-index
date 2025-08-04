package pkg

import (
	"go/parser"
	"go/token"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasAWSServiceMethods(t *testing.T) {
	tests := []struct {
		name           string
		sourceCode     string
		expectedResult bool
		description    string
	}{
		{
			name: "File with AWS SDK Resources method",
			sourceCode: `package service

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
}`,
			expectedResult: true,
			description:    "Should detect SDKResources method",
		},
		{
			name: "File with AWS SDK DataSources method",
			sourceCode: `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
		},
	}
}`,
			expectedResult: true,
			description:    "Should detect SDKDataSources method",
		},
		{
			name: "File with AWS Framework Resources method",
			sourceCode: `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceExample,
			TypeName: "aws_example_resource",
			Name:     "Example Resource",
		},
	}
}`,
			expectedResult: true,
			description:    "Should detect FrameworkResources method",
		},
		{
			name: "File with AWS Framework DataSources method",
			sourceCode: `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceExample,
			TypeName: "aws_example_datasource", 
			Name:     "Example DataSource",
		},
	}
}`,
			expectedResult: true,
			description:    "Should detect FrameworkDataSources method",
		},
		{
			name: "File with AWS Ephemeral Resources method",
			sourceCode: `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource {
	return []*inttypes.ServicePackageEphemeralResource{
		{
			Factory:  NewExampleEphemeralResource,
			TypeName: "aws_example_ephemeral",
			Name:     "Example Ephemeral",
		},
	}
}`,
			expectedResult: true,
			description:    "Should detect EphemeralResources method",
		},
		{
			name: "File with multiple AWS methods",
			sourceCode: `package service

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
			TypeName: "aws_example_datasource", 
			Name:     "Example DataSource",
		},
	}
}`,
			expectedResult: true,
			description:    "Should detect file with multiple AWS methods",
		},
		{
			name: "File with no AWS service methods",
			sourceCode: `package service

import (
	"context"
)

type servicePackage struct{}

func SomeOtherFunction() string {
	return "not an AWS service method"
}

func (p *servicePackage) SomeMethod(ctx context.Context) []string {
	return []string{"not", "aws", "service"}
}`,
			expectedResult: false,
			description:    "Should not detect file without AWS service methods",
		},
		{
			name: "Empty file",
			sourceCode: `package service

// Just a comment`,
			expectedResult: false,
			description:    "Should not detect empty file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the source code into AST
			fset := token.NewFileSet()
			astFile, err := parser.ParseFile(fset, "test.go", tt.sourceCode, parser.ParseComments)
			require.NoError(t, err, "Failed to parse source code")

			// Create a mock FileInfo
			fileInfo := &gophon.FileInfo{
				File: astFile,
			}

			// Test the hasAWSServiceMethods function
			result := hasAWSServiceMethods(fileInfo)
			assert.Equal(t, tt.expectedResult, result, tt.description)
		})
	}
}

// TestIdentifyServicePackageFiles tests the new behavior where all files with AWS methods are returned
func TestIdentifyServicePackageFiles(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedFileNames   []string
		expectError         bool
		expectedErrorSubstr string
		description         string
	}{
		{
			name: "Single file with AWS methods",
			files: map[string]string{
				"service_package_gen.go": `package service

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
}`,
				"other_file.go": `package service

func SomeOtherFunction() {
	// Not an AWS service file
}`,
			},
			expectedFileNames: []string{"service_package_gen.go"},
			expectError:       false,
			description:       "Should return single AWS service file",
		},
		{
			name: "Multiple files with AWS methods - return all",
			files: map[string]string{
				"service_package_gen.go": `package service

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
}`,
				"alternative_service.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceExample,
			TypeName: "aws_example_resource",
			Name:     "Example Resource",
		},
	}
}`,
				"other_file.go": `package service

func SomeOtherFunction() {
	// Not an AWS service file
}`,
			},
			expectedFileNames: []string{"service_package_gen.go", "alternative_service.go"},
			expectError:       false,
			description:       "Should return all files with AWS methods",
		},
		{
			name: "Three files with different AWS method types",
			files: map[string]string{
				"sdk_resources.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceExample1,
			TypeName: "aws_example1",
			Name:     "Example 1",
		},
	}
}`,
				"framework_data_sources.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource {
	return []*inttypes.ServicePackageFrameworkDataSource{
		{
			Factory:  newDataSourceExample,
			TypeName: "aws_example_datasource",
			Name:     "Example DataSource",
		},
	}
}`,
				"ephemeral_resources.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource {
	return []*inttypes.ServicePackageEphemeralResource{
		{
			Factory:  NewExampleEphemeralResource,
			TypeName: "aws_example_ephemeral",
			Name:     "Example Ephemeral",
		},
	}
}`,
				"regular_file.go": `package service

func RegularFunction() {
	// Not an AWS service file
}`,
			},
			expectedFileNames: []string{"sdk_resources.go", "framework_data_sources.go", "ephemeral_resources.go"},
			expectError:       false,
			description:       "Should return all three files with different AWS method types",
		},
		{
			name: "No files with AWS methods",
			files: map[string]string{
				"file1.go": `package service

func SomeFunction() {
	// Not an AWS service file
}`,
				"file2.go": `package service

func AnotherFunction() {
	// Also not an AWS service file
}`,
			},
			expectedFileNames:   []string{},
			expectError:         true,
			expectedErrorSubstr: "no AWS service methods found in package",
			description:         "Should return error when no AWS service files found",
		},
		{
			name:                "Empty package",
			files:               map[string]string{},
			expectedFileNames:   []string{},
			expectError:         true,
			expectedErrorSubstr: "no AWS service methods found in package",
			description:         "Should return error for empty package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock PackageInfo with files
			var files []*gophon.FileInfo

			for fileName, sourceCode := range tt.files {
				fset := token.NewFileSet()
				astFile, err := parser.ParseFile(fset, fileName, sourceCode, parser.ParseComments)
				require.NoError(t, err, "Failed to parse source code for %s", fileName)

				files = append(files, &gophon.FileInfo{
					FileName: fileName,
					File:     astFile,
				})
			}

			packageInfo := &gophon.PackageInfo{
				Files: files,
			}

			// Test the new identifyServicePackageFiles function (plural)
			results, err := identifyServicePackageFiles(packageInfo)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				if tt.expectedErrorSubstr != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorSubstr, tt.description)
				}
				assert.Empty(t, results, "Results should be empty when error occurs")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, results, "Results should not be empty when no error")

				// Extract actual file names from results
				var actualFileNames []string
				for _, fileInfo := range results {
					actualFileNames = append(actualFileNames, fileInfo.FileName)
				}

				// Sort both slices for comparison (order shouldn't matter)
				assert.ElementsMatch(t, tt.expectedFileNames, actualFileNames, tt.description)
			}
		})
	}
}

// TestProcessMultipleAWSServiceFiles tests the integration workflow where multiple AWS service files
// are processed and their results are merged into a single ServiceRegistration
func TestProcessMultipleAWSServiceFiles(t *testing.T) {
	tests := []struct {
		name                     string
		files                    map[string]string
		expectedSDKResourceCount int
		expectedFwResourceCount  int
		expectedEphemeralCount   int
		description              string
	}{
		{
			name: "Process two files with different AWS method types",
			files: map[string]string{
				"sdk_file.go": `package service

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
}`,
				"framework_file.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceExample,
			TypeName: "aws_example_resource",
			Name:     "Example Resource",
		},
	}
}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource {
	return []*inttypes.ServicePackageEphemeralResource{
		{
			Factory:  NewExampleEphemeralResource,
			TypeName: "aws_example_ephemeral",
			Name:     "Example Ephemeral",
		},
		{
			Factory:  NewAnotherEphemeralResource,
			TypeName: "aws_another_ephemeral",
			Name:     "Another Ephemeral",
		},
	}
}`,
			},
			expectedSDKResourceCount: 2,
			expectedFwResourceCount:  1,
			expectedEphemeralCount:   2,
			description:              "Should merge results from multiple files with different method types",
		},
		{
			name: "Process three files with overlapping method types",
			files: map[string]string{
				"file1.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceExample1,
			TypeName: "aws_example1",
			Name:     "Example 1",
		},
	}
}`,
				"file2.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceExample2,
			TypeName: "aws_example2",
			Name:     "Example 2",
		},
	}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceExample2,
			TypeName: "aws_framework_example2",
			Name:     "Framework Example 2",
		},
	}
}`,
				"file3.go": `package service

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceExample3,
			TypeName: "aws_framework_example3",
			Name:     "Framework Example 3",
		},
	}
}`,
			},
			expectedSDKResourceCount: 2, // from file1.go and file2.go
			expectedFwResourceCount:  2, // from file2.go and file3.go
			expectedEphemeralCount:   0,
			description:              "Should merge overlapping method types from multiple files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock PackageInfo with files
			var files []*gophon.FileInfo

			for fileName, sourceCode := range tt.files {
				fset := token.NewFileSet()
				astFile, err := parser.ParseFile(fset, fileName, sourceCode, parser.ParseComments)
				require.NoError(t, err, "Failed to parse source code for %s", fileName)

				files = append(files, &gophon.FileInfo{
					FileName: fileName,
					File:     astFile,
				})
			}

			packageInfo := &gophon.PackageInfo{
				Files: files,
			}

			// Step 1: Identify all AWS service files
			serviceFiles, err := identifyServicePackageFiles(packageInfo)
			require.NoError(t, err, "Should identify service files successfully")
			require.NotEmpty(t, serviceFiles, "Should find service files")

			// Step 2: Create a ServiceRegistration and process all files
			serviceReg := &ServiceRegistration{
				AWSSDKResources:         make(map[string]AWSResourceInfo),
				AWSSDKDataSources:       make(map[string]AWSResourceInfo),
				AWSFrameworkResources:   make(map[string]AWSResourceInfo),
				AWSFrameworkDataSources: make(map[string]AWSResourceInfo),
				AWSEphemeralResources:   make(map[string]AWSResourceInfo),
			}

			// Step 3: Process each AWS service file and merge results
			for _, fileInfo := range serviceFiles {
				processAWSServiceFile(fileInfo, serviceReg)
			}

			// Step 4: Verify merged results
			assert.Equal(t, tt.expectedSDKResourceCount, len(serviceReg.AWSSDKResources),
				"SDK Resources count should match expected: %s", tt.description)

			assert.Equal(t, tt.expectedFwResourceCount, len(serviceReg.AWSFrameworkResources),
				"Framework Resources count should match expected: %s", tt.description)

			assert.Equal(t, tt.expectedEphemeralCount, len(serviceReg.AWSEphemeralResources),
				"Ephemeral Resources count should match expected: %s", tt.description)
		})
	}
}

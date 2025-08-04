package pkg

import (
	"go/parser"
	"go/token"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentifyServicePackageFiles tests the new behavior where all files with AWS methods are returned
func TestIdentifyServicePackageFiles(t *testing.T) {
	tests := []struct {
		name              string
		files             map[string]string
		expectedFileNames []string
		expectError       bool
		expectedErrorSubstr string
		description       string
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

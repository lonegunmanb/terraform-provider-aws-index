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

func TestIdentifyServicePackageFile(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedFileName    string
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
			expectedFileName: "service_package_gen.go",
			expectError:      false,
			description:      "Should identify the correct AWS service file",
		},
		{
			name: "Multiple files with AWS methods - prefer service_package naming",
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
			},
			expectedFileName: "service_package_gen.go",
			expectError:      false,
			description:      "Should prefer file with service_package naming convention",
		},
		{
			name: "Multiple files with AWS methods - use first found (simplified)",
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
}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource {
	return []*inttypes.ServicePackageEphemeralResource{
		{
			Factory:  NewEphemeralExample2,
			TypeName: "aws_ephemeral_example2",
			Name:     "Ephemeral Example 2",
		},
	}
}`,
			},
			expectedFileName: "file1.go",
			expectError:      false,
			description:      "Should use first file found with AWS methods (simplified approach)",
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
			expectedFileName:    "",
			expectError:         true,
			expectedErrorSubstr: "no AWS service methods found in package",
			description:         "Should return error when no AWS service files found",
		},
		{
			name: "Empty package",
			files: map[string]string{},
			expectedFileName:    "",
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

			// Test the identifyServicePackageFile function
			result, err := identifyServicePackageFile(packageInfo)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				if tt.expectedErrorSubstr != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorSubstr, tt.description)
				}
				assert.Nil(t, result, "Result should be nil when error occurs")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, result, "Result should not be nil when no error")
				assert.Equal(t, tt.expectedFileName, result.FileName, tt.description)
			}
		})
	}
}

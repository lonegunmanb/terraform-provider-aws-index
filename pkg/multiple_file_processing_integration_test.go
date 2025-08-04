package pkg

import (
	"go/parser"
	"go/token"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				AWSSDKResources:           make(map[string]AWSResourceInfo),
				AWSSDKDataSources:         make(map[string]AWSResourceInfo),
				AWSFrameworkResources:     make(map[string]AWSResourceInfo),
				AWSFrameworkDataSources:   make(map[string]AWSResourceInfo),
				AWSEphemeralResources:     make(map[string]AWSResourceInfo),
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

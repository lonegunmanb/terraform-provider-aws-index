package pkg

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdentifyServicePackageFileLogging tests the logging behavior of the simplified single-file approach
func TestIdentifyServicePackageFileLogging(t *testing.T) {
	tests := []struct {
		name             string
		files            map[string]string
		expectedFileName string
		expectError      bool
		expectedOutput   string
		description      string
	}{
		{
			name: "Single file found - should log success",
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
			expectedOutput:   "Found AWS service file: service_package_gen.go\n",
			description:      "Should log when single AWS service file is found",
		},
		{
			name: "Multiple files found - should warn and use first",
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

type altServicePackage struct{}

func (p *altServicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceSomething,
			TypeName: "aws_s3_bucket_policy",
			Name:     "Bucket Policy",
		},
	}
}`,
			},
			expectedFileName: "alternative_service.go", // First in alphabetical order
			expectError:      false,
			expectedOutput:   "Warning: Multiple AWS service files found in package, using first: alternative_service.go\nFound AWS service file: alternative_service.go\n",
			description:      "Should warn about multiple files and use first found",
		},
		{
			name: "No AWS service files found - should not log success",
			files: map[string]string{
				"other_file.go": `package service

func SomeOtherFunction() {
	// Not an AWS service file
}`,
				"another_file.go": `package service

func AnotherFunction() {
	// Also not an AWS service file
}`,
			},
			expectError:    true,
			expectedOutput: "",
			description:    "Should not log success when no AWS service files are found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout for testing logging
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create PackageInfo with test files
			packageInfo := &gophon.PackageInfo{
				Files: make([]*gophon.FileInfo, 0),
			}

			for fileName, sourceCode := range tt.files {
				fset := token.NewFileSet()
				node, err := parser.ParseFile(fset, fileName, sourceCode, parser.ParseComments)
				require.NoError(t, err, "Failed to parse test source code for %s", fileName)

				fileInfo := &gophon.FileInfo{
					FileName: fileName,
					File:     node,
				}
				packageInfo.Files = append(packageInfo.Files, fileInfo)
			}

			// Call the function under test 
			result, err := identifyServicePackageFile(packageInfo)

			// Restore stdout and capture output
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			actualOutput := buf.String()

			// Verify results
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err, tt.description)
				require.NotNil(t, result, "Expected non-nil result")
				assert.Equal(t, tt.expectedFileName, result.FileName, tt.description)
			}

			// Verify logging output
			assert.Equal(t, tt.expectedOutput, actualOutput, "Logging output should match expected")
		})
	}
}

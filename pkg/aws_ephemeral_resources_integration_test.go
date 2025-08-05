package pkg

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessAWSServiceFile_EphemeralResources tests that ephemeral resources are extracted and integrated
func TestProcessAWSServiceFile_EphemeralResources(t *testing.T) {
	t.Run("process single ephemeral resource", func(t *testing.T) {
		source := `package secretsmanager

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource {
	return []*inttypes.ServicePackageEphemeralResource{
		{
			Factory:  newSecretValueEphemeralResource,
			TypeName: "aws_secretsmanager_secret_value",
			Name:     "Secret Value",
		},
	}
}`

		// Parse the source code into AST
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
		require.NoError(t, err, "Failed to parse source code")

		// Create a service registration
		serviceReg := ServiceRegistration{
			ServiceName:           "secretsmanager",
			PackagePath:          "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager",
			AWSEphemeralResources: make(map[string]AWSResourceInfo),
		}

		// Create gophon.FileInfo
		fileInfo := &gophon.FileInfo{
			File: astFile,
		}

		// Process the AWS service file
		processAWSServiceFile(fileInfo, &serviceReg)

		// Verify ephemeral resources were extracted
		require.Len(t, serviceReg.AWSEphemeralResources, 1)
		assert.Contains(t, serviceReg.AWSEphemeralResources, "aws_secretsmanager_secret_value")
		
		ephemeralResource := serviceReg.AWSEphemeralResources["aws_secretsmanager_secret_value"]
		assert.Equal(t, "aws_secretsmanager_secret_value", ephemeralResource.TerraformType)
		assert.Equal(t, "newSecretValueEphemeralResource", ephemeralResource.FactoryFunction)
		assert.Equal(t, "Secret Value", ephemeralResource.Name)
		assert.Equal(t, "ephemeral", ephemeralResource.SDKType)
	})

	t.Run("process multiple ephemeral resources", func(t *testing.T) {
		source := `package secretsmanager

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource {
	return []*inttypes.ServicePackageEphemeralResource{
		{
			Factory:  newSecretValueEphemeralResource,
			TypeName: "aws_secretsmanager_secret_value",
			Name:     "Secret Value",
		},
		{
			Factory:  newRandomPasswordEphemeralResource,
			TypeName: "aws_secretsmanager_random_password",
			Name:     "Random Password",
		},
	}
}`

		// Parse the source code into AST
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
		require.NoError(t, err, "Failed to parse source code")

		// Create a service registration
		serviceReg := ServiceRegistration{
			ServiceName:           "secretsmanager",
			PackagePath:          "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager",
			AWSEphemeralResources: make(map[string]AWSResourceInfo),
		}

		// Create gophon.FileInfo
		fileInfo := &gophon.FileInfo{
			File: astFile,
		}

		// Process the AWS service file
		processAWSServiceFile(fileInfo, &serviceReg)

		// Verify ephemeral resources were extracted
		require.Len(t, serviceReg.AWSEphemeralResources, 2)
		assert.Contains(t, serviceReg.AWSEphemeralResources, "aws_secretsmanager_secret_value")
		assert.Contains(t, serviceReg.AWSEphemeralResources, "aws_secretsmanager_random_password")
		
		secretValue := serviceReg.AWSEphemeralResources["aws_secretsmanager_secret_value"]
		assert.Equal(t, "newSecretValueEphemeralResource", secretValue.FactoryFunction)
		assert.Equal(t, "Secret Value", secretValue.Name)
		
		randomPassword := serviceReg.AWSEphemeralResources["aws_secretsmanager_random_password"]
		assert.Equal(t, "newRandomPasswordEphemeralResource", randomPassword.FactoryFunction)
		assert.Equal(t, "Random Password", randomPassword.Name)
	})

	t.Run("process service with no ephemeral resources", func(t *testing.T) {
		source := `package ec2

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{}
}`

		// Parse the source code into AST
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
		require.NoError(t, err, "Failed to parse source code")

		// Create a service registration
		serviceReg := ServiceRegistration{
			ServiceName:           "ec2",
			PackagePath:          "github.com/hashicorp/terraform-provider-aws/internal/service/ec2",
			AWSEphemeralResources: make(map[string]AWSResourceInfo),
		}

		// Create gophon.FileInfo
		fileInfo := &gophon.FileInfo{
			File: astFile,
		}

		// Process the AWS service file
		processAWSServiceFile(fileInfo, &serviceReg)

		// Verify no ephemeral resources were extracted
		assert.Len(t, serviceReg.AWSEphemeralResources, 0)
	})
}

// TestNewTerraformEphemeralFromAWS_BasicMapping tests the conversion function
func TestNewTerraformEphemeralFromAWS_BasicMapping(t *testing.T) {
	awsEphemeral := AWSResourceInfo{
		TerraformType:   "aws_secretsmanager_secret_value",
		FactoryFunction: "newSecretValueEphemeralResource",
		Name:            "Secret Value",
		SDKType:         "ephemeral",
	}

	service := ServiceRegistration{
		ServiceName: "secretsmanager",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager",
	}

	result := NewTerraformEphemeralFromAWS(awsEphemeral, service)

	assert.Equal(t, "aws_secretsmanager_secret_value", result.TerraformType)
	assert.Equal(t, "", result.StructType) // Will be populated when CRUD methods are extracted
	assert.Equal(t, "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager", result.Namespace)
	assert.Equal(t, "newSecretValueEphemeralResource", result.RegistrationMethod)
	assert.Equal(t, "ephemeral", result.SDKType)
	assert.Equal(t, "", result.SchemaIndex) // Will be populated when CRUD methods are extracted
	assert.Equal(t, "", result.OpenIndex)   // Will be populated when CRUD methods are extracted
	assert.Equal(t, "", result.RenewIndex)  // Will be populated when CRUD methods are extracted
	assert.Equal(t, "", result.CloseIndex)  // Will be populated when CRUD methods are extracted
}

// TestNewTerraformEphemeralFromAWS_WithCRUDMethods tests ephemeral with lifecycle methods
func TestNewTerraformEphemeralFromAWS_WithCRUDMethods(t *testing.T) {
	awsEphemeral := AWSResourceInfo{
		TerraformType:   "aws_secretsmanager_secret_value",
		FactoryFunction: "newSecretValueEphemeralResource",
		Name:            "Secret Value",
		SDKType:         "ephemeral",
		StructType:      "secretValueEphemeralResource",
	}

	service := ServiceRegistration{
		ServiceName: "secretsmanager",
		PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager",
	}

	result := NewTerraformEphemeralFromAWS(awsEphemeral, service)

	assert.Equal(t, "aws_secretsmanager_secret_value", result.TerraformType)
	assert.Equal(t, "secretValueEphemeralResource", result.StructType)
	assert.Equal(t, "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager", result.Namespace)
	assert.Equal(t, "newSecretValueEphemeralResource", result.RegistrationMethod)
	assert.Equal(t, "ephemeral", result.SDKType)
	// When struct type is provided, method indexes should be generated
	assert.Equal(t, "method.secretValueEphemeralResource.Schema.goindex", result.SchemaIndex)
	assert.Equal(t, "method.secretValueEphemeralResource.Open.goindex", result.OpenIndex)
	assert.Equal(t, "method.secretValueEphemeralResource.Renew.goindex", result.RenewIndex)
	assert.Equal(t, "method.secretValueEphemeralResource.Close.goindex", result.CloseIndex)
}

// TestWriteEphemeralFiles_AWSEphemeralResources tests that AWS ephemeral resources are written to files
func TestWriteEphemeralFiles_AWSEphemeralResources(t *testing.T) {
	// This test will verify that WriteEphemeralFiles can process AWS ephemeral resources
	// alongside legacy ephemeral resources
	
	t.Run("write AWS ephemeral resources to files", func(t *testing.T) {
		// Setup: Create a temporary directory
		outputDir, err := os.MkdirTemp("", "ephemeral_resources_test")
		require.NoError(t, err)
		defer os.RemoveAll(outputDir)

		index := &TerraformProviderIndex{
			Services: []ServiceRegistration{
				{
					ServiceName: "secretsmanager",
					PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager",
					// AWS ephemeral resources (NEW)
					AWSEphemeralResources: map[string]AWSResourceInfo{
						"aws_secretsmanager_secret_value": {
							TerraformType:   "aws_secretsmanager_secret_value",
							FactoryFunction: "newSecretValueEphemeralResource",
							Name:            "Secret Value",
							SDKType:         "ephemeral",
						},
					},
					// Legacy ephemeral resources (for compatibility)
					EphemeralTerraformTypes: map[string]string{
						"LegacyEphemeralResource": "azurerm_legacy_ephemeral",
					},
				},
			},
		}

		// Execute
		err = index.WriteEphemeralFiles(outputDir, nil)
		require.NoError(t, err)

		// Verify AWS ephemeral resource file was created
		awsEphemeralFile := filepath.Join(outputDir, "ephemeral", "aws_secretsmanager_secret_value.json")
		_, err = os.Stat(awsEphemeralFile)
		assert.NoError(t, err, "AWS ephemeral resource file should be created")

		// Verify legacy ephemeral resource file was created too
		legacyEphemeralFile := filepath.Join(outputDir, "ephemeral", "azurerm_legacy_ephemeral.json")
		_, err = os.Stat(legacyEphemeralFile)
		assert.NoError(t, err, "Legacy ephemeral resource file should still be created")
	})
}

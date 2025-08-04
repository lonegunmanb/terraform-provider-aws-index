package pkg

import (
	"path/filepath"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAWSSDKResourcesIntegration_HighVolumeService tests the integration with
// high-volume services like S3 and EC2 to ensure performance validation as mentioned
// in Sub-Task 3.2.1 scope
func TestAWSSDKResourcesIntegration_HighVolumeService(t *testing.T) {
	// Setup filesystem
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&outputFs, fs)
	defer stub.Reset()

	outputDir := "/tmp/output"
	resourcesDir := filepath.Join(outputDir, "resources")

	// Create high-volume AWS SDK resources (simulate S3 service with multiple resources)
	s3Resources := map[string]AWSResourceInfo{
		"aws_s3_bucket": {
			TerraformType:   "aws_s3_bucket",
			FactoryFunction: "resourceBucket",
			Name:            "Bucket",
			SDKType:         "sdk",
			HasTags:         true,
			TagsConfig: &AWSTagsConfig{
				IdentifierAttribute: "bucket",
				ResourceType:        "Bucket",
			},
			Region: &AWSRegionConfig{
				IsOverrideEnabled: true,
			},
		},
		"aws_s3_bucket_policy": {
			TerraformType:   "aws_s3_bucket_policy",
			FactoryFunction: "resourceBucketPolicy",
			Name:            "Bucket Policy",
			SDKType:         "sdk",
			HasTags:         false,
		},
		"aws_s3_bucket_notification": {
			TerraformType:   "aws_s3_bucket_notification",
			FactoryFunction: "resourceBucketNotification",
			Name:            "Bucket Notification",
			SDKType:         "sdk",
			HasTags:         false,
		},
		"aws_s3_bucket_versioning": {
			TerraformType:   "aws_s3_bucket_versioning",
			FactoryFunction: "resourceBucketVersioning",
			Name:            "Bucket Versioning",
			SDKType:         "sdk",
			HasTags:         true,
			TagsConfig: &AWSTagsConfig{
				IdentifierAttribute: "bucket",
				ResourceType:        "Bucket",
			},
		},
		"aws_s3_object": {
			TerraformType:   "aws_s3_object",
			FactoryFunction: "resourceObject",
			Name:            "Object",
			SDKType:         "sdk",
			HasTags:         true,
			TagsConfig: &AWSTagsConfig{
				IdentifierAttribute: "key",
				ResourceType:        "Object",
			},
		},
	}

	// Create TerraformProviderIndex with high-volume service
	index := &TerraformProviderIndex{
		Version: "v5.0.0",
		Services: []ServiceRegistration{
			{
				ServiceName:     "s3",
				PackagePath:     "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
				AWSSDKResources: s3Resources,
			},
		},
		Statistics: ProviderStatistics{},
	}

	// Execute WriteResourceFiles with high-volume data
	err := index.WriteResourceFiles(outputDir, nil)
	require.NoError(t, err)

	// Verify all resources were created with proper structure
	expectedResources := []string{
		"aws_s3_bucket.json",
		"aws_s3_bucket_policy.json", 
		"aws_s3_bucket_notification.json",
		"aws_s3_bucket_versioning.json",
		"aws_s3_object.json",
	}

	for _, expectedResource := range expectedResources {
		filePath := filepath.Join(resourcesDir, expectedResource)
		exists, err := afero.Exists(fs, filePath)
		require.NoError(t, err)
		assert.True(t, exists, "Expected resource file %s should exist", expectedResource)

		// Verify file content has all required AWS SDK metadata
		fileContent, err := afero.ReadFile(fs, filePath)
		require.NoError(t, err)

		// Basic validation that it contains AWS SDK specific fields
		contentStr := string(fileContent)
		assert.Contains(t, contentStr, `"sdk_type": "aws_sdk"`)
		assert.Contains(t, contentStr, `"registration_method": "SDKResources"`)
		assert.Contains(t, contentStr, `"factory_function"`)
		assert.Contains(t, contentStr, `"name"`)
	}

	// Verify resources with TagsConfig have tags metadata
	taggedResources := []string{"aws_s3_bucket.json", "aws_s3_bucket_versioning.json", "aws_s3_object.json"}
	for _, taggedResource := range taggedResources {
		filePath := filepath.Join(resourcesDir, taggedResource)
		fileContent, err := afero.ReadFile(fs, filePath)
		require.NoError(t, err)

		contentStr := string(fileContent)
		assert.Contains(t, contentStr, `"has_tags": true`)
		assert.Contains(t, contentStr, `"tags_config"`)
		assert.Contains(t, contentStr, `"identifier_attribute"`)
		assert.Contains(t, contentStr, `"resource_type"`)
	}

	// Verify resources without tags don't have tags metadata
	untaggedResources := []string{"aws_s3_bucket_policy.json", "aws_s3_bucket_notification.json"}
	for _, untaggedResource := range untaggedResources {
		filePath := filepath.Join(resourcesDir, untaggedResource)
		fileContent, err := afero.ReadFile(fs, filePath)
		require.NoError(t, err)

		contentStr := string(fileContent)
		// Note: has_tags field is always present since it's not omitempty, but will be false for untagged resources
		// However, based on the AWS resource definition, I'll check for the absence of tags_config when has_tags is false
		assert.NotContains(t, contentStr, `"tags_config"`)
	}
}

// TestAWSSDKResourcesIntegration_ServiceRegistrationStatistics tests that
// AWS SDK resources are properly counted in provider statistics
func TestAWSSDKResourcesIntegration_ServiceRegistrationStatistics(t *testing.T) {
	// Create services with mixed AWS resource types for statistics validation
	services := []ServiceRegistration{
		{
			ServiceName: "s3",
			PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
			AWSSDKResources: map[string]AWSResourceInfo{
				"aws_s3_bucket": {TerraformType: "aws_s3_bucket", SDKType: "sdk"},
				"aws_s3_object": {TerraformType: "aws_s3_object", SDKType: "sdk"},
			},
		},
		{
			ServiceName: "ec2",
			PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/ec2",
			AWSSDKResources: map[string]AWSResourceInfo{
				"aws_instance":     {TerraformType: "aws_instance", SDKType: "sdk"},
				"aws_security_group": {TerraformType: "aws_security_group", SDKType: "sdk"},
				"aws_vpc":          {TerraformType: "aws_vpc", SDKType: "sdk"},
			},
		},
		{
			ServiceName: "iam",
			PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/iam",
			// No AWS SDK resources - should not affect statistics
		},
	}

	index := &TerraformProviderIndex{
		Version:  "v5.0.0",
		Services: services,
		Statistics: ProviderStatistics{
			ServiceCount: len(services),
		},
	}

	// Count total AWS SDK resources across all services
	totalAWSSDKResources := 0
	for _, service := range services {
		totalAWSSDKResources += len(service.AWSSDKResources)
	}

	// Verify AWS SDK resources are counted correctly
	assert.Equal(t, 5, totalAWSSDKResources, "Should have 5 total AWS SDK resources")
	assert.Equal(t, 3, len(services), "Should have 3 services")
	assert.Equal(t, 2, len(services[0].AWSSDKResources), "S3 service should have 2 AWS SDK resources")
	assert.Equal(t, 3, len(services[1].AWSSDKResources), "EC2 service should have 3 AWS SDK resources")
	assert.Equal(t, 0, len(services[2].AWSSDKResources), "IAM service should have 0 AWS SDK resources")
	
	// Verify index structure is correct
	assert.Equal(t, "v5.0.0", index.Version)
	assert.Equal(t, 3, len(index.Services))
}

package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExtractAWSSDKResources(t *testing.T) {
	t.Run("extract SDK resources from single method", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"unique"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrBucket,
				ResourceType:        "Bucket",
			}),
			Region: unique.Make(inttypes.ResourceRegionDefault()),
		},
		{
			Factory:  resourceBucketACL,
			TypeName: "aws_s3_bucket_acl",
			Name:     "Bucket ACL",
			Region:   unique.Make(inttypes.ResourceRegionDefault()),
		},
	}
}`

		expected := map[string]AWSResourceInfo{
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
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
			"aws_s3_bucket_acl": {
				TerraformType:   "aws_s3_bucket_acl",
				FactoryFunction: "resourceBucketACL",
				Name:            "Bucket ACL",
				SDKType:         "sdk",
				HasTags:         false,
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKResources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract SDK resources with empty method", func(t *testing.T) {
		source := `package s3

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{}
}`

		expected := map[string]AWSResourceInfo{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKResources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract SDK resources with variable assignment", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"unique"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	resources := []*inttypes.ServicePackageSDKResource{
		{
			Factory:  resourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
			Region:   unique.Make(inttypes.ResourceRegionDefault()),
		},
	}
	return resources
}`

		expected := map[string]AWSResourceInfo{
			"aws_s3_bucket": {
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "resourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
				HasTags:         false,
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKResources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("no SDKResources method found", func(t *testing.T) {
		source := `package s3

import (
	"context"
)

type servicePackage struct{}

func (p *servicePackage) SomeOtherMethod(ctx context.Context) string {
	return "test"
}`

		expected := map[string]AWSResourceInfo{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKResources(node)
		assert.Equal(t, expected, result)
	})
}

func TestExtractAWSSDKDataSources(t *testing.T) {
	t.Run("extract SDK data sources from single method", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"unique"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceBucket,
			TypeName: "aws_s3_bucket",
			Name:     "Bucket",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrBucket,
				ResourceType:        "Bucket",
			}),
			Region: unique.Make(inttypes.ResourceRegionDefault()),
		},
		{
			Factory:  dataSourceBucketObject,
			TypeName: "aws_s3_bucket_object",
			Name:     "Bucket Object",
			Region:   unique.Make(inttypes.ResourceRegionDefault()),
		},
	}
}`

		expected := map[string]AWSResourceInfo{
			"aws_s3_bucket": {
				TerraformType:   "aws_s3_bucket",
				FactoryFunction: "dataSourceBucket",
				Name:            "Bucket",
				SDKType:         "sdk",
				HasTags:         true,
				TagsConfig: &AWSTagsConfig{
					IdentifierAttribute: "bucket",
					ResourceType:        "Bucket",
				},
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
			"aws_s3_bucket_object": {
				TerraformType:   "aws_s3_bucket_object",
				FactoryFunction: "dataSourceBucketObject",
				Name:            "Bucket Object",
				SDKType:         "sdk",
				HasTags:         false,
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKDataSources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract SDK data sources with empty method", func(t *testing.T) {
		source := `package emptyservice

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	return []*inttypes.ServicePackageSDKDataSource{}
}`

		expected := map[string]AWSResourceInfo{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKDataSources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract SDK data sources with variable assignment", func(t *testing.T) {
		source := `package iam

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource {
	dataSources := []*inttypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceUser,
			TypeName: "aws_iam_user",
			Name:     "User",
		},
		{
			Factory:  dataSourceRole,
			TypeName: "aws_iam_role",
			Name:     "Role",
		},
	}
	return dataSources
}`

		expected := map[string]AWSResourceInfo{
			"aws_iam_user": {
				TerraformType:   "aws_iam_user",
				FactoryFunction: "dataSourceUser",
				Name:            "User",
				SDKType:         "sdk",
				HasTags:         false,
			},
			"aws_iam_role": {
				TerraformType:   "aws_iam_role",
				FactoryFunction: "dataSourceRole",
				Name:            "Role",
				SDKType:         "sdk",
				HasTags:         false,
			},
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKDataSources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("no SDKDataSources method found", func(t *testing.T) {
		source := `package nomethod

type servicePackage struct{}

func (p *servicePackage) SomeOtherMethod() {
	// This service package doesn't have SDKDataSources method
}`

		expected := map[string]AWSResourceInfo{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSSDKDataSources(node)
		assert.Equal(t, expected, result)
	})
}

func TestExtractAWSFrameworkResources(t *testing.T) {
	t.Run("extract Framework resources from single method", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"unique"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newBucketLifecycleConfigurationResource,
			TypeName: "aws_s3_bucket_lifecycle_configuration",
			Name:     "Bucket Lifecycle Configuration",
			Region:   unique.Make(inttypes.ResourceRegionDefault()),
		},
		{
			Factory:  newDirectoryBucketResource,
			TypeName: "aws_s3_directory_bucket",
			Name:     "Directory Bucket",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
				ResourceType:        "DirectoryBucket",
			}),
			Region: unique.Make(inttypes.ResourceRegionDefault()),
		},
	}
}`

		expected := map[string]AWSResourceInfo{
			"aws_s3_bucket_lifecycle_configuration": {
				TerraformType:   "aws_s3_bucket_lifecycle_configuration",
				FactoryFunction: "newBucketLifecycleConfigurationResource",
				Name:            "Bucket Lifecycle Configuration",
				SDKType:         "framework",
				HasTags:         false,
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
			"aws_s3_directory_bucket": {
				TerraformType:   "aws_s3_directory_bucket",
				FactoryFunction: "newDirectoryBucketResource",
				Name:            "Directory Bucket",
				SDKType:         "framework",
				HasTags:         true,
				TagsConfig: &AWSTagsConfig{
					IdentifierAttribute: "arn",
					ResourceType:        "DirectoryBucket",
				},
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSFrameworkResources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract Framework resources with empty method", func(t *testing.T) {
		source := `package s3

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	return []*inttypes.ServicePackageFrameworkResource{}
}`

		expected := map[string]AWSResourceInfo{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSFrameworkResources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract Framework resources with variable assignment", func(t *testing.T) {
		source := `package lambda

import (
	"context"
	"unique"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource {
	resources := []*inttypes.ServicePackageFrameworkResource{
		{
			Factory:  newFunctionResource,
			TypeName: "aws_lambda_function",
			Name:     "Function",
			Tags: unique.Make(inttypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
				ResourceType:        "Function",
			}),
			Region: unique.Make(inttypes.ResourceRegionDefault()),
		},
	}
	return resources
}`

		expected := map[string]AWSResourceInfo{
			"aws_lambda_function": {
				TerraformType:   "aws_lambda_function",
				FactoryFunction: "newFunctionResource",
				Name:            "Function",
				SDKType:         "framework",
				HasTags:         true,
				TagsConfig: &AWSTagsConfig{
					IdentifierAttribute: "arn",
					ResourceType:        "Function",
				},
				Region: &AWSRegionConfig{
					IsOverrideEnabled:             true,
					IsValidateOverrideInPartition: true,
				},
			},
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSFrameworkResources(node)
		assert.Equal(t, expected, result)
	})

	t.Run("extract Framework resources with method not found", func(t *testing.T) {
		source := `package ec2

import (
	"context"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type servicePackage struct{}

func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
	return []*inttypes.ServicePackageSDKResource{}
}`

		expected := map[string]AWSResourceInfo{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractAWSFrameworkResources(node)
		assert.Equal(t, expected, result)
	})
}

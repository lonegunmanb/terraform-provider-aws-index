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

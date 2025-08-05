package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
)

// TestHelpers provides common test data and utilities for unit tests
// This reduces duplication across test files

// CreateTestTerraformProviderIndex creates a standard test TerraformProviderIndex
// with sample data across all AWS resource types
func CreateTestTerraformProviderIndex() *TerraformProviderIndex {
	return &TerraformProviderIndex{
		Version: "v5.0.0",
		Services: []ServiceRegistration{
			{
				ServiceName: "s3",
				PackagePath: "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
				AWSSDKResources: map[string]AWSResourceInfo{
					"aws_s3_bucket_policy": {
						TerraformType:   "aws_s3_bucket_policy",
						Name:            "BucketPolicy",
						FactoryFunction: "resourceBucketPolicy",
						SDKType:         "aws_sdk",
					},
				},
				AWSSDKDataSources: map[string]AWSResourceInfo{
					"aws_s3_bucket": {
						TerraformType:   "aws_s3_bucket",
						Name:            "Bucket",
						FactoryFunction: "dataSourceS3Bucket",
						SDKType:         "aws_sdk",
					},
				},
				AWSFrameworkResources: map[string]AWSResourceInfo{
					"aws_s3_bucket": {
						TerraformType:   "aws_s3_bucket",
						Name:            "Bucket",
						FactoryFunction: "newBucketResource",
						SDKType:         "aws_framework",
						StructType:      "bucketResource",
					},
				},
				AWSFrameworkDataSources: make(map[string]AWSResourceInfo),
				AWSEphemeralResources:   make(map[string]AWSResourceInfo),
				ResourceTerraformTypes: map[string]string{
					"bucketResource": "aws_s3_bucket",
				},
				DataSourceTerraformTypes: make(map[string]string),
				EphemeralTerraformTypes:  make(map[string]string),
				ResourceCRUDMethods: map[string]*LegacyResourceCRUDFunctions{
					"aws_s3_bucket_policy": {
						CreateMethod: "resourceBucketPolicyCreate",
						ReadMethod:   "resourceBucketPolicyRead",
						UpdateMethod: "resourceBucketPolicyUpdate",
						DeleteMethod: "resourceBucketPolicyDelete",
					},
				},
				DataSourceMethods: map[string]*LegacyDataSourceMethods{
					"aws_s3_bucket": {
						ReadMethod: "dataSourceS3BucketRead",
					},
				},
			},
		},
		Statistics: ProviderStatistics{
			ServiceCount:       1,
			TotalResources:     2, // 1 SDK + 1 Framework = 2 total
			TotalDataSources:   1, // 1 SDK data source
			LegacyResources:    0, // No longer used
			ModernResources:    0, // No longer used
			EphemeralResources: 0, // No ephemeral resources in test data
		},
	}
}

// CreateTestServiceRegistration creates a standard test ServiceRegistration
// with empty collections ready for specific test cases
func CreateTestServiceRegistration(serviceName string) ServiceRegistration {
	return ServiceRegistration{
		ServiceName:                 serviceName,
		PackagePath:                 "github.com/hashicorp/terraform-provider-aws/internal/service/" + serviceName,
		AWSSDKResources:             make(map[string]AWSResourceInfo),
		AWSSDKDataSources:           make(map[string]AWSResourceInfo),
		AWSFrameworkResources:       make(map[string]AWSResourceInfo),
		AWSFrameworkDataSources:     make(map[string]AWSResourceInfo),
		AWSEphemeralResources:       make(map[string]AWSResourceInfo),
		ResourceCRUDMethods:         make(map[string]*LegacyResourceCRUDFunctions),
		DataSourceMethods:           make(map[string]*LegacyDataSourceMethods),
		ResourceTerraformTypes:      make(map[string]string),
		DataSourceTerraformTypes:    make(map[string]string),
		EphemeralTerraformTypes:     make(map[string]string),
		functions:                   make(map[string]*gophon.FunctionInfo),
	}
}

// CreateTestAWSResourceInfo creates common test resource info for different types
func CreateTestAWSResourceInfo(resourceType, terraformType, factoryFunction, name string) AWSResourceInfo {
	info := AWSResourceInfo{
		TerraformType:   terraformType,
		Name:            name,
		FactoryFunction: factoryFunction,
	}

	switch resourceType {
	case "sdk_resource", "sdk_datasource":
		info.SDKType = "sdk"
	case "framework_resource", "framework_datasource":
		info.SDKType = "framework"
		// Framework resources typically have struct types derived from factory function
		if factoryFunction != "" && len(factoryFunction) > 3 && factoryFunction[:3] == "new" {
			// Convert "newBucketResource" to "bucketResource"
			info.StructType = factoryFunction[3:]
		}
	case "ephemeral_resource":
		info.SDKType = "ephemeral"
		// Ephemeral resources typically have struct types derived from factory function
		if factoryFunction != "" && len(factoryFunction) > 3 && factoryFunction[:3] == "new" {
			// Convert "newBucketEphemeralResource" to "bucketEphemeralResource"
			info.StructType = factoryFunction[3:]
		}
	}

	return info
}

// Sample resource test data for common test scenarios
var (
	// Common SDK Resource
	TestSDKResourceS3Bucket = CreateTestAWSResourceInfo(
		"sdk_resource",
		"aws_s3_bucket",
		"resourceBucket",
		"Bucket",
	)

	// Common SDK DataSource
	TestSDKDataSourceS3Bucket = CreateTestAWSResourceInfo(
		"sdk_datasource",
		"aws_s3_bucket",
		"dataSourceBucket",
		"Bucket",
	)

	// Common Framework Resource
	TestFrameworkResourceS3Bucket = CreateTestAWSResourceInfo(
		"framework_resource",
		"aws_s3_bucket",
		"newBucketResource",
		"Bucket",
	)

	// Common Framework DataSource
	TestFrameworkDataSourceS3Bucket = CreateTestAWSResourceInfo(
		"framework_datasource",
		"aws_s3_bucket",
		"newBucketDataSource",
		"Bucket",
	)

	// Common Ephemeral Resource
	TestEphemeralResourceLambdaInvocation = CreateTestAWSResourceInfo(
		"ephemeral_resource",
		"aws_lambda_invocation",
		"newInvocationEphemeralResource",
		"Invocation",
	)

	// Common CRUD methods
	TestCRUDMethods = &LegacyResourceCRUDFunctions{
		CreateMethod: "resourceBucketCreate",
		ReadMethod:   "resourceBucketRead",
		UpdateMethod: "resourceBucketUpdate",
		DeleteMethod: "resourceBucketDelete",
	}

	// Common DataSource read method
	TestDataSourceMethods = &LegacyDataSourceMethods{
		ReadMethod: "dataSourceBucketRead",
	}
)

// CreateTestPackageInfo creates a mock PackageInfo for testing annotation scanning
func CreateTestPackageInfo(serviceName string, files []*gophon.FileInfo) *gophon.PackageInfo {
	return &gophon.PackageInfo{
		Files:     files,
		Functions: []*gophon.FunctionInfo{}, // Functions will be discovered by tests
	}
}

// CreateExpectedJSON creates expected JSON content for resource file testing
func CreateExpectedJSON(terraformType, sdkType, namespace, registrationMethod string) map[string]interface{} {
	return map[string]interface{}{
		"terraform_type":      terraformType,
		"sdk_type":           sdkType,
		"namespace":          namespace,
		"registration_method": registrationMethod,
	}
}

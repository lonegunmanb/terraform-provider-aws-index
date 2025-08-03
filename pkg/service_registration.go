package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"os"
)

// ServiceRegistration represents all registration methods found in a single service package
type ServiceRegistration struct {
	Package              *gophon.PackageInfo                     `json:"-"`
	ServiceName          string                                  `json:"service_name"`           // "s3", "ec2", etc.
	PackagePath          string                                  `json:"package_path"`           // "internal/service/s3"
	
	// AWS 5-category structure (NEW)
	AWSSDKResources          map[string]AWSResourceInfo `json:"aws_sdk_resources"`          // SDK resources from SDKResources()
	AWSSDKDataSources        map[string]AWSResourceInfo `json:"aws_sdk_data_sources"`       // SDK data sources from SDKDataSources()
	AWSFrameworkResources    map[string]AWSResourceInfo `json:"aws_framework_resources"`    // Framework resources from FrameworkResources()
	AWSFrameworkDataSources  map[string]AWSResourceInfo `json:"aws_framework_data_sources"` // Framework data sources from FrameworkDataSources()
	AWSEphemeralResources    map[string]AWSResourceInfo `json:"aws_ephemeral_resources"`    // Ephemeral resources from EphemeralResources()
	
	// Legacy AzureRM structure (DEPRECATED - will be removed)
	SupportedResources   map[string]string                       `json:"supported_resources,omitempty"`    // Legacy map-based resources
	SupportedDataSources map[string]string                       `json:"supported_data_sources,omitempty"` // Legacy map-based data sources
	Resources            []string                                `json:"resources,omitempty"`              // Modern slice-based resources
	DataSources          []string                                `json:"data_sources,omitempty"`           // Modern slice-based data sources
	EphemeralFunctions   []string                                `json:"ephemeral_functions,omitempty"`    // Function-based ephemeral resources
	ResourceCRUDMethods  map[string]*LegacyResourceCRUDFunctions `json:"resource_crud_methods,omitempty"`  // CRUD methods for legacy resources
	DataSourceMethods    map[string]*LegacyDataSourceMethods     `json:"data_source_methods,omitempty"`    // Methods for legacy data sources
	// New mappings between Terraform types and struct types
	ResourceTerraformTypes   map[string]string `json:"resource_terraform_types,omitempty"`    // StructType -> TerraformType for modern resources
	DataSourceTerraformTypes map[string]string `json:"data_source_terraform_types,omitempty"` // StructType -> TerraformType for modern data sources
	EphemeralTerraformTypes  map[string]string `json:"ephemeral_terraform_types,omitempty"`   // StructType -> TerraformType for ephemeral resources
}

func newServiceRegistration(packageInfo *gophon.PackageInfo, entry os.DirEntry) ServiceRegistration {
	return ServiceRegistration{
		Package:                  packageInfo,
		ServiceName:              entry.Name(),
		PackagePath:              packageInfo.Files[0].Package,
		
		// AWS 5-category structure (NEW)
		AWSSDKResources:         make(map[string]AWSResourceInfo),
		AWSSDKDataSources:       make(map[string]AWSResourceInfo),
		AWSFrameworkResources:   make(map[string]AWSResourceInfo),
		AWSFrameworkDataSources: make(map[string]AWSResourceInfo),
		AWSEphemeralResources:   make(map[string]AWSResourceInfo),
		
		// Legacy AzureRM structure (DEPRECATED)
		SupportedResources:       make(map[string]string),
		SupportedDataSources:     make(map[string]string),
		Resources:                []string{},
		DataSources:              []string{},
		EphemeralFunctions:       []string{},
		ResourceCRUDMethods:      make(map[string]*LegacyResourceCRUDFunctions),
		DataSourceMethods:        make(map[string]*LegacyDataSourceMethods),
		ResourceTerraformTypes:   make(map[string]string),
		DataSourceTerraformTypes: make(map[string]string),
		EphemeralTerraformTypes:  make(map[string]string),
	}
}

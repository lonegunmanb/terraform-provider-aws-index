package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"os"
)

// ServiceRegistration represents all registration methods found in a single service package
type ServiceRegistration struct {
	Package     *gophon.PackageInfo `json:"-"`
	ServiceName string              `json:"service_name"` // "s3", "ec2", etc.
	PackagePath string              `json:"package_path"` // "internal/service/s3"

	// AWS 5-category structure (NEW)
	AWSSDKResources         map[string]AWSResource `json:"aws_sdk_resources"`          // SDK resources from SDKResources()
	AWSSDKDataSources       map[string]AWSResource `json:"aws_sdk_data_sources"`       // SDK data sources from SDKDataSources()
	AWSFrameworkResources   map[string]AWSResource `json:"aws_framework_resources"`    // Framework resources from FrameworkResources()
	AWSFrameworkDataSources map[string]AWSResource `json:"aws_framework_data_sources"` // Framework data sources from FrameworkDataSources()
	AWSEphemeralResources   map[string]AWSResource `json:"aws_ephemeral_resources"`    // Ephemeral resources from EphemeralResources()

	// Terraform type mappings for Framework resources (struct-based)
	ResourceTerraformTypes   map[string]string `json:"resource_terraform_types,omitempty"`    // StructType -> TerraformType for Framework resources
	DataSourceTerraformTypes map[string]string `json:"data_source_terraform_types,omitempty"` // StructType -> TerraformType for Framework data sources
	EphemeralTerraformTypes  map[string]string `json:"ephemeral_terraform_types,omitempty"`   // StructType -> TerraformType for ephemeral resources

	// CRUD method mappings for SDK resources (function-based)
	ResourceCRUDMethods map[string]*LegacyResourceCRUDFunctions `json:"resource_crud_methods,omitempty"` // CRUD methods for SDK resources
	DataSourceMethods   map[string]*LegacyDataSourceMethods     `json:"data_source_methods,omitempty"`   // Methods for SDK data sources
}

func newServiceRegistration(packageInfo *gophon.PackageInfo, entry os.FileInfo) ServiceRegistration {
	svc := ServiceRegistration{
		Package:     packageInfo,
		ServiceName: entry.Name(),
		PackagePath: packageInfo.Files[0].Package,

		// AWS 5-category structure (NEW)
		AWSSDKResources:         make(map[string]AWSResource),
		AWSSDKDataSources:       make(map[string]AWSResource),
		AWSFrameworkResources:   make(map[string]AWSResource),
		AWSFrameworkDataSources: make(map[string]AWSResource),
		AWSEphemeralResources:   make(map[string]AWSResource),

		// Terraform type mappings and CRUD methods
		ResourceCRUDMethods:      make(map[string]*LegacyResourceCRUDFunctions),
		DataSourceMethods:        make(map[string]*LegacyDataSourceMethods),
		ResourceTerraformTypes:   make(map[string]string),
		DataSourceTerraformTypes: make(map[string]string),
		EphemeralTerraformTypes:  make(map[string]string),
	}
	return svc
}

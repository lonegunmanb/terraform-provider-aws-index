package pkg

import "fmt"

// TerraformDataSource represents information about a Terraform data source
type TerraformDataSource struct {
	TerraformType      string `json:"terraform_type"`
	StructType         string `json:"struct_type"`
	Namespace          string `json:"namespace"`           // "github.com/hashicorp/terraform-provider-aws/internal/service"
	RegistrationMethod string `json:"registration_method"` // "func.SupportedDataSources", "DataSources", etc.
	SDKType            string `json:"sdk_type"`            // "legacy_pluginsdk", "modern_sdk"
	SchemaIndex        string `json:"schema_index,omitempty"`
	ReadIndex          string `json:"read_index,omitempty"`
	AttributeIndex     string `json:"attribute_index,omitempty"`
}

// NewTerraformDataSourceInfo creates a TerraformDataSource struct
func NewTerraformDataSourceInfo(terraformType, structType, registrationMethod, sdkType string, serviceReg ServiceRegistration) TerraformDataSource {
	if sdkType == "legacy_pluginsdk" {
		return TerraformDataSource{
			TerraformType:      terraformType,
			StructType:         "",
			Namespace:          serviceReg.PackagePath,
			RegistrationMethod: registrationMethod,
			SDKType:            sdkType,
			// Optional fields can be added later when we have more sophisticated AST parsing
			SchemaIndex:    fmt.Sprintf("func.%s.goindex", registrationMethod),
			ReadIndex:      fmt.Sprintf("func.%s.goindex", serviceReg.DataSourceMethods[terraformType].ReadMethod),
			AttributeIndex: fmt.Sprintf("func.%s.goindex", registrationMethod),
		}
	}
	return TerraformDataSource{
		TerraformType:      serviceReg.DataSourceTerraformTypes[structType],
		StructType:         structType,
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "",
		SDKType:            sdkType,
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaIndex:    fmt.Sprintf("method.%s.Arguments.goindex", structType),
		ReadIndex:      fmt.Sprintf("method.%s.Read.goindex", structType),
		AttributeIndex: fmt.Sprintf("method.%s.Attributes.goindex", structType),
	}
}

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
		readMethod := registrationMethod // Default fallback
		if serviceReg.DataSourceMethods != nil {
			if dataSourceMethods, exists := serviceReg.DataSourceMethods[terraformType]; exists && dataSourceMethods != nil && dataSourceMethods.ReadMethod != "" {
				readMethod = dataSourceMethods.ReadMethod
			}
		}

		return TerraformDataSource{
			TerraformType:      terraformType,
			StructType:         "",
			Namespace:          serviceReg.PackagePath,
			RegistrationMethod: registrationMethod,
			SDKType:            sdkType,
			// Optional fields can be added later when we have more sophisticated AST parsing
			SchemaIndex:    fmt.Sprintf("func.%s.goindex", registrationMethod),
			ReadIndex:      fmt.Sprintf("func.%s.goindex", readMethod),
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

// NewTerraformDataSourceFromAWSSDK creates a TerraformDataSource struct from AWS SDK data source info
func NewTerraformDataSourceFromAWSSDK(awsDataSource AWSResourceInfo, serviceReg ServiceRegistration) TerraformDataSource {
	// Use specific data source methods if available, otherwise fall back to factory function
	schemaIndex := fmt.Sprintf("func.%s.goindex", awsDataSource.FactoryFunction)
	var readIndex string
	attributeIndex := fmt.Sprintf("func.%s.goindex", awsDataSource.FactoryFunction)

	// Check if we have extracted data source methods for this terraform type
	if dataSourceMethods, exists := serviceReg.DataSourceMethods[awsDataSource.TerraformType]; exists && dataSourceMethods != nil {
		if dataSourceMethods.ReadMethod != "" {
			readIndex = fmt.Sprintf("func.%s.goindex", dataSourceMethods.ReadMethod)
		}
	}

	return TerraformDataSource{
		TerraformType:      awsDataSource.TerraformType,
		StructType:         "", // AWS SDK data sources don't have struct types
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "SDKDataSources",
		SDKType:            "aws_sdk",
		SchemaIndex:        schemaIndex,
		ReadIndex:          readIndex,
		AttributeIndex:     attributeIndex,
	}
}

// NewTerraformDataSourceFromAWSFramework creates a TerraformDataSource struct from AWS Framework data source info
func NewTerraformDataSourceFromAWSFramework(awsDataSource AWSResourceInfo, serviceReg ServiceRegistration) TerraformDataSource {
	// Framework data sources use the actual struct type extracted from the factory function
	structType := awsDataSource.StructType

	return TerraformDataSource{
		TerraformType:      awsDataSource.TerraformType,
		StructType:         structType, // Framework data sources use struct types
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "FrameworkDataSources",
		SDKType:            "aws_framework",
		// Framework data sources use method-based indexes on struct types
		SchemaIndex:    fmt.Sprintf("method.%s.Schema.goindex", structType),
		ReadIndex:      fmt.Sprintf("method.%s.Read.goindex", structType),
		AttributeIndex: fmt.Sprintf("method.%s.Schema.goindex", structType),
	}
}

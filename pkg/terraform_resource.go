package pkg

import "fmt"

// TerraformResource represents information about a Terraform resource
type TerraformResource struct {
	TerraformType      string `json:"terraform_type"` // "aws_vpc"
	StructType         string `json:"struct_type"`
	Namespace          string `json:"namespace"`           // "github.com/hashicorp/terraform-provider-aws/internal/service/resource/ec2"
	RegistrationMethod string `json:"registration_method"` // "SupportedResources", "Resources", etc.
	SDKType            string `json:"sdk_type"`            // "legacy_pluginsdk", "modern_sdk"
	SchemaIndex        string `json:"schema_index,omitempty"`
	CreateIndex        string `json:"create_index,omitempty"`
	ReadIndex          string `json:"read_index,omitempty"`
	UpdateIndex        string `json:"update_index,omitempty"`
	DeleteIndex        string `json:"delete_index,omitempty"`
	AttributeIndex     string `json:"attribute_index,omitempty"`
}

func NewTerraformResourceInfo(terraformType, structType, registrationMethod, sdkType string, serviceReg ServiceRegistration) TerraformResource {
	if sdkType == "legacy_pluginsdk" {
		result := TerraformResource{
			TerraformType:      terraformType,
			StructType:         "",
			Namespace:          serviceReg.PackagePath,
			RegistrationMethod: registrationMethod,
			SDKType:            sdkType,
			// Optional fields can be added later when we have more sophisticated AST parsing
			SchemaIndex:    fmt.Sprintf("func.%s.goindex", registrationMethod),
			CreateIndex:    "",
			ReadIndex:      "",
			UpdateIndex:    "",
			DeleteIndex:    "",
			AttributeIndex: fmt.Sprintf("func.%s.goindex", registrationMethod),
		}
		// Add CRUD methods if available
		if crudMethods, exists := serviceReg.ResourceCRUDMethods[terraformType]; exists && crudMethods != nil {
			result.CreateIndex = fmt.Sprintf("func.%s.goindex", crudMethods.CreateMethod)
			result.ReadIndex = fmt.Sprintf("func.%s.goindex", crudMethods.ReadMethod)
			result.UpdateIndex = fmt.Sprintf("func.%s.goindex", crudMethods.UpdateMethod)
			result.DeleteIndex = fmt.Sprintf("func.%s.goindex", crudMethods.DeleteMethod)
		}
		return result
	}
	return TerraformResource{
		TerraformType:      serviceReg.ResourceTerraformTypes[structType],
		StructType:         structType,
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "",
		SDKType:            sdkType,
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaIndex:    fmt.Sprintf("method.%s.Arguments.goindex", structType),
		CreateIndex:    fmt.Sprintf("method.%s.Create.goindex", structType),
		ReadIndex:      fmt.Sprintf("method.%s.Read.goindex", structType),
		UpdateIndex:    fmt.Sprintf("method.%s.Update.goindex", structType),
		DeleteIndex:    fmt.Sprintf("method.%s.Delete.goindex", structType),
		AttributeIndex: fmt.Sprintf("method.%s.Attributes.goindex", structType),
	}
}

// NewTerraformResourceFromAWSSDK creates a TerraformResource from AWS SDK resource info
func NewTerraformResourceFromAWSSDK(awsResource AWSResourceInfo, serviceReg ServiceRegistration) TerraformResource {
	result := TerraformResource{
		TerraformType:      awsResource.TerraformType,
		StructType:         "", // AWS SDK resources don't have struct types
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "SDKResources",
		SDKType:            "aws_sdk",
		// Schema and Attribute indexes always use the factory function
		SchemaIndex:    fmt.Sprintf("func.%s.goindex", awsResource.FactoryFunction),
		AttributeIndex: fmt.Sprintf("func.%s.goindex", awsResource.FactoryFunction),
	}

	// Use extracted CRUD methods if available (same pattern as legacy plugin SDK resources)
	if crudMethods, exists := serviceReg.ResourceCRUDMethods[awsResource.TerraformType]; exists && crudMethods != nil {
		if crudMethods.CreateMethod != "" {
			result.CreateIndex = fmt.Sprintf("func.%s.goindex", crudMethods.CreateMethod)
		}
		if crudMethods.ReadMethod != "" {
			result.ReadIndex = fmt.Sprintf("func.%s.goindex", crudMethods.ReadMethod)
		}
		if crudMethods.UpdateMethod != "" {
			result.UpdateIndex = fmt.Sprintf("func.%s.goindex", crudMethods.UpdateMethod)
		}
		if crudMethods.DeleteMethod != "" {
			result.DeleteIndex = fmt.Sprintf("func.%s.goindex", crudMethods.DeleteMethod)
		}
	}

	return result
}

// NewTerraformResourceFromAWSFramework creates a TerraformResource from AWS Framework resource info
func NewTerraformResourceFromAWSFramework(awsResource AWSResourceInfo, serviceReg ServiceRegistration) TerraformResource {
	// Framework resources use the actual struct type extracted from the factory function
	structType := awsResource.StructType
	
	result := TerraformResource{
		TerraformType:      awsResource.TerraformType,
		StructType:         structType, // Framework resources use struct types
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "FrameworkResources",
		SDKType:            "aws_framework",
		// Framework resources use method-based indexes on struct types
		SchemaIndex:    fmt.Sprintf("method.%s.Schema.goindex", structType),
		AttributeIndex: fmt.Sprintf("method.%s.Schema.goindex", structType),
	}

	// Framework resources use struct-based methods for CRUD operations
	if crudMethods, exists := serviceReg.ResourceCRUDMethods[awsResource.TerraformType]; exists && crudMethods != nil {
		if crudMethods.CreateMethod != "" {
			result.CreateIndex = fmt.Sprintf("method.%s.Create.goindex", structType)
		}
		if crudMethods.ReadMethod != "" {
			result.ReadIndex = fmt.Sprintf("method.%s.Read.goindex", structType)
		}
		if crudMethods.UpdateMethod != "" {
			result.UpdateIndex = fmt.Sprintf("method.%s.Update.goindex", structType)
		}
		if crudMethods.DeleteMethod != "" {
			result.DeleteIndex = fmt.Sprintf("method.%s.Delete.goindex", structType)
		}
	}

	return result
}

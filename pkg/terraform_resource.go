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
	return TerraformResource{
		TerraformType:      awsResource.TerraformType,
		StructType:         "", // AWS SDK resources don't have struct types
		Namespace:          serviceReg.PackagePath,
		RegistrationMethod: "SDKResources",
		SDKType:            "aws_sdk",
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaIndex:    fmt.Sprintf("func.%s.goindex", awsResource.FactoryFunction),
		CreateIndex:    fmt.Sprintf("func.%s.create.goindex", awsResource.FactoryFunction),
		ReadIndex:      fmt.Sprintf("func.%s.read.goindex", awsResource.FactoryFunction),
		UpdateIndex:    fmt.Sprintf("func.%s.update.goindex", awsResource.FactoryFunction),
		DeleteIndex:    fmt.Sprintf("func.%s.delete.goindex", awsResource.FactoryFunction),
		AttributeIndex: fmt.Sprintf("func.%s.goindex", awsResource.FactoryFunction),
	}
}

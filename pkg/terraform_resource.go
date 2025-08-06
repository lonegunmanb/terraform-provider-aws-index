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

// NewTerraformResourceFromAWSSDK creates a TerraformResource from AWS SDK resource info
func NewTerraformResourceFromAWSSDK(awsResource AWSResource, serviceReg ServiceRegistration) TerraformResource {
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
func NewTerraformResourceFromAWSFramework(awsResource AWSResource, serviceReg ServiceRegistration) TerraformResource {
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

	result.CreateIndex = fmt.Sprintf("method.%s.Create.goindex", structType)
	result.ReadIndex = fmt.Sprintf("method.%s.Read.goindex", structType)
	result.UpdateIndex = fmt.Sprintf("method.%s.Update.goindex", structType)
	result.DeleteIndex = fmt.Sprintf("method.%s.Delete.goindex", structType)

	return result
}

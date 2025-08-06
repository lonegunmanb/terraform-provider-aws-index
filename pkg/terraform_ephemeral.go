package pkg

import "fmt"

// TerraformEphemeral represents information about a Terraform ephemeral resource
type TerraformEphemeral struct {
	TerraformType      string `json:"terraform_type"` // "aws_secretsmanager_secret_version"
	StructType         string `json:"struct_type"`
	Namespace          string `json:"namespace"`
	RegistrationMethod string `json:"registration_method"`
	SDKType            string `json:"sdk_type"`
	SchemaIndex        string `json:"schema_index,omitempty"`
	OpenIndex          string `json:"open_index,omitempty"`
	RenewIndex         string `json:"renew_index,omitempty"`
	CloseIndex         string `json:"close_index,omitempty"`
}

// NewTerraformEphemeralInfo creates a TerraformEphemeral struct (legacy approach)
func NewTerraformEphemeralInfo(structType string, service ServiceRegistration) TerraformEphemeral {
	return TerraformEphemeral{
		TerraformType:      service.EphemeralTerraformTypes[structType],
		StructType:         structType,
		Namespace:          service.PackagePath,
		RegistrationMethod: "EphemeralResources",
		SDKType:            "ephemeral",
		// Optional fields can be added later when we have more sophisticated AST parsing
		SchemaIndex: fmt.Sprintf("method.%s.Schema.goindex", structType),
		OpenIndex:   fmt.Sprintf("method.%s.Open.goindex", structType),
		RenewIndex:  fmt.Sprintf("method.%s.Renew.goindex", structType),
		CloseIndex:  fmt.Sprintf("method.%s.Close.goindex", structType),
	}
}

// NewTerraformEphemeralFromAWS creates a TerraformEphemeral struct from AWS ephemeral resource information
func NewTerraformEphemeralFromAWS(awsEphemeral AWSResource, service ServiceRegistration) TerraformEphemeral {
	ephemeral := TerraformEphemeral{
		TerraformType:      awsEphemeral.TerraformType,
		StructType:         awsEphemeral.StructType,
		Namespace:          service.PackagePath,
		RegistrationMethod: awsEphemeral.FactoryFunction,
		SDKType:            awsEphemeral.SDKType,
	}

	// Set lifecycle method indexes if we have struct type (for method resolution)
	if awsEphemeral.StructType != "" {
		ephemeral.SchemaIndex = fmt.Sprintf("method.%s.Schema.goindex", awsEphemeral.StructType)
		ephemeral.OpenIndex = fmt.Sprintf("method.%s.Open.goindex", awsEphemeral.StructType)
		ephemeral.RenewIndex = fmt.Sprintf("method.%s.Renew.goindex", awsEphemeral.StructType)
		ephemeral.CloseIndex = fmt.Sprintf("method.%s.Close.goindex", awsEphemeral.StructType)
	}

	return ephemeral
}

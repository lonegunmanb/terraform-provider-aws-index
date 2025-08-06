package pkg

// AWSResource represents information about an AWS resource extracted from service package
type AWSResource struct {
	TerraformType   string `json:"terraform_type"`
	FactoryFunction string `json:"factory_function"`
	Name            string `json:"name"`
	SDKType         string `json:"sdk_type"`              // "sdk", "framework", "ephemeral"
	StructType      string `json:"struct_type,omitempty"` // For framework resources: "customModelsDataSource"
}

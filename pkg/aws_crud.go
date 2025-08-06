package pkg

// AWSCRUDMethods represents CRUD methods extracted from AWS factory functions
type AWSCRUDMethods struct {
	CreateMethod string `json:"create_method,omitempty"` // "resourceBucketCreate"
	ReadMethod   string `json:"read_method,omitempty"`   // "resourceBucketRead"
	UpdateMethod string `json:"update_method,omitempty"` // "resourceBucketUpdate"
	DeleteMethod string `json:"delete_method,omitempty"` // "resourceBucketDelete"

	// Framework resource methods (for struct-based resources)
	SchemaMethod string `json:"schema_method,omitempty"` // "Schema"

	// Ephemeral resource specific methods
	OpenMethod  string `json:"open_method,omitempty"`  // "Open"
	RenewMethod string `json:"renew_method,omitempty"` // "Renew"
	CloseMethod string `json:"close_method,omitempty"` // "Close"
}

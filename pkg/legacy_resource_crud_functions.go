package pkg

// LegacyResourceCRUDFunctions represents CRUD methods extracted from legacy plugin SDK resources
type LegacyResourceCRUDFunctions struct {
	CreateMethod string `json:"create_method,omitempty"` // "keyVaultCreateFunc"
	ReadMethod   string `json:"read_method,omitempty"`   // "keyVaultReadFunc"
	UpdateMethod string `json:"update_method,omitempty"` // "keyVaultUpdateFunc"
	DeleteMethod string `json:"delete_method,omitempty"` // "keyVaultDeleteFunc"
}

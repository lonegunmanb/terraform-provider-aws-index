package pkg

// ProviderStatistics represents summary statistics for the provider
type ProviderStatistics struct {
	ServiceCount       int `json:"service_count"`
	TotalDataSources   int `json:"total_data_sources"`
	TotalResources     int `json:"total_resources"`
	LegacyResources    int `json:"legacy_resources"`
	ModernResources    int `json:"modern_resources"`
	EphemeralResources int `json:"ephemeral_resources"`
}

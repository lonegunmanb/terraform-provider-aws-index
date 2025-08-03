package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestScanTerraformProviderServices(t *testing.T) {
	// Use test harness instead of actual terraform provider source
	testHarnessPath := filepath.Join("testharness", "internal", "services")

	// Scan the test harness services with no progress callback for testing
	index, err := ScanTerraformProviderServices(testHarnessPath, "github.com/lonegunmanb/terraform-provider-aws-index", "test-version", nil)
	require.NoError(t, err)
	require.NotNil(t, index)
	for _, svc := range index.Services {
		assert.Equal(t, fmt.Sprintf("github.com/lonegunmanb/terraform-provider-aws-index/testharness/internal/services/%s", svc.ServiceName), svc.PackagePath)
	}

	// Basic validation
	assert.Equal(t, "test-version", index.Version)
	assert.NotEmpty(t, index.Services, "Should find at least some services")
	assert.Greater(t, index.Statistics.ServiceCount, 0, "Should have at least one service")

	// Should find our test services
	serviceNames := make(map[string]bool)
	for _, service := range index.Services {
		serviceNames[service.ServiceName] = true
	}
	assert.Contains(t, serviceNames, "keyvault", "Should find keyvault service")
	assert.Contains(t, serviceNames, "resource", "Should find resource service")
	assert.Contains(t, serviceNames, "compute", "Should find compute service")
	assert.Contains(t, serviceNames, "storage", "Should find storage service")

	// Validate keyvault service specifically (has all registration types)
	var keyvaultService *ServiceRegistration
	for _, service := range index.Services {
		if service.ServiceName == "keyvault" {
			keyvaultService = &service
			break
		}
	}
	require.NotNil(t, keyvaultService, "Should find keyvault service")

	// Keyvault should have all registration types
	assert.NotEmpty(t, keyvaultService.SupportedResources, "Keyvault should have legacy resources")
	assert.NotEmpty(t, keyvaultService.SupportedDataSources, "Keyvault should have legacy data sources")
	assert.NotEmpty(t, keyvaultService.Resources, "Keyvault should have modern resources")
	assert.NotEmpty(t, keyvaultService.DataSources, "Keyvault should have modern data sources")
	assert.NotEmpty(t, keyvaultService.EphemeralFunctions, "Keyvault should have ephemeral resources")

	expectedSupportedResources := map[string]string{
		"azurerm_key_vault":               "resourceKeyVault",
		"azurerm_key_vault_certificate":   "resourceKeyVaultCertificate",
		"azurerm_key_vault_access_policy": "resourceKeyVaultAccessPolicy",
	}
	expectedSupportedDataSources := map[string]string{
		"azurerm_key_vault":             "dataSourceKeyVault",
		"azurerm_key_vault_certificate": "dataSourceKeyVaultCertificate",
		"azurerm_key_vault_secret":      "dataSourceKeyVaultSecret",
	}
	// Check specific expected values
	assert.Equal(t, expectedSupportedResources, keyvaultService.SupportedResources)
	assert.Equal(t, expectedSupportedDataSources, keyvaultService.SupportedDataSources)
	assert.Contains(t, keyvaultService.Resources, "KeyVaultCertificateContactsResource")
	assert.Contains(t, keyvaultService.DataSources, "EncryptedValueDataSource")
	assert.Contains(t, keyvaultService.EphemeralFunctions, "NewKeyVaultCertificateEphemeralResource")
	assert.Contains(t, keyvaultService.EphemeralFunctions, "NewKeyVaultSecretEphemeralResource")

	// Validate statistics make sense
	assert.Greater(t, index.Statistics.TotalResources, 0)
	assert.Greater(t, index.Statistics.TotalDataSources, 0)
	assert.Greater(t, index.Statistics.LegacyResources, 0)
	assert.Greater(t, index.Statistics.ModernResources, 0)
	assert.Greater(t, index.Statistics.EphemeralResources, 0)
}

func TestScanTerraformProviderServicesWithProgress(t *testing.T) {
	// Use test harness
	testHarnessPath := filepath.Join("testharness", "internal", "services")

	// Track progress updates
	var progressUpdates []ProgressInfo
	progressCallback := func(info ProgressInfo) {
		progressUpdates = append(progressUpdates, info)
	}

	// Scan with progress callback
	index, err := ScanTerraformProviderServices(testHarnessPath, "github.com/lonegunmanb/terraform-provider-aws-index", "test-version", progressCallback)
	require.NoError(t, err)
	require.NotNil(t, index)

	// Validate progress updates
	assert.NotEmpty(t, progressUpdates, "Should have received progress updates")

	// Should have at least initial and completion updates
	assert.GreaterOrEqual(t, len(progressUpdates), 2, "Should have at least initial and completion updates")

	// First update should be 0%
	firstUpdate := progressUpdates[0]
	assert.Equal(t, "scanning", firstUpdate.Phase)
	assert.Equal(t, 0.0, firstUpdate.Percentage)

	// Last update should be 100%
	lastUpdate := progressUpdates[len(progressUpdates)-1]
	assert.Equal(t, "scanning", lastUpdate.Phase)
	assert.Equal(t, 100.0, lastUpdate.Percentage)
	assert.Equal(t, "Completed", lastUpdate.Current)
}

func TestWriteIndexFilesWithProgress(t *testing.T) {
	// Use test harness
	testHarnessPath := filepath.Join("testharness", "internal", "services")

	// Scan first
	index, err := ScanTerraformProviderServices(testHarnessPath, "github.com/lonegunmanb/terraform-provider-aws-index", "test-version", nil)
	require.NoError(t, err)
	require.NotNil(t, index)

	// Track progress updates for writing
	var progressUpdates []ProgressInfo
	progressCallback := func(info ProgressInfo) {
		progressUpdates = append(progressUpdates, info)
	}

	// Create temp directory for output
	tempDir := t.TempDir()

	// Write with progress callback
	err = index.WriteIndexFiles(tempDir, progressCallback)
	require.NoError(t, err)

	// Validate progress updates
	assert.NotEmpty(t, progressUpdates, "Should have received progress updates")

	// Should have initial and completion updates
	assert.GreaterOrEqual(t, len(progressUpdates), 2, "Should have at least initial and completion updates")

	// All updates should be indexing phase
	for _, update := range progressUpdates {
		assert.Equal(t, "indexing", update.Phase)
	}

	// Last update should be 100%
	lastUpdate := progressUpdates[len(progressUpdates)-1]
	assert.Equal(t, 100.0, lastUpdate.Percentage)
	assert.Equal(t, "Completed", lastUpdate.Current)
}

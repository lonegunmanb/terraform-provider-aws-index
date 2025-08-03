package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseSource(source string) (*ast.File, error) {
	fset := token.NewFileSet()
	return parser.ParseFile(fset, "", source, parser.ParseComments)
}

func TestExtractSupportedResourcesMappings(t *testing.T) {
	// Test case with the exact example provided
	source := `package resource

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_management_lock":                      resourceManagementLock(),
		"azurerm_management_group_template_deployment": managementGroupTemplateDeploymentResource(),
		"azurerm_resource_group":                       resourceResourceGroup(),
		"azurerm_resource_group_template_deployment":   resourceGroupTemplateDeploymentResource(),
		"azurerm_subscription_template_deployment":     subscriptionTemplateDeploymentResource(),
		"azurerm_tenant_template_deployment":           tenantTemplateDeploymentResource(),
	}

	return resources
}`

	expected := map[string]string{
		"azurerm_management_lock":                      "resourceManagementLock",
		"azurerm_management_group_template_deployment": "managementGroupTemplateDeploymentResource",
		"azurerm_resource_group":                       "resourceResourceGroup",
		"azurerm_resource_group_template_deployment":   "resourceGroupTemplateDeploymentResource",
		"azurerm_subscription_template_deployment":     "subscriptionTemplateDeploymentResource",
		"azurerm_tenant_template_deployment":           "tenantTemplateDeploymentResource",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractSupportedResourcesMappings(node)
	assert.Equal(t, expected, result)
}

func TestExtractSupportedResourcesDirectReturn(t *testing.T) {
	// Test case with direct map return (no intermediate variable)
	source := `package resource

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault":        resourceKeyVault(),
		"azurerm_key_vault_secret": resourceKeyVaultSecret(),
	}
}`

	expected := map[string]string{
		"azurerm_key_vault":        "resourceKeyVault",
		"azurerm_key_vault_secret": "resourceKeyVaultSecret",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractSupportedResourcesMappings(node)

	assert.Equal(t, expected, result)
}

func TestExtractSupportedResourcesEmptyMethod(t *testing.T) {
	// Test case with empty method
	source := `package resource

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{}
}`

	node, err := parseSource(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	result := extractSupportedResourcesMappings(node)

	assert.Empty(t, result)
}

func TestExtractSupportedResourcesNoMethod(t *testing.T) {
	// Test case with no SupportedResources method
	source := `package resource

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault": dataSourceKeyVault(),
	}
}`

	node, err := parseSource(source)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	result := extractSupportedResourcesMappings(node)

	assert.Empty(t, result)
}

func TestExtractSupportedDataSourcesMappings(t *testing.T) {
	// Test case based on the actual keyvault service example
	source := `package keyvault

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault_access_policy":      dataSourceKeyVaultAccessPolicy(),
		"azurerm_key_vault_certificate":        dataSourceKeyVaultCertificate(),
		"azurerm_key_vault_certificate_data":   dataSourceKeyVaultCertificateData(),
		"azurerm_key_vault_certificate_issuer": dataSourceKeyVaultCertificateIssuer(),
		"azurerm_key_vault_key":                dataSourceKeyVaultKey(),
		"azurerm_key_vault_secret":             dataSourceKeyVaultSecret(),
		"azurerm_key_vault_secrets":            dataSourceKeyVaultSecrets(),
		"azurerm_key_vault":                    dataSourceKeyVault(),
		"azurerm_key_vault_certificates":       dataSourceKeyVaultCertificates(),
	}
}`

	expected := map[string]string{
		"azurerm_key_vault_access_policy":      "dataSourceKeyVaultAccessPolicy",
		"azurerm_key_vault_certificate":        "dataSourceKeyVaultCertificate",
		"azurerm_key_vault_certificate_data":   "dataSourceKeyVaultCertificateData",
		"azurerm_key_vault_certificate_issuer": "dataSourceKeyVaultCertificateIssuer",
		"azurerm_key_vault_key":                "dataSourceKeyVaultKey",
		"azurerm_key_vault_secret":             "dataSourceKeyVaultSecret",
		"azurerm_key_vault_secrets":            "dataSourceKeyVaultSecrets",
		"azurerm_key_vault":                    "dataSourceKeyVault",
		"azurerm_key_vault_certificates":       "dataSourceKeyVaultCertificates",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractSupportedDataSourcesMappings(node)
	assert.Equal(t, expected, result)
}

func TestExtractSupportedDataSourcesWithVariable(t *testing.T) {
	// Test case with intermediate variable (like SupportedResources pattern)
	source := `package resource

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	dataSources := map[string]*pluginsdk.Resource{
		"azurerm_client_config":    dataSourceArmClientConfig(),
		"azurerm_resource_group":   dataSourceArmResourceGroup(),
		"azurerm_subscription":     dataSourceArmSubscription(),
	}
	return dataSources
}`

	expected := map[string]string{
		"azurerm_client_config":  "dataSourceArmClientConfig",
		"azurerm_resource_group": "dataSourceArmResourceGroup",
		"azurerm_subscription":   "dataSourceArmSubscription",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractSupportedDataSourcesMappings(node)
	assert.Equal(t, expected, result)
}

func TestExtractSupportedDataSourcesEmpty(t *testing.T) {
	// Test case with empty SupportedDataSources method
	source := `package resource

func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractSupportedDataSourcesMappings(node)
	assert.Empty(t, result)
}

func TestExtractSupportedDataSourcesNoMethod(t *testing.T) {
	// Test case with no SupportedDataSources method
	source := `package resource

func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_resource_group": resourceResourceGroup(),
	}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractSupportedDataSourcesMappings(node)
	assert.Empty(t, result)
}

func TestExtractDataSourcesStructTypes(t *testing.T) {
	// Test case based on the actual keyvault service example
	source := `package keyvault

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		EncryptedValueDataSource{},
	}
}`

	expected := []string{"EncryptedValueDataSource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractDataSourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractDataSourcesStructTypesMultiple(t *testing.T) {
	// Test case with multiple struct types
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		FirstDataSource{},
		SecondDataSource{},
		ThirdDataSource{},
	}
}`

	expected := []string{"FirstDataSource", "SecondDataSource", "ThirdDataSource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractDataSourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractDataSourcesStructTypesWithVariable(t *testing.T) {
	// Test case with intermediate variable
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	dataSources := []sdk.DataSource{
		ConfigDataSource{},
		InfoDataSource{},
	}
	return dataSources
}`

	expected := []string{"ConfigDataSource", "InfoDataSource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractDataSourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractDataSourcesStructTypesEmpty(t *testing.T) {
	// Test case with empty DataSources method
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractDataSourcesStructTypes(node)
	assert.Empty(t, result)
}

func TestExtractDataSourcesStructTypesNoMethod(t *testing.T) {
	// Test case with no DataSources method
	source := `package service

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		SomeResource{},
	}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractDataSourcesStructTypes(node)
	assert.Empty(t, result)
}

func TestExtractResourcesStructTypes(t *testing.T) {
	// Test case based on the actual keyvault service example
	source := `package keyvault

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		KeyVaultCertificateContactsResource{},
	}
}`

	expected := []string{"KeyVaultCertificateContactsResource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractResourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractResourcesStructTypesMultiple(t *testing.T) {
	// Test case with multiple struct types
	source := `package service

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		FirstResource{},
		SecondResource{},
		ThirdResource{},
	}
}`

	expected := []string{"FirstResource", "SecondResource", "ThirdResource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractResourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractResourcesStructTypesWithVariable(t *testing.T) {
	// Test case with intermediate variable
	source := `package service

func (r Registration) Resources() []sdk.Resource {
	resources := []sdk.Resource{
		ConfigResource{},
		InfoResource{},
	}
	return resources
}`

	expected := []string{"ConfigResource", "InfoResource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractResourcesStructTypes(node)
	assert.Equal(t, expected, result)
}

func TestExtractResourcesStructTypesEmpty(t *testing.T) {
	// Test case with empty slice
	source := `package service

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractResourcesStructTypes(node)
	assert.Empty(t, result)
}

func TestExtractResourcesStructTypesNoMethod(t *testing.T) {
	// Test case with no Resources method
	source := `package service

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{
		SomeDataSource{},
	}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractResourcesStructTypes(node)
	assert.Empty(t, result)
}

func TestExtractEphemeralResourcesFunctions(t *testing.T) {
	// Test case based on the actual keyvault service example
	source := `package keyvault

func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewKeyVaultCertificateEphemeralResource,
	}
}`

	expected := []string{"NewKeyVaultCertificateEphemeralResource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractEphemeralResourcesFunctions(node)
	assert.Equal(t, expected, result)
}

func TestExtractEphemeralResourcesFunctionsMultiple(t *testing.T) {
	// Test case with multiple function names
	source := `package service

func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewFirstEphemeralResource,
		NewSecondEphemeralResource,
		NewThirdEphemeralResource,
	}
}`

	expected := []string{"NewFirstEphemeralResource", "NewSecondEphemeralResource", "NewThirdEphemeralResource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractEphemeralResourcesFunctions(node)
	assert.Equal(t, expected, result)
}

func TestExtractEphemeralResourcesFunctionsWithVariable(t *testing.T) {
	// Test case with intermediate variable
	source := `package service

func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource {
	ephemeralResources := []func() ephemeral.EphemeralResource{
		NewConfigEphemeralResource,
		NewInfoEphemeralResource,
	}
	return ephemeralResources
}`

	expected := []string{"NewConfigEphemeralResource", "NewInfoEphemeralResource"}

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractEphemeralResourcesFunctions(node)
	assert.Equal(t, expected, result)
}

func TestExtractEphemeralResourcesFunctionsEmpty(t *testing.T) {
	// Test case with empty slice
	source := `package service

func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractEphemeralResourcesFunctions(node)
	assert.Empty(t, result)
}

func TestExtractEphemeralResourcesFunctionsNoMethod(t *testing.T) {
	// Test case with no EphemeralResources method
	source := `package service

func (r Registration) Resources() []sdk.Resource {
	return []sdk.Resource{
		SomeResource{},
	}
}`

	node, err := parseSource(source)
	require.NoError(t, err)

	result := extractEphemeralResourcesFunctions(node)
	assert.Empty(t, result)
}

func TestExtractLegacyResourceCRUDMethods_DirectReturn(t *testing.T) {
	// Test case with direct return pattern
	source := `package keyvault

func resourceKeyVault() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		CreateContext: keyVaultCreateFunc,
		ReadContext:   keyVaultReadFunc,
		UpdateContext: keyVaultUpdateFunc,
		DeleteContext: keyVaultDeleteFunc,
		
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
		
		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
		},
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "keyVaultCreateFunc",
		ReadMethod:   "keyVaultReadFunc",
		UpdateMethod: "keyVaultUpdateFunc",
		DeleteMethod: "keyVaultDeleteFunc",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractLegacyResourceCRUDMethods_VariableAssignment(t *testing.T) {
	// Test case with variable assignment pattern
	source := `package storage

func resourceStorageAccount() *pluginsdk.Resource {
	resource := &pluginsdk.Resource{
		CreateFunc: storageAccountCreateFunc,
		ReadFunc:   storageAccountReadFunc,
		UpdateFunc: storageAccountUpdateFunc,
		DeleteFunc: storageAccountDeleteFunc,
		
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
	}
	return resource
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "storageAccountCreateFunc",
		ReadMethod:   "storageAccountReadFunc",
		UpdateMethod: "storageAccountUpdateFunc",
		DeleteMethod: "storageAccountDeleteFunc",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractLegacyResourceCRUDMethods_MixedContextAndFunc(t *testing.T) {
	// Test case with mixed Context and Func naming
	source := `package network

func resourceVirtualNetwork() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		CreateContext: vnetCreateFunc,
		ReadFunc:      vnetReadFunc,
		UpdateContext: vnetUpdateFunc,
		DeleteFunc:    vnetDeleteFunc,
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "vnetCreateFunc",
		ReadMethod:   "vnetReadFunc",
		UpdateMethod: "vnetUpdateFunc",
		DeleteMethod: "vnetDeleteFunc",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractLegacyResourceCRUDMethods_SelectorExpression(t *testing.T) {
	// Test case with selector expressions (package.Function)
	source := `package compute

func resourceVirtualMachine() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		CreateContext: compute.VMCreateFunc,
		ReadContext:   compute.VMReadFunc,
		UpdateContext: compute.VMUpdateFunc,
		DeleteContext: compute.VMDeleteFunc,
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "VMCreateFunc",
		ReadMethod:   "VMReadFunc",
		UpdateMethod: "VMUpdateFunc",
		DeleteMethod: "VMDeleteFunc",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractLegacyResourceCRUDMethods_PartialMethods(t *testing.T) {
	// Test case with only some CRUD methods defined
	source := `package example

func resourceReadOnlyResource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		CreateContext: resourceCreateFunc,
		ReadContext:   resourceReadFunc,
		// No Update or Delete methods
		
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "resourceCreateFunc",
		ReadMethod:   "resourceReadFunc",
		UpdateMethod: "", // Empty
		DeleteMethod: "", // Empty
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractLegacyResourceCRUDMethods_NoResourceFunction(t *testing.T) {
	// Test case with no pluginsdk.Resource function
	source := `package example

func someOtherFunction() string {
	return "not a resource"
}

func anotherFunction() *SomeOtherType {
	return &SomeOtherType{}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "",
		ReadMethod:   "",
		UpdateMethod: "",
		DeleteMethod: "",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractLegacyResourceCRUDMethods_EmptyResourceStruct(t *testing.T) {
	// Test case with empty pluginsdk.Resource struct
	source := `package example

func resourceEmptyResource() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		// No fields defined
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "",
		ReadMethod:   "",
		UpdateMethod: "",
		DeleteMethod: "",
	}

	node, err := parseSource(source)
	require.NoError(t, err)

	result, err := extractLegacyResourceCRUDMethods(node)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExtractTerraformTypeFromResourceTypeMethod(t *testing.T) {
	testCases := []struct {
		name           string
		src            string
		structName     string
		expectedResult string
	}{
		{
			name: "value receiver",
			src: `package test

type ContainerAppEnvironmentDaprComponentResource struct{}

func (r ContainerAppEnvironmentDaprComponentResource) ResourceType() string {
	return "azurerm_container_app_environment_dapr_component"
}`,
			structName:     "ContainerAppEnvironmentDaprComponentResource",
			expectedResult: "azurerm_container_app_environment_dapr_component",
		},
		{
			name: "pointer receiver",
			src: `package test

type KeyVaultResource struct{}

func (r *KeyVaultResource) ResourceType() string {
	return "azurerm_key_vault"
}`,
			structName:     "KeyVaultResource",
			expectedResult: "azurerm_key_vault",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the source code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tc.src, parser.ParseComments)
			assert.NoError(t, err)

			// Create mock package info
			packageInfo := &gophon.PackageInfo{
				Files: []*gophon.FileInfo{
					{
						File: file,
					},
				},
			}

			// Test the extraction
			result := extractTerraformTypeFromResourceTypeMethod(packageInfo, tc.structName)

			// Verify the result
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestExtractTerraformTypeFromDataSourceResourceTypeMethod(t *testing.T) {
	testCases := []struct {
		name           string
		src            string
		structName     string
		expectedResult string
	}{
		{
			name: "data source with value receiver",
			src: `package test

type KeyVaultDataSource struct{}

func (d KeyVaultDataSource) ResourceType() string {
	return "azurerm_key_vault"
}`,
			structName:     "KeyVaultDataSource",
			expectedResult: "azurerm_key_vault",
		},
		{
			name: "data source with pointer receiver",
			src: `package test

type ClientConfigDataSource struct{}

func (d *ClientConfigDataSource) ResourceType() string {
	return "azurerm_client_config"
}`,
			structName:     "ClientConfigDataSource",
			expectedResult: "azurerm_client_config",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the source code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tc.src, parser.ParseComments)
			assert.NoError(t, err)

			// Create mock package info
			packageInfo := &gophon.PackageInfo{
				Files: []*gophon.FileInfo{
					{
						File: file,
					},
				},
			}

			// Test the extraction (using the same function since data sources also use ResourceType method)
			result := extractTerraformTypeFromResourceTypeMethod(packageInfo, tc.structName)

			// Verify the result
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestExtractTerraformTypeFromMetadataMethod(t *testing.T) {
	testCases := []struct {
		name           string
		src            string
		structName     string
		expectedResult string
	}{
		{
			name: "ephemeral resource with value receiver",
			src: `package test

type KeyVaultSecretEphemeralResource struct{}

func (e KeyVaultSecretEphemeralResource) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "azurerm_key_vault_secret"
}`,
			structName:     "KeyVaultSecretEphemeralResource",
			expectedResult: "azurerm_key_vault_secret",
		},
		{
			name: "ephemeral resource with pointer receiver",
			src: `package test

type KeyVaultCertificateEphemeralResource struct{}

func (e *KeyVaultCertificateEphemeralResource) Metadata(_ context.Context, _ ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "azurerm_key_vault_certificate"
}`,
			structName:     "KeyVaultCertificateEphemeralResource",
			expectedResult: "azurerm_key_vault_certificate",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the source code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tc.src, parser.ParseComments)
			assert.NoError(t, err)

			// Create mock package info
			packageInfo := &gophon.PackageInfo{
				Files: []*gophon.FileInfo{
					{
						File: file,
					},
				},
			}

			// Test the extraction (using the metadata method extraction function)
			result := extractTerraformTypeFromMetadataMethod(packageInfo, tc.structName)

			// Verify the result
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

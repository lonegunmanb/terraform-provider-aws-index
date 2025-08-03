package keyvault

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	pluginsdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Dummy types for test harness
type Registration struct{}

// Dummy structs for modern SDK
type EncryptedValueDataSource struct{}

func (e EncryptedValueDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (e EncryptedValueDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (e EncryptedValueDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

type KeyVaultCertificateContactsResource struct{}

func (k KeyVaultCertificateContactsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (k KeyVaultCertificateContactsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (k KeyVaultCertificateContactsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (k KeyVaultCertificateContactsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (k KeyVaultCertificateContactsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (k KeyVaultCertificateContactsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault":               resourceKeyVault(),
		"azurerm_key_vault_certificate":   resourceKeyVaultCertificate(),
		"azurerm_key_vault_access_policy": resourceKeyVaultAccessPolicy(),
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	return map[string]*pluginsdk.Resource{
		"azurerm_key_vault":             dataSourceKeyVault(),
		"azurerm_key_vault_certificate": dataSourceKeyVaultCertificate(),
		"azurerm_key_vault_secret":      dataSourceKeyVaultSecret(),
	}
}

// DataSources returns a list of Data Sources supported by this Service
func (r Registration) DataSources() []datasource.DataSource {
	return []datasource.DataSource{
		EncryptedValueDataSource{},
	}
}

// Resources returns a list of Resources supported by this Service
func (r Registration) Resources() []resource.Resource {
	return []resource.Resource{
		KeyVaultCertificateContactsResource{},
	}
}

// EphemeralResources returns a list of ephemeral Resources supported by this Service
func (r Registration) EphemeralResources() []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewKeyVaultCertificateEphemeralResource,
		NewKeyVaultSecretEphemeralResource,
	}
}

// Dummy functions for legacy SDK
func resourceKeyVault() *pluginsdk.Resource              { return nil }
func resourceKeyVaultCertificate() *pluginsdk.Resource   { return nil }
func resourceKeyVaultAccessPolicy() *pluginsdk.Resource  { return nil }
func dataSourceKeyVault() *pluginsdk.Resource            { return nil }
func dataSourceKeyVaultCertificate() *pluginsdk.Resource { return nil }
func dataSourceKeyVaultSecret() *pluginsdk.Resource      { return nil }

// Dummy functions for ephemeral resources
func NewKeyVaultCertificateEphemeralResource() ephemeral.EphemeralResource { return nil }
func NewKeyVaultSecretEphemeralResource() ephemeral.EphemeralResource      { return nil }

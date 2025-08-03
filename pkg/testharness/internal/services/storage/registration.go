package storage

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	pluginsdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Dummy types for test harness
type Registration struct{}

// Dummy structs for modern SDK
type AccountResource struct{}

func (a AccountResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AccountResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

type BlobDataSource struct{}

func (b BlobDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (b BlobDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (b BlobDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

// SupportedResources with variable assignment pattern
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_storage_account": resourceStorageAccount(),
		"azurerm_storage_blob":    resourceStorageBlob(),
	}
	return resources
}

// DataSources with variable assignment pattern
func (r Registration) DataSources() []datasource.DataSource {
	dataSources := []datasource.DataSource{
		BlobDataSource{},
	}
	return dataSources
}

// Resources with variable assignment pattern
func (r Registration) Resources() []resource.Resource {
	resources := []resource.Resource{
		AccountResource{},
	}
	return resources
}

// Dummy functions for legacy SDK
func resourceStorageAccount() *pluginsdk.Resource { return nil }
func resourceStorageBlob() *pluginsdk.Resource    { return nil }

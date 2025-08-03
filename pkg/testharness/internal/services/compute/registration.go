package compute

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Dummy types for test harness
type Registration struct{}

// Dummy structs for modern SDK
type VirtualMachineResource struct{}

func (v VirtualMachineResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

type VirtualMachineScaleSetResource struct{}

func (v VirtualMachineScaleSetResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineScaleSetResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineScaleSetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineScaleSetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineScaleSetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineScaleSetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

type AvailabilitySetDataSource struct{}

func (a AvailabilitySetDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AvailabilitySetDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (a AvailabilitySetDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

type VirtualMachineDataSource struct{}

func (v VirtualMachineDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (v VirtualMachineDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

// DataSources returns a list of Data Sources supported by this Service
func (r Registration) DataSources() []datasource.DataSource {
	return []datasource.DataSource{
		AvailabilitySetDataSource{},
		VirtualMachineDataSource{},
	}
}

// Resources returns a list of Resources supported by this Service
func (r Registration) Resources() []resource.Resource {
	return []resource.Resource{
		VirtualMachineResource{},
		VirtualMachineScaleSetResource{},
	}
}

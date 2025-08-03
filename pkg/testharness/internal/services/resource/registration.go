package resource

import (
	pluginsdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Dummy types for test harness
type Registration struct{}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*pluginsdk.Resource {
	resources := map[string]*pluginsdk.Resource{
		"azurerm_resource_group":                     resourceResourceGroup(),
		"azurerm_management_lock":                    resourceManagementLock(),
		"azurerm_resource_group_template_deployment": resourceGroupTemplateDeploymentResource(),
	}
	return resources
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*pluginsdk.Resource {
	dataSources := map[string]*pluginsdk.Resource{
		"azurerm_client_config":  dataSourceArmClientConfig(),
		"azurerm_resource_group": dataSourceArmResourceGroup(),
		"azurerm_subscription":   dataSourceArmSubscription(),
	}
	return dataSources
}

// Dummy functions for legacy SDK
func resourceResourceGroup() *pluginsdk.Resource                   { return nil }
func resourceManagementLock() *pluginsdk.Resource                  { return nil }
func resourceGroupTemplateDeploymentResource() *pluginsdk.Resource { return nil }
func dataSourceArmClientConfig() *pluginsdk.Resource               { return nil }
func dataSourceArmResourceGroup() *pluginsdk.Resource              { return nil }
func dataSourceArmSubscription() *pluginsdk.Resource               { return nil }

# Terraform Provider AWS Index

An automated indexing system that generates comprehensive indexes for the HashiCorp Terraform AzureRM provider, enabling AI agents, IDEs, and development tools to better understand and work with Terraform provider code.

## 🎯 Purpose

This repository automatically monitors the [`hashicorp/terraform-provider-azurerm`](https://github.com/hashicorp/terraform-provider-azurerm) repository for new releases and generates structured indexes containing:

- **Terraform Resources** (e.g., `azurerm_resource_group`, `azurerm_key_vault`)
- **Data Sources** (e.g., `azurerm_client_config`, `azurerm_subscription`)
- **Ephemeral Resources** (e.g., `azurerm_key_vault_certificate`)
- **Go Symbol Information** (functions, types, methods)
- **CRUD Method Mappings** (Create, Read, Update, Delete operations)

## 📁 Index File Organization

The generated indexes are organized in a structured directory layout:

```text
index/
├── terraform-provider-azurerm-index.json    # Master index with metadata
├── resources/                               # Individual resource mappings
│   ├── azurerm_resource_group.json
│   ├── azurerm_key_vault.json
│   ├── azurerm_virtual_machine.json
│   └── ... (1000+ resource files)
├── datasources/                             # Individual data source mappings
│   ├── azurerm_client_config.json
│   ├── azurerm_subscription.json
│   ├── azurerm_key_vault.json
│   └── ... (200+ data source files)
├── ephemeral/                               # Individual ephemeral resource mappings
│   ├── azurerm_key_vault_certificate.json
│   ├── azurerm_key_vault_secret.json
│   └── ... (ephemeral resource files)
└── internal/                                # Go symbol indexes (if enabled)
    ├── func.NewSomething.goindex
    ├── type.SomeType.goindex
    └── ... (Go function/type indexes)
```

### Index File Structure

Each resource/data source/ephemeral resource has its own JSON file containing:

#### Resource Example (`resources/azurerm_key_vault.json`)

```json
{
  "terraform_type": "azurerm_key_vault",
  "struct_type": "",
  "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
  "registration_method": "resourceKeyVault",
  "sdk_type": "legacy_pluginsdk",
  "schema_index": "func.resourceKeyVault.goindex",
  "create_index": "func.resourceKeyVaultCreate.goindex",
  "read_index": "func.resourceKeyVaultRead.goindex",
  "update_index": "func.resourceKeyVaultUpdate.goindex",
  "delete_index": "func.resourceKeyVaultDelete.goindex",
  "attribute_index": "func.resourceKeyVault.goindex"
}
```

#### Data Source Example (`datasources/azurerm_client_config.json`)

```json
{
  "terraform_type": "azurerm_client_config",
  "struct_type": "",
  "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/authorization",
  "registration_method": "dataSourceArmClientConfig",
  "sdk_type": "legacy_pluginsdk",
  "schema_index": "func.dataSourceArmClientConfig.goindex",
  "read_index": "func.dataSourceArmClientConfigRead.goindex",
  "attribute_index": "func.dataSourceArmClientConfig.goindex"
}
```

#### Ephemeral Resource Example (`ephemeral/azurerm_key_vault_certificate.json`)

```json
{
  "terraform_type": "azurerm_key_vault_certificate",
  "struct_type": "KeyVaultCertificateEphemeralResource",
  "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault",
  "registration_method": "EphemeralResources",
  "sdk_type": "ephemeral",
  "schema_index": "method.KeyVaultCertificateEphemeralResource.Schema.goindex",
  "open_index": "method.KeyVaultCertificateEphemeralResource.Open.goindex",
  "renew_index": "method.KeyVaultCertificateEphemeralResource.Renew.goindex",
  "close_index": "method.KeyVaultCertificateEphemeralResource.Close.goindex"
}
```

## 🚀 Usage Examples

### For AI Agents and Language Models

#### 1. Finding Resource Implementation Details

```bash
# Get information about azurerm_key_vault resource
curl https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/main/index/resources/azurerm_key_vault.json
```

#### 2. Discovering Available Resources

```bash
# List all available resources
curl https://api.github.com/repos/lonegunmanb/terraform-provider-azurerm-index/contents/index/resources
```

#### 3. Finding CRUD Methods for Development

```bash
# Get CRUD method names for azurerm_resource_group
curl https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/main/index/resources/azurerm_resource_group.json | jq '.create_index, .read_index, .update_index, .delete_index'
```

### Supported Provider Versions

- **Latest Stable**: Always tracks the latest stable release (from `v4.25.0`)
- **Version History**: Tagged releases match the upstream provider versions
- **SDK Support**: Handles both Legacy Plugin SDK and Modern Terraform Plugin Framework

## 🛠️ Technical Architecture

### Multi-SDK Support

- **Legacy Plugin SDK**: Resources using `pluginsdk.Resource` structs
- **Modern Framework**: Resources using the newer Terraform Plugin Framework
- **Ephemeral Resources**: Temporary resources with Open/Renew/Close lifecycle

### Progress Tracking

Rich progress bars with:

- 🔄 Real-time progress indicators
- 📊 Completion percentages and item counts
- ⏱️ Elapsed time and ETA calculations
- ⚡ Processing rates (items/second)

## 📊 Statistics

Based on the latest Terraform Provider AzureRM version:

- **🏗️ Resources**: ~1,250 Terraform resources (e.g., `azurerm_resource_group`)
- **📖 Data Sources**: ~285 data sources (e.g., `azurerm_client_config`)
- **⚡ Ephemeral Resources**: ~15 ephemeral resources (e.g., `azurerm_key_vault_certificate`)
- **📦 Services**: 134 Azure service packages (e.g., `keyvault`, `compute`, `network`)
- **🔧 SDK Types**: Legacy Plugin SDK, Modern Framework, and Ephemeral support

## 🤝 Contributing

This repository is automatically maintained, but contributions are welcome:

1. **Bug Reports**: File issues for incorrect or missing index information
2. **Feature Requests**: Suggest improvements to the indexing system
3. **Tool Integration**: Share examples of how you're using these indexes

## 📄 License

This project is licensed under the same terms as the HashiCorp Terraform Provider AzureRM (Mozilla Public License 2.0).

## 🔗 Related Projects

- [HashiCorp Terraform Provider AzureRM](https://github.com/hashicorp/terraform-provider-azurerm) - The source provider being indexed
- [Terraform](https://terraform.io) - Infrastructure as Code tool
- [Gophon](https://github.com/lonegunmanb/gophon) - Go symbol indexing tool (if used for additional Go indexes)
- [`terraform-mcp-eva`](https://github.com/lonegunmanb/terraform-mcp-eva) - An experimental MCP serer that helps Terraform module developers to make their life easier.
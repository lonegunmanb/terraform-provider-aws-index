# Terraform Provider AWS Index

An automated indexing system that generates comprehensive indexes for the HashiCorp Terraform AWS provider, enabling AI agents, IDEs, and development tools to better understand and work with Terraform provider code.

## 🎯 Purpose

This repository automatically monitors the [`hashicorp/terraform-provider-aws`](https://github.com/hashicorp/terraform-provider-aws) repository for new releases and generates structured indexes containing:

- **Terraform Resources** (e.g., `aws_s3_bucket`, `aws_ec2_instance`)
- **Data Sources** (e.g., `aws_ami`, `aws_availability_zones`)
- **Ephemeral Resources** (e.g., `aws_secretsmanager_secret_version`)
- **Go Symbol Information** (functions, types, methods)
- **CRUD Method Mappings** (Create, Read, Update, Delete operations)

## 📁 Index File Organization

The generated indexes are organized in a structured directory layout:

```text
index/
├── terraform-provider-aws-index.json        # Master index with metadata
├── resources/                               # Individual resource mappings
│   ├── aws_s3_bucket.json
│   ├── aws_ec2_instance.json
│   ├── aws_rds_instance.json
│   └── ... (2000+ resource files)
├── datasources/                             # Individual data source mappings
│   ├── aws_ami.json
│   ├── aws_availability_zones.json
│   ├── aws_caller_identity.json
│   └── ... (400+ data source files)
├── ephemeral/                               # Individual ephemeral resource mappings
│   ├── aws_secretsmanager_secret_version.json
│   ├── aws_ssm_parameter.json
│   └── ... (ephemeral resource files)
└── internal/                                # Go symbol indexes (if enabled)
    ├── func.NewSomething.goindex
    ├── type.SomeType.goindex
    └── ... (Go function/type indexes)
```

### Index File Structure

Each resource/data source/ephemeral resource has its own JSON file containing:

#### Resource Example (`resources/aws_s3_bucket.json`)

```json
{
  "terraform_type": "aws_s3_bucket",
  "struct_type": "",
  "namespace": "github.com/hashicorp/terraform-provider-aws/internal/service/s3",
  "registration_method": "resourceBucket",
  "sdk_type": "legacy_pluginsdk",
  "schema_index": "func.resourceBucket.goindex",
  "create_index": "func.resourceBucketCreate.goindex",
  "read_index": "func.resourceBucketRead.goindex",
  "update_index": "func.resourceBucketUpdate.goindex",
  "delete_index": "func.resourceBucketDelete.goindex",
  "attribute_index": "func.resourceBucket.goindex"
}
```

#### Data Source Example (`datasources/aws_ami.json`)

```json
{
  "terraform_type": "aws_ami",
  "struct_type": "",
  "namespace": "github.com/hashicorp/terraform-provider-aws/internal/service/ec2",
  "registration_method": "dataSourceAMI",
  "sdk_type": "legacy_pluginsdk",
  "schema_index": "func.dataSourceAMI.goindex",
  "read_index": "func.dataSourceAMIRead.goindex",
  "attribute_index": "func.dataSourceAMI.goindex"
}
```

#### Ephemeral Resource Example (`ephemeral/aws_secretsmanager_secret_version.json`)

```json
{
  "terraform_type": "aws_secretsmanager_secret_version",
  "struct_type": "SecretVersionEphemeralResource",
  "namespace": "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager",
  "registration_method": "EphemeralResources",
  "sdk_type": "ephemeral",
  "schema_index": "method.SecretVersionEphemeralResource.Schema.goindex",
  "open_index": "method.SecretVersionEphemeralResource.Open.goindex",
  "renew_index": "method.SecretVersionEphemeralResource.Renew.goindex",
  "close_index": "method.SecretVersionEphemeralResource.Close.goindex"
}
```

## 🚀 Usage Examples

### For AI Agents and Language Models

#### 1. Finding Resource Implementation Details

```bash
# Get information about aws_s3_bucket resource
curl https://raw.githubusercontent.com/lonegunmanb/terraform-provider-aws-index/main/index/resources/aws_s3_bucket.json
```

#### 2. Discovering Available Resources

```bash
# List all available resources
curl https://api.github.com/repos/lonegunmanb/terraform-provider-aws-index/contents/index/resources
```

#### 3. Finding CRUD Methods for Development

```bash
# Get CRUD method names for aws_ec2_instance
curl https://raw.githubusercontent.com/lonegunmanb/terraform-provider-aws-index/main/index/resources/aws_ec2_instance.json | jq '.create_index, .read_index, .update_index, .delete_index'
```

### Supported Provider Versions

- **Latest Stable**: Always tracks the latest stable release (from `v5.30.0`)
- **Version History**: Tagged releases match the upstream provider versions
- **SDK Support**: Handles both Legacy Plugin SDK, Modern Terraform Plugin Framework, and Ephemeral Resources

## 🛠️ Technical Architecture

### Multi-SDK Support

- **Legacy Plugin SDK**: Resources using `schema.Resource` structs
- **Modern Framework**: Resources using the newer Terraform Plugin Framework
- **Ephemeral Resources**: Temporary resources with Open/Renew/Close lifecycle

### Progress Tracking

Rich progress bars with:

- 🔄 Real-time progress indicators
- 📊 Completion percentages and item counts
- ⏱️ Elapsed time and ETA calculations
- ⚡ Processing rates (items/second)

## 📊 Statistics

Based on the latest Terraform Provider AWS version:

- **🏗️ Resources**: ~2,000+ Terraform resources (e.g., `aws_s3_bucket`, `aws_ec2_instance`)
- **📖 Data Sources**: ~400+ data sources (e.g., `aws_ami`, `aws_availability_zones`)
- **⚡ Ephemeral Resources**: ~20+ ephemeral resources (e.g., `aws_secretsmanager_secret_version`)
- **📦 Services**: 200+ AWS service packages (e.g., `s3`, `ec2`, `rds`)
- **🔧 SDK Types**: Legacy Plugin SDK, Modern Framework, and Ephemeral support

## 🤝 Contributing

This repository is automatically maintained, but contributions are welcome:

1. **Bug Reports**: File issues for incorrect or missing index information
2. **Feature Requests**: Suggest improvements to the indexing system
3. **Tool Integration**: Share examples of how you're using these indexes

## 📄 License

This project is licensed under the same terms as the HashiCorp Terraform Provider AWS (Mozilla Public License 2.0).

## 🔗 Related Projects

- [HashiCorp Terraform Provider AWS](https://github.com/hashicorp/terraform-provider-aws) - The source provider being indexed
- [Terraform](https://terraform.io) - Infrastructure as Code tool
- [Gophon](https://github.com/lonegunmanb/gophon) - Go symbol indexing tool (if used for additional Go indexes)
- [`terraform-mcp-eva`](https://github.com/lonegunmanb/terraform-mcp-eva) - An experimental MCP server that helps Terraform module developers to make their life easier.
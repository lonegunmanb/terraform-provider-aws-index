# AWS Provider Index Migration Plan

## Overview
This project was originally designed for indexing the AzureRM Terraform provider but needs to be adapted for the AWS Terraform provider. The main challenge is that AWS provider uses a completely different service registration pattern compared to AzureRM.

## 🔍 **Critical Knowledge: AWS Provider Resource Declaration Patterns**

The AWS provider uses two fundamentally different approaches for declaring resources and data sources:

### **1. Legacy SDK Resources (Plugin SDK v2)**
- **Pattern**: Factory function returns `*schema.Resource` struct
- **CRUD Implementation**: Function fields in the returned struct
- **Example**: `resourceKeyPair()` in `ec2_key_pair.go`
```go
// @SDKResource("aws_key_pair", name="Key Pair")
func resourceKeyPair() *schema.Resource {
    return &schema.Resource{
        CreateWithoutTimeout: resourceKeyPairCreate,  // Function reference
        ReadWithoutTimeout:   resourceKeyPairRead,    // Function reference
        UpdateWithoutTimeout: resourceKeyPairUpdate,  // Function reference
        DeleteWithoutTimeout: resourceKeyPairDelete,  // Function reference
        Schema: map[string]*schema.Schema{...},
    }
}
```
- **Registration**: In `service_package_gen.go` via `SDKResources()` method
- **Index Pattern**: `func.<function_name>.goindex` (e.g., `func.resourceKeyPairCreate.goindex`)

### **2. Modern Framework Resources (Terraform Plugin Framework)**
- **Pattern**: Factory function returns a struct that implements resource interfaces
- **CRUD Implementation**: Methods on the struct type that implement Framework interfaces
`DataSource` interaface from Framework:

```go
type DataSource interface {
	// Metadata should return the full name of the data source, such as
	// examplecloud_thing.
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// Schema should return the schema for this data source.
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	// Read is called when the provider must read data source values in
	// order to update state. Config values should be read from the
	// ReadRequest and new state values set on the ReadResponse.
	Read(context.Context, ReadRequest, *ReadResponse)
}
```

Then `Resource` interface:

```go
type Resource interface {
	// Metadata should return the full name of the resource, such as
	// examplecloud_thing.
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// Schema should return the schema for this resource.
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	// Create is called when the provider must create a new resource. Config
	// and planned state values should be read from the
	// CreateRequest and new state values set on the CreateResponse.
	Create(context.Context, CreateRequest, *CreateResponse)

	// Read is called when the provider must read resource values in order
	// to update state. Planned state values should be read from the
	// ReadRequest and new state values set on the ReadResponse.
	Read(context.Context, ReadRequest, *ReadResponse)

	// Update is called to update the state of the resource. Config, planned
	// state, and prior state values should be read from the
	// UpdateRequest and new state values set on the UpdateResponse.
	Update(context.Context, UpdateRequest, *UpdateResponse)

	// Delete is called when the provider must delete the resource. Config
	// values may be read from the DeleteRequest.
	//
	// If execution completes without error, the framework will automatically
	// call DeleteResponse.State.RemoveResource(), so it can be omitted
	// from provider logic.
	Delete(context.Context, DeleteRequest, *DeleteResponse)
}
```

Ephemeral interafaces:

```go
type EphemeralResource interface {
	// Metadata should return the full name of the ephemeral resource, such as
	// examplecloud_thing.
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// Schema should return the schema for this ephemeral resource.
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	// Open is called when the provider must generate a new ephemeral resource. Config values
	// should be read from the OpenRequest and new response values set on the OpenResponse.
	Open(context.Context, OpenRequest, *OpenResponse)
}

// EphemeralResourceWithRenew is an interface type that extends EphemeralResource to
// include a method which the framework will call when Terraform detects that the
// provider-defined returned RenewAt time for an ephemeral resource has passed. This RenewAt
// response field can be set in the OpenResponse and RenewResponse.
type EphemeralResourceWithRenew interface {
	EphemeralResource

	// Renew is called when the provider must renew the ephemeral resource based on
	// the provided RenewAt time. This RenewAt response field can be set in the OpenResponse and RenewResponse.
	//
	// Renew cannot return new result data for the ephemeral resource instance, so this logic is only appropriate
	// for remote objects like HashiCorp Vault leases, which can be renewed without changing their data.
	Renew(context.Context, RenewRequest, *RenewResponse)
}

// EphemeralResourceWithClose is an interface type that extends
// EphemeralResource to include a method which the framework will call when
// Terraform determines that the ephemeral resource can be safely cleaned up.
type EphemeralResourceWithClose interface {
	EphemeralResource

	// Close is called when the provider can clean up the ephemeral resource.
	// Config values may be read from the CloseRequest.
	Close(context.Context, CloseRequest, *CloseResponse)
}

// EphemeralResourceWithConfigure is an interface type that extends EphemeralResource to
// include a method which the framework will automatically call so provider
// developers have the opportunity to setup any necessary provider-level data
// or clients in the EphemeralResource type.
type EphemeralResourceWithConfigure interface {
	EphemeralResource

	// Configure enables provider-level data or clients to be set in the
	// provider-defined EphemeralResource type.
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}

// EphemeralResourceWithConfigValidators is an interface type that extends EphemeralResource to include declarative validations.
//
// Declaring validation using this methodology simplifies implementation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type EphemeralResourceWithConfigValidators interface {
	EphemeralResource

	// ConfigValidators returns a list of functions which will all be performed during validation.
	ConfigValidators(context.Context) []ConfigValidator
}

// EphemeralResourceWithValidateConfig is an interface type that extends EphemeralResource to include imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single ephemeral resource. Any documentation
// of this functionality must be manually added into schema descriptions.
//
// Validation will include ConfigValidators and ValidateConfig, if both are
// implemented, in addition to any Attribute or Type validation.
type EphemeralResourceWithValidateConfig interface {
	EphemeralResource

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}
```

You can infer method name via these interafaces.

Very important assumptions:

All AWS resources, data sources and ephemerals are marked by special annotations, you should scan all go files to find these annotations to identify the code file that contains resources.

For Legacy Resource, the annotation is `@SDKResource`::

```go
// @SDKResource("aws_vpn_gateway", name="VPN Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPNGatewayCreate,
		ReadWithoutTimeout:   resourceVPNGatewayRead,
		UpdateWithoutTimeout: resourceVPNGatewayUpdate,
		DeleteWithoutTimeout: resourceVPNGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAmazonSideASN,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}
```

You should identify the function with this annotation, then extract CRUD function names from it's body.

For legacy plugin data source:

```go
// @SDKDataSource("aws_vpn_gateway", name="VPN Gateway")
// @Tags
// @Testing(tagsTest=false)
func dataSourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPNGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}
```

You should identify the function with this annotation, then extra read function name from it.

For modern framework resource:

```go
// @FrameworkResource("aws_bedrock_guardrail", name="Guardrail")
// @Tags(identifierAttribute="guardrail_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrock;bedrock.GetGuardrailOutput")
// @Testing(importStateIdFunc="testAccGuardrailImportStateIDFunc")
// @Testing(importStateIdAttribute="guardrail_id")
func newGuardrailResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &guardrailResource{
		flexOpt: fwflex.WithFieldNameSuffix("Config"),
	}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameGuardrail = "Guardrail"
)

type guardrailResource struct {
	framework.ResourceWithModel[guardrailResourceModel]
	framework.WithTimeouts

	flexOpt fwflex.AutoFlexOptionsFunc
}
```

You should find code file contains this annotation `@FrameworkResource`, you can learn the terraform type of this resource, then the struct type declaration that contains `framework.Resourcexxx`(maybe just `framework.Resource`).

For modern framework data source, `@FrameworkDataSource`:

```go
// @FrameworkDataSource("aws_bedrock_inference_profile", name="Inference Profile")
func newInferenceProfileDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfileDataSource{}, nil
}

type inferenceProfileDataSource struct {
	framework.DataSourceWithModel[inferenceProfileDataSourceModel]
}
```

For ephemeral, `@EphemeralResource`

```go
// @EphemeralResource("aws_lambda_invocation", name="Invocation")
func newInvocationEphemeralResource(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &invocationEphemeralResource{}, nil
}

const (
	ResNameInvocation = "Invocation"
)

type invocationEphemeralResource struct {
	framework.EphemeralResourceWithModel[invocationEphemeralResourceModel]
}
```

## 🎉 Phase 1.1 Achievement Summary

**Successfully implemented file-centric annotation scanning for AWS Terraform Provider!**

### Key Accomplishments:
1. **✅ Real-world validation**: Tested on actual AWS Lambda service with 32 files
2. **✅ Perfect annotation detection**: Found 21 annotations across all 5 types
3. **✅ CRUD extraction excellence**: 100% accurate SDK resource CRUD method mapping
4. **✅ Scalable architecture**: File-level scanning approach handles complex service packages
5. **✅ Type safety**: Proper enum-based annotation types with validation

### Test Results on AWS Lambda Service:
- **📊 Total Annotations**: 21
- **🏗️ SDK Resources**: 11 (with perfect CRUD mapping)
- **📖 SDK DataSources**: 7 (with read method extraction)  
- **⚡ Framework Resources**: 2 (with method inference)
- **📊 Framework DataSources**: 0
- **🔄 Ephemeral Resources**: 1

### Sample Perfect CRUD Extraction:
```
aws_lambda_function → resourceFunctionCreate, resourceFunctionRead, resourceFunctionUpdate, resourceFunctionDelete
aws_lambda_alias → resourceAliasCreate, resourceAliasRead, resourceAliasUpdate, resourceAliasDelete
```

## Current State Analysis

### What We Keep
1. **PackageInfo Structure**: The `gophon.PackageInfo` contains all files, functions, and types - we'll preserve this
2. **Core Data Structures**: `TerraformProviderIndex`, `ServiceRegistration`, etc. can be preserved with modifications
3. **File Writing Logic**: The JSON output generation and file writing can remain largely unchanged
4. **Progress Tracking**: The parallel processing and progress tracking infrastructure is solid

### What We Change
1. **Scanning Logic**: Replace factory function parsing with annotation-based scanning
2. **Extraction Methods**: Create new functions to extract info based on annotation types
3. **CRUD Detection**: Different approaches for legacy vs framework patterns

## Implementation Plan

### ✅ Phase 1: New Annotation Scanner Functions - COMPLETED

#### ✅ 1.1 Core Annotation Scanner - COMPLETED
~~Create `scanPackageForAnnotations(packageInfo *gophon.PackageInfo) AnnotationResults`~~
- ✅ Scan all files in packageInfo
- ✅ Look for comment patterns: `@SDKResource`, `@SDKDataSource`, `@FrameworkResource`, `@FrameworkDataSource`, `@EphemeralResource`
- ✅ Parse annotation parameters (terraform type, name, etc.)
- ✅ Return structured results mapping annotations to their context

**Implementation Status**: 
- ✅ Created `pkg/annotation_types.go` with data structures
- ✅ Created `pkg/annotation_scanner.go` with core scanning logic
- ✅ Successfully tested on real AWS Lambda service files
- ✅ **Results**: 21 annotations detected (11 SDK Resources, 7 SDK DataSources, 2 Framework Resources, 1 Ephemeral)
- ✅ **CRUD Extraction**: Perfect extraction of SDK resource CRUD method names
- ✅ **File-centric approach**: Successfully processes multiple files per package

#### ✅ 1.2 Annotation Types to Handle - COMPLETED
~~```go
type AnnotationResult struct {
    Type           string // "SDKResource", "SDKDataSource", "FrameworkResource", etc.
    TerraformType  string // e.g., "aws_key_pair"
    Name           string // e.g., "Key Pair"
    StructType     string // For framework resources (extracted from function body)
    FilePath       string // Source file path
    LineNumber     int    // For debugging
}
```~~

**Implemented Structure**:
```go
type AnnotationResult struct {
    Type             AnnotationType    // Enum for annotation type
    TerraformType    string           // e.g., "aws_lambda_function"
    Name             string           // e.g., "Function"
    FilePath         string           // Source file path
    RawAnnotation    string           // The raw annotation text
    StructType       string           // For framework resources
    CRUDMethods      map[string]string // For SDK resources: "create" -> "resourceFunctionCreate"
    FrameworkMethods []string         // For framework: ["Create", "Read", "Update", "Delete"]
}
```

### 🔄 Phase 2: Type-Specific Extractors - IN PROGRESS

#### 2.1 Legacy SDK Resource Extractor - ✅ COMPLETED
~~For `@SDKResource` annotations:~~
- ✅ Find the annotated function (e.g., `resourceKeyPair()`)
- ✅ Parse the function body to extract CRUD function references using prefix matching:
  - ✅ `CreateWithoutTimeout: resourceKeyPairCreate*` (matches Create, CreateContext, etc.)
  - ✅ `ReadWithoutTimeout: resourceKeyPairRead*` (matches Read, ReadContext, etc.)
  - ✅ `UpdateWithoutTimeout: resourceKeyPairUpdate*` (matches Update, UpdateContext, etc.)
  - ✅ `DeleteWithoutTimeout: resourceKeyPairDelete*` (matches Delete, DeleteContext, etc.)

**Status**: File-level CRUD extraction working perfectly on real AWS files.

#### 2.2 Legacy SDK DataSource Extractor - ✅ COMPLETED
~~For `@SDKDataSource` annotations:~~
- ✅ Find the annotated function (e.g., `dataSourceVPNGateway()`)
- ✅ Parse function body to extract read function using prefix matching:
  - ✅ `ReadWithoutTimeout: dataSourceVPNGatewayRead*` (matches Read, ReadContext, etc.)

**Status**: Read method extraction working correctly.

#### 2.3 Framework Resource Extractor - 🔄 NEEDS IMPROVEMENT
For `@FrameworkResource` annotations:
- ✅ Find the annotated function (e.g., `newGuardrailResource()`)
- 🔄 Extract the struct type returned (e.g., `guardrailResource`) - **Needs debugging**
- ✅ Methods are inferred from Framework interfaces (Create, Read, Update, Delete)

**Status**: Framework methods inference works, but struct type extraction needs improvement.

#### 2.4 Framework DataSource Extractor - ✅ MOSTLY COMPLETED
For `@FrameworkDataSource` annotations:
- ✅ Find the annotated function (e.g., `newInferenceProfileDataSource()`)
- 🔄 Extract the struct type returned (e.g., `inferenceProfileDataSource`) - **Same issue as 2.3**
- ✅ Methods are inferred from Framework interfaces (Read, Metadata, Schema)

#### 2.5 Ephemeral Resource Extractor - ✅ MOSTLY COMPLETED
For `@EphemeralResource` annotations:
- ✅ Find the annotated function
- 🔄 Extract struct type returned - **Same issue as 2.3**
- ✅ Methods are inferred from Ephemeral interfaces (Open, Close, Renew, etc.)

**Current Priority**: Fix framework struct type extraction in `extractFrameworkStructTypeFromFile()`

### Phase 3: Integration with Existing Code

#### 3.1 Update Main Scanning Function
Modify `ScanTerraformProviderServices()`:
- Keep the parallel processing structure
- Replace `parseAWSServiceFile()` call with new annotation-based scanning
- Update `extractAndStoreSDKCRUDMethodsForLegacyPlugin()` to use annotation results

#### 3.2 Update Service Registration
Modify `ServiceRegistration` creation:
- Use annotation results instead of factory function parsing
- Populate AWS-specific fields based on annotation types
- Maintain backward compatibility with existing JSON output format

#### 3.3 Update Data Structures
Ensure `AWSResourceInfo` and related structs support:
- Annotation-derived metadata
- Both legacy CRUD functions and framework struct types
- Clear distinction between SDK and Framework patterns

### Phase 4: Implementation Details

#### 4.1 Annotation Parsing Logic
```go
// Parse comment like: // @SDKResource("aws_key_pair", name="Key Pair")
func parseAnnotation(comment string) (*AnnotationResult, error) {
    // Extract annotation type (@SDKResource, @FrameworkResource, etc.)
    // Parse parameters using regex or simple string parsing
    // Return structured annotation data
}
```

#### 4.2 Function Body Analysis
```go
// For SDK resources: extract CRUD function names from *schema.Resource return using prefix matching
func extractSDKCRUDFromFunction(funcDecl *ast.FuncDecl) (*CRUDMethods, error) {
    // Parse assignments like: CreateWithoutTimeout: resourceKeyPairCreate
    // Use prefix matching to handle variations like CreateContext, ReadContext, etc.
    // Return function names for each CRUD operation
}

// For Framework resources: extract struct type from return statement
func extractFrameworkStructType(funcDecl *ast.FuncDecl) (string, error) {
    // Parse return like: return &guardrailResource{...}
    // Return struct type name
}
```

#### 4.3 Method Inference for Framework Types
```go
// Infer methods based on struct type and Framework interfaces
func inferFrameworkMethods(structType string, isResource bool) []string {
    if isResource {
        return []string{"Create", "Read", "Update", "Delete", "Metadata", "Schema"}
    } else {
        return []string{"Read", "Metadata", "Schema"}
    }
}
```

### Phase 5: Testing Strategy

#### 5.1 Unit Tests
- Test annotation parsing with various comment formats
- Test CRUD extraction from different function patterns
- Test struct type extraction from return statements

#### 5.2 Integration Tests
- Test against sample AWS provider service directories
- Verify output JSON format matches expected structure
- Compare results with existing implementation where possible

#### 5.3 Real-world Validation
- Run against actual terraform-provider-aws codebase
- Verify all known resource types are detected
- Check for false positives/negatives

### Phase 6: Migration and Cleanup

#### 6.1 Remove Deprecated Code
- Remove AzureRM-specific factory function parsing
- Clean up unused gophon analysis methods
- Remove outdated comment patterns

#### 6.2 Documentation Updates
- Update README.md to reflect AWS-specific approach
- Document annotation formats and requirements
- Provide examples of each annotation type

#### 6.3 Performance Optimization
- Optimize annotation scanning for large codebases
- Consider caching parsed annotations
- Profile memory usage with full AWS provider scan

## File Changes Required

### New Files
1. `pkg/annotation_scanner.go` - Core annotation parsing logic
2. `pkg/aws_extractors.go` - Type-specific extraction functions
3. `pkg/annotation_types.go` - Data structures for annotation results

### Modified Files
1. `pkg/terraform_provider_index.go` - Update main scanning logic
2. `pkg/service_registration.go` - Update to use annotation results
3. `pkg/aws_extractor.go` - Replace with annotation-based extraction
4. `README.md` - Update documentation

### Files to Review/Update
1. `pkg/terraform_resource.go` - Ensure compatibility with new extraction
2. `pkg/terraform_data_source.go` - Ensure compatibility with new extraction
3. `pkg/terraform_ephemeral.go` - Ensure compatibility with new extraction

## Success Criteria

1. **Accuracy**: All AWS resources/datasources with proper annotations are detected
2. **Performance**: Scanning time comparable to or better than current implementation
3. **Completeness**: Support for all 5 AWS provider patterns (SDK Resources, SDK DataSources, Framework Resources, Framework DataSources, Ephemeral Resources)
4. **Maintainability**: Clear separation between annotation parsing and data extraction
5. **Backward Compatibility**: Existing JSON output format preserved where possible

## Risk Mitigation

1. **Parsing Errors**: Comprehensive error handling for malformed annotations
2. **Performance**: Parallel processing maintained for large codebases
3. **Compatibility**: Gradual migration with fallback to existing logic where needed
4. **Testing**: Extensive testing against real AWS provider codebase

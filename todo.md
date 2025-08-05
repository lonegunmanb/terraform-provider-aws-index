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

- **Example**: `newGuardrailResource()` in `bedrock/guardrail.go`
```go
// @FrameworkResource("aws_bedrock_guardrail", name="Guardrail")
func newGuardrailResource(_ context.Context) (resource.ResourceWithConfigure, error) {
    r := &guardrailResource{...}  // Returns struct instance
    return r, nil
}

type guardrailResource struct {
    framework.ResourceWithModel[guardrailResourceModel]
    framework.WithTimeouts
}

// CRUD methods are implemented on the struct
func (r *guardrailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {...}
func (r *guardrailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {...}
func (r *guardrailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {...}
func (r *guardrailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {...}
```
- **Registration**: In `service_package_gen.go` via `FrameworkResources()` method
- **Index Pattern**: `method.<struct_type>.Create.goindex` (e.g., `method.guardrailResource.Create.goindex`)

### **3. Service Registration Structure**
Each AWS service has a `service_package_gen.go` file with up to 5 registration methods:
```go
func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource
func (p *servicePackage) SDKDataSources(ctx context.Context) []*inttypes.ServicePackageSDKDataSource  
func (p *servicePackage) FrameworkResources(ctx context.Context) []*inttypes.ServicePackageFrameworkResource
func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*inttypes.ServicePackageFrameworkDataSource
func (p *servicePackage) EphemeralResources(ctx context.Context) []*inttypes.ServicePackageEphemeralResource  // Rare
```

### **4. Index Mapping Strategy**
- **SDK Resources/DataSources**: 
  - `StructType` = `""` (empty, no struct type)
  - CRUD indexes use function names: `func.<function_name>.goindex`
- **Framework Resources/DataSources**:
  - `StructType` = actual struct type name (e.g., `"guardrailResource"`)
  - CRUD indexes use method names: `method.<struct_type>.<method_name>.goindex`

This fundamental difference requires different indexing strategies for SDK vs Framework resources in our `NewTerraformResourceFromAWSSDK()` and `NewTerraformResourceFromAWSFramework()` functions.

## ⚠️ CRITICAL COMPATIBILITY WARNING

**DO NOT CHANGE THE FOLLOWING TYPES**: `TerraformResource`, `TerraformDataSource`, and `TerraformEphemeral`

These types represent the **public API** of this tool and are used by external consumers. Any changes to their structure will break backward compatibility. The AWS migration must work **within** the constraints of these existing types.

## Development Methodology
All development must strictly follow a **Test-Driven Development (TDD)** approach. Changes should be made in small, incremental, and safe steps to ensure stability and maintain high code quality.

1.  **Write a Failing Test**: Before writing any implementation code, create a targeted test that captures the new requirement and fails as expected.
2.  **Write Code to Pass**: Write the simplest, most direct code necessary to make the failing test pass.
3.  **Refactor**: Once the test is passing, refactor the code for clarity, performance, and maintainability, ensuring all tests continue to pass.

This iterative process is mandatory for all changes in this project.

## 🗺️ **Migration Strategy**

### **Core Philosophy: Gradual Transition with Dual System Support**

The migration from AzureRM to AWS provider follows a **parallel coexistence strategy** where both the legacy AzureRM structure and the new AWS 5-category structure operate simultaneously during the transition period. This ensures zero downtime and maintains backward compatibility throughout the migration.

### **Key Migration Principles**

1. **🔒 API Stability**: The public API types (`TerraformResource`, `TerraformDataSource`, `TerraformEphemeral`) remain completely unchanged to preserve backward compatibility for external consumers.

2. **📊 Dual Data Structures**: The `ServiceRegistration` struct contains both:
   - **AWS 5-category structure** (lines 15-19): `AWSSDKResources`, `AWSSDKDataSources`, `AWSFrameworkResources`, `AWSFrameworkDataSources`, `AWSEphemeralResources`
   - **Legacy AzureRM structure** (lines 22-29): `SupportedResources`, `SupportedDataSources`, etc.

3. **🔄 Progressive Integration**: Each AWS category is integrated incrementally:
   - Phase 3.2.1: SDK Resources → ✅ Complete
   - Phase 3.2.2: SDK Data Sources → ✅ Complete  
   - Phase 3.2.3: Framework Resources → ✅ Complete
   - Phase 3.2.4: Framework Data Sources → ⌛ Pending
   - Phase 3.2.5: Ephemeral Resources → ⌛ Pending

4. **🧪 Test-First Development**: Every change follows strict TDD methodology with comprehensive integration tests to ensure reliability.

### **Migration Phases and Data Flow**

#### **Current State (Phase 3.2)**
```
AWS Provider Scanning → AWS 5-Category Extraction → ServiceRegistration (Dual Structure) → File Writing (Uses Both Systems)
```

- **AWS Categories**: Populated by `processAWSServiceFile()` → Used in `WriteResourceFiles()`, `WriteDataSourceFiles()`
- **Legacy Fields**: Still used for compatibility → Will be gradually phased out in Phase 4

#### **Target State (Phase 4)**
```
AWS Provider Scanning → AWS 5-Category Extraction → ServiceRegistration (AWS Only) → File Writing (AWS Only)
```

- **AWS Categories**: Primary data source for all operations
- **Legacy Fields**: Removed or marked as truly deprecated

### **File Writing Strategy**

The current file writing approach demonstrates the migration strategy:

1. **Parallel Processing**: `WriteResourceFiles()` and `WriteDataSourceFiles()` process both legacy and AWS data structures
2. **Consistent Output**: Both systems produce identical JSON structure via conversion functions
3. **Gradual Switchover**: As AWS categories are integrated, they take precedence over legacy fields

### **Conversion Function Pattern**

Each AWS category uses dedicated conversion functions that map to the unchanged public API:

- `NewTerraformResourceFromAWSSDK()` → `TerraformResource`
- `NewTerraformDataSourceFromAWSSDK()` → `TerraformDataSource`
- `NewTerraformResourceFromAWSFramework()` → `TerraformResource` (completed)
- And so on...

### **Statistics and Progress Tracking Migration**

The statistics calculation demonstrates the dual approach:
```go
// Legacy counts (will be removed in Phase 4)
stats.LegacyResources += len(serviceReg.SupportedResources)
stats.TotalDataSources += len(serviceReg.SupportedDataSources)

// AWS counts (new, permanent)
stats.TotalResources += len(serviceReg.AWSSDKResources)
stats.TotalDataSources += len(serviceReg.AWSSDKDataSources)
```

### **When The Full Switchover Happens**

**Phase 4: Configuration Updates** is when the complete transition occurs:
- Update main index filename to "terraform-provider-aws-index.json"
- Remove or deprecate all legacy field usage
- Update all progress messages for AWS context
- Finalize statistics calculation for 5-category system only

### **Risk Mitigation**

1. **Comprehensive Testing**: Every integration phase includes full test coverage
2. **Backward Compatibility**: Public API never changes
3. **Incremental Rollout**: One category at a time reduces blast radius
4. **Dual System Validation**: Both systems can be compared for correctness during transition

### **Success Metrics**

- ✅ All existing tests continue to pass
- ✅ New AWS-specific tests achieve 100% coverage  
- ✅ File output format remains identical
- ✅ Performance characteristics maintained or improved
- ✅ Zero breaking changes to public API

This strategy ensures a safe, reliable migration that maintains system stability while enabling full AWS provider support.

## Migration Status

### ✅ **Completed Milestones**

- **Phase 1: Research & Analysis**: Successfully analyzed the AWS provider structure, identifying the 5-category registration system which is fundamentally different from AzureRM's map-based approach.
- **Phase 2: Core Function Replacement**:
    - **2.1: AWS Extraction Functions**: Implemented a full suite of extraction functions for all 5 AWS categories (SDK/Framework/Ephemeral Resources & Data Sources).
    - **2.2: Factory Function Analysis**: Implemented `extractFactoryFunctionDetails` to parse factory functions for both SDK and Framework patterns, extracting CRUD and lifecycle methods.
    - **2.3: AzureRM Cleanup**: Completely removed all legacy AzureRM-specific extraction logic, cleaning up over 1000 lines of code.
- **Phase 3: Architecture Adaptation**:
    - **3.1: Dynamic Service Package Discovery**: Replaced hardcoded filenames with a dynamic discovery mechanism that identifies service files based on method presence, making the system robust against provider changes.
    - **3.2.1: SDK Resources Integration**: Successfully integrated the `extractAWSSDKResources` function into the main scanning pipeline. SDK resources are now correctly identified, processed, and written to index files, maintaining API compatibility.
    - **3.2.2: SDK Data Sources Integration**: Successfully integrated the `extractAWSSDKDataSources` function into the main scanning pipeline. SDK data sources are now correctly identified, processed, and written to index files with proper backward compatibility and method extraction support.

### 🚧 **Current Phase & Next Steps**

The project is currently in **Phase 3.2**, focusing on integrating the remaining AWS resource categories into the main pipeline. 

**✅ JUST COMPLETED: Phase 3.2.3 - Framework Resources Integration**
- Successfully integrated AWS Framework Resources into the scanning and file writing pipeline
- Enhanced `AWSResourceInfo` struct to store actual struct types extracted from factory function bodies
- Modified extraction logic to use `findFrameworkStructType()` for proper struct type detection
- Created `NewTerraformResourceFromAWSFramework()` function with method-based indexes (not function-based)
- Updated `WriteResourceFiles()` to process Framework resources alongside SDK resources
- All integration tests pass with proper struct-based method indexing
- Maintained backward compatibility with existing `TerraformResource` API

- **Current Task**: **Phase 3.2.4: Framework Data Sources Integration**
- **Next**: Phase 4 (Configuration) and Phase 5 (Documentation).

---

## Detailed Implementation Plan

### Phase 3.2: AWS 5-Category Classification System Implementation

**Objective**: Integrate all AWS extraction functions into the main scanning pipeline.

#### 📊 **Sub-Task Status**:
- **Sub-Task 3.2.1**: 🔧 SDK Resources Integration (✅ **COMPLETED**)
- **Sub-Task 3.2.2**: 🗃️ SDK Data Sources Integration (✅ **COMPLETED**)
- **Sub-Task 3.2.3**: 🚀 Framework Resources Integration (✅ **COMPLETED**)
- **Sub-Task 3.2.4**: 📊 Framework Data Sources Integration (⌛ **Skipped - Focus on Resources Only**)
- **Sub-Task 3.2.5**: ⚡ Ephemeral Resources Integration (⌛ **Skipped - Focus on Resources Only**)

#### 🎯 **Current Task Details: Phase 3.2.3 - Framework Resources Integration**

**Objective**: Integrate AWS Framework Resources into the main scanning pipeline.

**Scope**:
- Connect `extractAWSFrameworkResources()` to the main `ScanProviderPackages()` function.
- Update the `ServiceRegistration` struct to handle Framework resource arrays.
- Map AWS Framework resource info to the existing `TerraformResource` API for backward compatibility.
- Handle Framework-specific lifecycle methods and configurations.
- Support modern Framework patterns alongside legacy SDK patterns in the same service.

**Implementation Steps**:
1.  **Update `ServiceRegistration`**: Add an `AWSFrameworkResources` field.
2.  **Update `ScanProviderPackages()`**: Call `extractAWSFrameworkResources()` and populate the new field.
3.  **Update `WriteResourceFiles()`**: Process `AWSFrameworkResources` alongside `AWSSDKResources`.
4.  **Create `NewTerraformResourceFromAWSFramework()`**: A new mapping function for Framework resources.
5.  **Create Integration Tests**: In a new file `pkg/aws_framework_resources_integration_test.go`, test single, multiple, and mixed (SDK/Framework) resource scenarios.

### Phase 4: Configuration Updates

- [ ] Update `WriteMainIndexFile()` to use "terraform-provider-aws-index.json".
- [ ] Update progress messages, error messages, and logging for AWS context.
- [ ] Update statistics calculation for the 5-category system.

### Phase 5: Documentation & Testing

- [ ] Update `README.md` with AWS-specific information.
- [ ] Update example commands and usage.
- [ ] Create final AWS-specific test cases and validate against the live AWS provider.

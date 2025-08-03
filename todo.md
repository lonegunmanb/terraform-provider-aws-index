# AWS Provider Index Migration Plan

## Overview
This project was originally designed for indexing the AzureRM Terraform provider but needs to be adapted for the AWS Terraform provider. The main challenge is that AWS provider uses a completely different service registration pattern compared to AzureRM.

## Current Issues

### 1. AzureRM-Specific Extraction Functions
The following functions in `pkg/terraform_provider_index.go` are specifically designed for AzureRM patterns and won't work with AWS:

- `extractSupportedResourcesMappings()`
- `extractSupportedDataSourcesMappings()`
- `extractResourcesStructTypes()`
- `extractDataSourcesStructTypes()`
- `extractEphemeralResourcesFunctions()`

### 2. File Naming References
- `WriteMainIndexFile()` still references "terraform-provider-azurerm-index.json" (Line 190)

### 3. Documentation Updates Needed
- `README.md` contains AzureRM-specific references
- Project description needs updating

## Analysis Needed

### AWS Provider Structure Investigation
Before implementing fixes, we need to understand:

1. **AWS Service Registration Pattern**: How does AWS provider register resources and data sources?
2. **Directory Structure**: What's the typical structure of AWS provider services?
3. **Resource/DataSource Patterns**: How are AWS resources and data sources defined?
4. **Modern vs Legacy**: Does AWS provider use both modern SDK and legacy patterns like AzureRM?

### AWS Provider Analysis Results

After examining the AWS provider codebase in `tmp/terraform-provider-aws/`, the following key differences from AzureRM were identified:

#### Service Registration Architecture
**AWS Provider** uses a fundamentally different approach compared to **AzureRM**:

1. **Multi-Method Registration**: Instead of single `SupportedResources()` and `SupportedDataSources()` methods, AWS uses five separate methods:
   - `SDKDataSources()` - Legacy SDK data sources
   - `SDKResources()` - Legacy SDK resources  
   - `FrameworkDataSources()` - Modern Framework data sources
   - `FrameworkResources()` - Modern Framework resources
   - `EphemeralResources()` - Modern Framework ephemeral resources

2. **Structured Configuration**: Each method returns structured arrays containing:
   - `Factory` function reference
   - `TypeName` (Terraform resource type)
   - `Name` (human-readable name)
   - `Tags` configuration
   - `Region` settings
   - `Identity` and `Import` configurations

3. **Generated Service Packages**: All service registrations are in auto-generated `service_package_gen.go` files at:
   ```
   internal/service/{service}/service_package_gen.go
   ```

#### Key Architectural Differences

| Aspect | AzureRM | AWS |
|--------|---------|-----|
| Registration | Map-based (`map[string]*schema.Resource`) | Structured arrays |
| Information Source | AST parsing required | Direct configuration access |
| SDK Types | Legacy + Modern mixed | Clearly separated by method |
| Ephemeral Resources | Basic function list | Full configuration objects |
| Factory Functions | Map values | Structured Factory field |

#### AWS Provider Service Structure Example (S3)
```go
func (p *servicePackage) SDKResources(ctx context.Context) []*inttypes.ServicePackageSDKResource {
    return []*inttypes.ServicePackageSDKResource{
        {
            Factory:  resourceBucket,           // Factory function
            TypeName: "aws_s3_bucket",         // Terraform type
            Name:     "Bucket",                // Human name
            Tags: unique.Make(inttypes.ServicePackageResourceTags{
                IdentifierAttribute: names.AttrBucket,
                ResourceType:        "Bucket",
            }),
            Region: unique.Make(inttypes.ResourceRegionDefault()),
            Identity: inttypes.RegionalSingleParameterIdentity(names.AttrBucket),
        },
        // ... more resources
    }
}
```

#### Extraction Strategy Implications
The structured nature of AWS provider registration means:

1. **Direct Information Access**: No complex AST parsing needed for basic resource information
2. **Factory Function Analysis**: Still need to parse Factory functions for CRUD method details
3. **Type Classification**: Easy classification into the five categories
4. **Metadata Rich**: More metadata available directly (tags, region, identity, import)

## Implementation Plan (Updated)

### ~~Phase 1: Research & Analysis~~ ✅ COMPLETED

- [x] ~~Examine AWS provider codebase structure in `tmp/terraform-provider-aws/`~~
- [x] ~~Identify AWS-specific registration patterns~~
- [x] ~~Document differences between AzureRM and AWS approaches~~
- [x] ~~Design new extraction functions for AWS patterns~~

**Analysis Results**: AWS uses structured service package configurations instead of map-based registrations, requiring completely different extraction approaches.

### Phase 2: Core Function Replacement

**Strategy**: Complete replacement of AzureRM-specific functions with AWS-specific implementations.

#### 2.1 New AWS Extraction Functions Required

- [x] `extractAWSSDKResources()` - Extract from `SDKResources()` method **✅ COMPLETED**
- [x] `extractAWSSDKDataSources()` - Extract from `SDKDataSources()` method **✅ COMPLETED**
- [x] `extractAWSFrameworkResources()` - Extract from `FrameworkResources()` method **✅ COMPLETED**
- [x] `extractAWSFrameworkDataSources()` - Extract from `FrameworkDataSources()` method **✅ COMPLETED**
- [x] `extractAWSEphemeralResources()` - Extract from `EphemeralResources()` method **✅ COMPLETED**

#### 2.2 Factory Function Analysis
- [ ] `extractFactoryFunctionDetails()` - Parse Factory functions for CRUD methods
- [ ] Update CRUD extraction to handle both Legacy SDK and Framework patterns

#### 2.3 Remove AzureRM Functions
- [ ] Remove `extractSupportedResourcesMappings()`
- [ ] Remove `extractSupportedDataSourcesMappings()`
- [ ] Remove `extractResourcesStructTypes()`
- [ ] Remove `extractDataSourcesStructTypes()`
- [ ] Remove `extractEphemeralResourcesFunctions()`

### Phase 3: Architecture Adaptation

#### 3.1 Service Package Discovery
- [ ] Update service discovery to scan `internal/service/*/service_package_gen.go`
- [ ] Handle the new 5-category classification system
- [ ] Maintain backward compatibility for output formats

#### 3.2 Metadata Enhancement
- [ ] Leverage AWS's rich metadata (Tags, Region, Identity, Import)
- [ ] Add new fields to output JSON for AWS-specific features
- [ ] Update progress tracking for 5 categories instead of 3

### Phase 4: Configuration Updates

- [ ] Update `WriteMainIndexFile()` to use "terraform-provider-aws-index.json"
- [ ] Update progress messages for AWS context
- [ ] Update error messages and logging
- [ ] Update statistics calculation for 5-category system

### Phase 5: Documentation & Testing

- [ ] Update `README.md` with AWS-specific information
- [ ] Update example commands and usage
- [ ] Create AWS-specific test cases
- [ ] Test with actual AWS provider codebase
- [ ] Validate generated index files against AWS resources

## Implementation Recommendations

### 1. Prioritized Approach
**Recommended Order**:
1. Start with SDK Resources (most similar to existing logic)
2. Add SDK DataSources
3. Add Framework Resources and DataSources
4. Add Ephemeral Resources (newest feature)

### 2. Backward Compatibility Considerations
- **Recommendation**: Focus solely on AWS, remove AzureRM support
- **Rationale**: The architectures are fundamentally incompatible
- **Alternative**: If backward compatibility needed, create separate code paths

### 3. Testing Strategy
- Use existing AWS provider in `tmp/terraform-provider-aws/`
- Focus on high-volume services (S3, EC2, IAM) for initial testing
- Validate against known resource counts (~2000+ resources, ~400+ data sources)

### 4. Error Handling
- Handle missing service packages gracefully
- Provide clear error messages for parsing failures
- Add validation for Factory function availability

## Questions for Discussion

### ✅ Resolved Through Analysis

1. **AWS Provider Analysis**: ~~Should we start by examining the AWS provider structure in `tmp/terraform-provider-aws/`?~~
   - **RESOLVED**: Analysis completed. AWS uses 5-category structured registration system.

2. **Architecture Understanding**: ~~How does AWS provider register resources and data sources?~~
   - **RESOLVED**: AWS uses `service_package_gen.go` files with structured arrays instead of map-based registration.

### 🔄 Updated Questions for Discussion

1. **Implementation Scope**: Should we implement all 5 categories simultaneously or phase them?
   - **Recommendation**: Start with SDK Resources/DataSources (most similar to existing), then add Framework types

2. **Backward Compatibility**: Do we want to maintain support for both AzureRM and AWS, or focus solely on AWS?
   - **Recommendation**: Focus solely on AWS due to fundamental architectural differences
   - **Alternative**: Create separate tool/branch for AWS if backward compatibility is required

3. **Function Naming Strategy**: How should we name the new extraction functions?
   - **Recommendation**: Use AWS-specific names (`extractAWSSDKResources`) for clarity
   - **Alternative**: Generic names (`extractProviderResources`) if planning future multi-provider support

4. **Output Format Evolution**: Should we enhance JSON output to include AWS-specific metadata?
   - **Proposal**: Add optional fields for Tags, Region, Identity, Import configurations
   - **Maintain**: Backward compatibility with existing JSON structure

5. **Testing Strategy**: How should we validate the new AWS extraction logic?
   - **Proposal**: Use known AWS resource counts as validation benchmarks
   - **Method**: Compare against official AWS provider documentation

### 🆕 New Technical Decisions Needed

1. **Service Package Discovery**: Should we scan all `internal/service/*/service_package_gen.go` files or use a service registry?

2. **Factory Function Analysis**: How deep should we parse Factory functions for CRUD details?

3. **Error Handling**: How should we handle missing or malformed service package files?

4. **Performance**: Should we implement parallel processing for the 5 categories?

## Next Steps

**PENDING APPROVAL**: All changes must be approved before implementation.

1. First, we should examine the AWS provider codebase structure
2. Identify the specific patterns used by AWS provider
3. Design new extraction functions based on AWS patterns
4. Create a detailed implementation plan for each function

---

## Status Update

**Current Status**: Analysis Completed - Ready for Implementation

### Phase 1 Results ✅
- **AWS Provider Architecture**: Fully analyzed and documented
- **Key Differences**: Identified 5-category system vs AzureRM's map-based approach  
- **Implementation Strategy**: Defined with prioritized phases
- **Technical Challenges**: Mapped out with solutions

### Next Immediate Actions � **CURRENT STATUS**
1. **✅ Implementation Started**: Begin with `extractAWSSDKResources()` function
2. **🚧 TDD Phase**: Created comprehensive test cases for AWS SDK resource extraction
3. **⏳ Implementation Phase**: Need to implement actual extraction function to pass tests

### Recent Progress Updates ⚡
- **✅ Test Framework**: Created `aws_extractor_test.go` with comprehensive test cases
- **✅ Data Structures**: Defined `AWSResourceInfo`, `AWSTagsConfig`, `AWSRegionConfig` structures
- **✅ Test Coverage**: 4 test scenarios covering main extraction patterns:
  - Direct return from `SDKResources()` method
  - Empty method handling
  - Variable assignment pattern
  - Method not found scenario
- **⏳ Implementation**: Need to create `extractAWSSDKResources()` function in `aws_extractor.go`

### Key Implementation Insights
- **Complexity Reduction**: AWS's structured approach actually simplifies extraction vs AzureRM's AST parsing
- **Enhanced Metadata**: AWS provides richer metadata directly in service configurations
- **Clear Separation**: The 5-category system provides clean architectural boundaries

**Ready to Proceed**: All research completed, implementation plan defined, awaiting approval to begin Phase 2.

### Development Methodology Requirement ⚠️
- **TDD Mandatory**: All subsequent changes must strictly follow Test-Driven Development (TDD) approach
- **Implementation Order**: Write tests first, then implement functionality to pass the tests
- **Test Coverage**: Ensure comprehensive test coverage for all new AWS extraction functions

---

## Current Implementation Status (August 3, 2025)

### Phase 2.1: AWS Extraction Functions - 🚧 80% COMPLETED

#### ✅ Completed Functions:

1. **✅ extractAWSSDKResources()**: Extracts SDK resources from `SDKResources()` method
   - **Test Coverage**: 4 comprehensive test scenarios 
   - **Implementation**: Complete AST parsing for SDK resource structures
   - **Status**: All tests passing ✅

2. **✅ extractAWSSDKDataSources()**: Extracts SDK data sources from `SDKDataSources()` method  
   - **Test Coverage**: 4 comprehensive test scenarios
   - **Implementation**: Complete AST parsing for SDK data source structures
   - **Status**: All tests passing ✅

3. **✅ extractAWSFrameworkResources()**: Extracts Framework resources from `FrameworkResources()` method
   - **Test Coverage**: 4 comprehensive test scenarios
   - **Implementation**: Complete AST parsing for Framework resource structures  
   - **Status**: All tests passing ✅

4. **✅ extractAWSFrameworkDataSources()**: Extracts Framework data sources from `FrameworkDataSources()` method
   - **Test Coverage**: 4 comprehensive test scenarios
   - **Implementation**: Complete AST parsing for Framework data source structures
   - **Status**: All tests passing ✅

#### 🔄 Next Implementation (Phase 2.1 Final):

**🎯 Current Target**: All Phase 2.1 functions completed! **✅ COMPLETED**

#### ⏳ Remaining Functions:
- [x] `extractAWSEphemeralResources()` - Extract from `EphemeralResources()` method **✅ COMPLETED**

#### 📁 Files Structure:
- `pkg/aws_extractor_test.go` - Comprehensive test cases for all AWS extraction functions
- `pkg/aws_extractor.go` - Implementation of AWS-specific extraction functions

#### 🔍 Key Implementation Patterns Established:
- **TDD Approach**: Write tests first, then implement functionality
- **AST Parsing**: Robust parsing of AWS service package structures
- **Multiple Return Patterns**: Support for direct returns, variable assignments, and declarations
- **Rich Metadata Extraction**: Tags, Region, Identity, and Import configurations
- **Error Handling**: Graceful handling of missing methods and malformed structures

#### 📊 Progress Summary:
- **SDK Functions**: 2/2 completed (100%) ✅
- **Framework Functions**: 2/2 completed (100%) ✅  
- **Ephemeral Functions**: 0/1 completed (0%) ⏳
- **Overall Progress**: 4/5 functions completed (80%) 🚧

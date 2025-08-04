# AWS Provider Index Migration Plan

## Overview
This project was originally designed for indexing the AzureRM Terraform provider but needs to be adapted for the AWS Terraform provider. The main challenge is that AWS provider uses a completely different service registration pattern compared to AzureRM.

## ⚠️ CRITICAL COMPATIBILITY WARNING

**DO NOT CHANGE THE FOLLOWING TYPES**: `TerraformResource`, `TerraformDataSource`, and `TerraformEphemeral`

These types represent the **public API** of this tool and are used by external consumers. Any changes to their structure will break backward compatibility. The AWS migration must work **within** the constraints of these existing types.

## Development Methodology
All development must strictly follow a **Test-Driven Development (TDD)** approach. Changes should be made in small, incremental, and safe steps to ensure stability and maintain high code quality.

1.  **Write a Failing Test**: Before writing any implementation code, create a targeted test that captures the new requirement and fails as expected.
2.  **Write Code to Pass**: Write the simplest, most direct code necessary to make the failing test pass.
3.  **Refactor**: Once the test is passing, refactor the code for clarity, performance, and maintainability, ensuring all tests continue to pass.

This iterative process is mandatory for all changes in this project.

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

### 🚧 **Current Phase & Next Steps**

The project is currently in **Phase 3.2**, focusing on integrating the remaining AWS resource categories into the main pipeline. To maintain focus, we are prioritizing **resources only** and skipping data sources and ephemeral resources for now.

- **Current Task**: **Phase 3.2.3: Framework Resources Integration**
- **Next**: Phase 4 (Configuration) and Phase 5 (Documentation).

---

## Detailed Implementation Plan

### Phase 3.2: AWS 5-Category Classification System Implementation

**Objective**: Integrate all AWS extraction functions into the main scanning pipeline.

#### 📊 **Sub-Task Status**:
- **Sub-Task 3.2.1**: 🔧 SDK Resources Integration (✅ **COMPLETED**)
- **Sub-Task 3.2.2**: 🗃️ SDK Data Sources Integration (⌛ **Skipped - Focus on Resources Only**)
- **Sub-Task 3.2.3**: 🚀 Framework Resources Integration (🎯 **CURRENT TASK**)
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

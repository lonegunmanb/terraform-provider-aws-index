package pkg

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/spf13/afero"
)

var outputFs = afero.NewOsFs()
var inputFs = afero.NewOsFs() // Add filesystem abstraction for input operations

// TerraformProviderIndex represents the complete index of a Terraform provider
type TerraformProviderIndex struct {
	Version    string                `json:"version"`    // Provider version
	Services   []ServiceRegistration `json:"services"`   // All service registrations
	Statistics ProviderStatistics    `json:"statistics"` // Summary statistics
}

// ScanTerraformProviderServices scans the specified directory for Terraform provider services
// and extracts all registration information into a structured index
func ScanTerraformProviderServices(dir, basePkgUrl string, version string, progressCallback ProgressCallback) (*TerraformProviderIndex, error) {
	// Read the services directory to get all service subdirectories
	entries, err := afero.ReadDir(inputFs, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read services directory: %w", err)
	}

	// Filter entries to only include directories
	var dirEntries []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			dirEntries = append(dirEntries, entry)
		}
	}

	totalServices := len(dirEntries)
	if totalServices == 0 {
		return &TerraformProviderIndex{
			Version:    version,
			Services:   []ServiceRegistration{},
			Statistics: ProviderStatistics{},
		}, nil
	}

	// Create progress tracker
	progressTracker := NewProgressTracker("scanning", totalServices, progressCallback)

	// Set up parallel processing
	numWorkers := runtime.NumCPU()
	if numWorkers > len(dirEntries) {
		numWorkers = len(dirEntries)
	}

	// Channels for work distribution and result collection
	entryChan := make(chan os.FileInfo, len(dirEntries))
	resultChan := make(chan ServiceRegistration, len(dirEntries))
	var wg sync.WaitGroup

	// Send all directory entries to the work channel
	for _, entry := range dirEntries {
		entryChan <- entry
	}
	close(entryChan)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range entryChan {
				servicePath := filepath.Join(dir, entry.Name())

				// Scan the individual service package
				packageInfo, err := gophon.ScanSinglePackage(servicePath, basePkgUrl)

				// Update progress
				progressTracker.UpdateProgress(entry.Name())

				if err != nil || packageInfo == nil || len(packageInfo.Files) == 0 {
					// Skip services that can't be scanned (might not be valid Go packages)
					continue
				}

				serviceReg := newServiceRegistration(packageInfo, entry)

				// Phase 3: Use annotation-based scanning instead of file-by-file parsing
				err = parseAWSServiceFileWithAnnotations(packageInfo, &serviceReg)
				if err != nil {
					// Log error but continue with other services
					continue
				}

				// NOTE: extractAndStoreSDKCRUDMethodsForLegacyPlugin is no longer needed
				// because CRUD methods are now extracted directly by the annotation scanner

				// Only include services that have at least one AWS registration method
				if len(serviceReg.AWSSDKResources) > 0 || len(serviceReg.AWSSDKDataSources) > 0 ||
					len(serviceReg.AWSFrameworkResources) > 0 || len(serviceReg.AWSFrameworkDataSources) > 0 ||
					len(serviceReg.AWSEphemeralResources) > 0 {
					resultChan <- serviceReg
				}
			}
		}()
	}

	// Wait for all workers to complete and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and build final data structures
	var services []ServiceRegistration
	stats := ProviderStatistics{}

	for serviceReg := range resultChan {
		services = append(services, serviceReg)
		stats.ServiceCount++

		// AWS 5-category statistics
		stats.TotalResources += len(serviceReg.AWSSDKResources)
		stats.TotalResources += len(serviceReg.AWSFrameworkResources)
		stats.TotalDataSources += len(serviceReg.AWSSDKDataSources)
		stats.TotalDataSources += len(serviceReg.AWSFrameworkDataSources)
		stats.EphemeralResources += len(serviceReg.AWSEphemeralResources)
	}

	// Final statistics calculation
	stats.LegacyResources = 0 // No longer used
	stats.ModernResources = 0 // No longer used

	// Report scanning completion
	progressTracker.Complete()

	return &TerraformProviderIndex{
		Version:    version,
		Services:   services,
		Statistics: stats,
	}, nil
}

// WriteIndexFiles writes all index files to the specified output directory
// This is the main method that orchestrates writing all index files
func (index *TerraformProviderIndex) WriteIndexFiles(outputDir string, progressCallback ProgressCallback) error {
	// Calculate total number of files to write
	totalFiles := 1 // main index file
	for _, service := range index.Services {
		// AWS 5-category file counts
		totalFiles += len(service.AWSSDKResources)         // AWS SDK resources
		totalFiles += len(service.AWSFrameworkResources)   // AWS Framework resources
		totalFiles += len(service.AWSSDKDataSources)       // AWS SDK data sources
		totalFiles += len(service.AWSFrameworkDataSources) // AWS Framework data sources
		totalFiles += len(service.AWSEphemeralResources)   // AWS Ephemeral resources
		totalFiles += len(service.EphemeralTerraformTypes) // Framework ephemeral resources (backward compatibility)
	}

	// Create progress tracker
	progressTracker := NewProgressTracker("indexing", totalFiles, progressCallback)

	// Create directory structure
	if err := index.CreateDirectoryStructure(outputDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Write main index file
	if err := index.WriteMainIndexFile(outputDir); err != nil {
		return fmt.Errorf("failed to write main index file: %w", err)
	}
	progressTracker.UpdateProgress("main index file")

	// Write individual resource files
	if err := index.WriteResourceFiles(outputDir, progressTracker); err != nil {
		return fmt.Errorf("failed to write resource files: %w", err)
	}

	// Write individual data source files
	if err := index.WriteDataSourceFiles(outputDir, progressTracker); err != nil {
		return fmt.Errorf("failed to write data source files: %w", err)
	}

	// Write individual ephemeral resource files
	if err := index.WriteEphemeralFiles(outputDir, progressTracker); err != nil {
		return fmt.Errorf("failed to write ephemeral files: %w", err)
	}

	// Report completion
	progressTracker.Complete()

	return nil
}

// WriteMainIndexFile writes the main terraform-provider-aws-index.json file
func (index *TerraformProviderIndex) WriteMainIndexFile(outputDir string) error {
	mainIndexPath := filepath.Join(outputDir, "terraform-provider-aws-index.json")
	return index.WriteJSONFile(mainIndexPath, index)
}

// processCallbacksParallel runs a slice of callbacks in parallel
func processCallbacksParallel(tasks []func() error) error {
	if len(tasks) == 0 {
		return nil
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(tasks) {
		numWorkers = len(tasks)
	}

	callbackChan := make(chan func() error, len(tasks))
	errorChan := make(chan error, len(tasks))
	var wg sync.WaitGroup

	// Send all callbacks to the channel
	for _, callback := range tasks {
		callbackChan <- callback
	}
	close(callbackChan)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for callback := range callbackChan {
				if err := callback(); err != nil {
					errorChan <- err
					return
				}
			}
		}()
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteResourceFiles writes individual JSON files for each resource
func (index *TerraformProviderIndex) WriteResourceFiles(outputDir string, progressTracker *ProgressTracker) error {
	resourcesDir := filepath.Join(outputDir, "resources")
	var tasks []func() error

	for _, service := range index.Services {
		// Process AWS SDK resources
		for terraformType, awsResourceInfo := range service.AWSSDKResources {
			// Capture variables for closure
			tfType := terraformType
			awsResource := awsResourceInfo
			svc := service

			tasks = append(tasks, func() error {
				// Create AWS-specific resource info using only core TerraformResource fields
				awsResourceData := NewTerraformResourceFromAWSSDK(awsResource, svc)

				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(resourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, awsResourceData); err != nil {
					return fmt.Errorf("failed to write AWS SDK resource file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("resource %s", tfType))
				return nil
			})
		}

		// Process AWS Framework resources (NEW)
		for terraformType, awsResourceInfo := range service.AWSFrameworkResources {
			// Capture variables for closure
			tfType := terraformType
			awsResource := awsResourceInfo
			svc := service

			tasks = append(tasks, func() error {
				// Create AWS Framework-specific resource info using only core TerraformResource fields
				awsResourceData := NewTerraformResourceFromAWSFramework(awsResource, svc)

				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(resourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, awsResourceData); err != nil {
					return fmt.Errorf("failed to write AWS Framework resource file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("resource %s", tfType))
				return nil
			})
		}
	}

	return processCallbacksParallel(tasks)
}

// WriteDataSourceFiles writes individual JSON files for each data source
func (index *TerraformProviderIndex) WriteDataSourceFiles(outputDir string, progressTracker *ProgressTracker) error {
	dataSourcesDir := filepath.Join(outputDir, "datasources")
	var tasks []func() error

	for _, service := range index.Services {
		// Process AWS SDK data sources
		for terraformType, awsDataSourceInfo := range service.AWSSDKDataSources {
			// Capture variables for closure
			tfType := terraformType
			awsDataSource := awsDataSourceInfo
			svc := service

			tasks = append(tasks, func() error {
				// Create AWS-specific data source info using only core TerraformDataSource fields
				awsDataSourceData := NewTerraformDataSourceFromAWSSDK(awsDataSource, svc)

				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(dataSourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, awsDataSourceData); err != nil {
					return fmt.Errorf("failed to write AWS SDK data source file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("data source %s", tfType))
				return nil
			})
		}

		// Process AWS Framework data sources (NEW)
		for terraformType, awsDataSourceInfo := range service.AWSFrameworkDataSources {
			// Capture variables for closure
			tfType := terraformType
			awsDataSource := awsDataSourceInfo
			svc := service

			tasks = append(tasks, func() error {
				// Create AWS Framework-specific data source info using only core TerraformDataSource fields
				awsDataSourceData := NewTerraformDataSourceFromAWSFramework(awsDataSource, svc)

				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(dataSourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, awsDataSourceData); err != nil {
					return fmt.Errorf("failed to write AWS Framework data source file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("data source %s", tfType))
				return nil
			})
		}
	}

	return processCallbacksParallel(tasks)
}

// WriteEphemeralFiles writes individual JSON files for each ephemeral resource
func (index *TerraformProviderIndex) WriteEphemeralFiles(outputDir string, progressTracker *ProgressTracker) error {
	ephemeralDir := filepath.Join(outputDir, "ephemeral")

	// Ensure ephemeral directory exists even if no files will be written
	if err := outputFs.MkdirAll(ephemeralDir, 0755); err != nil {
		return fmt.Errorf("failed to create ephemeral directory %s: %w", ephemeralDir, err)
	}

	var tasks []func() error

	for _, service := range index.Services {
		// Process legacy ephemeral resources (for backward compatibility)
		for structType, tfType := range service.EphemeralTerraformTypes {
			// Capture variables for closure
			structT := structType
			svc := service
			terraformType := tfType

			tasks = append(tasks, func() error {

				ephemeralInfo := NewTerraformEphemeralInfo(structT, svc)
				fileName := fmt.Sprintf("%s.json", terraformType)
				filePath := filepath.Join(ephemeralDir, fileName)

				if err := index.WriteJSONFile(filePath, ephemeralInfo); err != nil {
					return fmt.Errorf("failed to write ephemeral resource file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("ephemeral %s", terraformType))
				return nil
			})
		}

		// Process AWS ephemeral resources (NEW - Phase 3.2.5)
		for _, awsEphemeral := range service.AWSEphemeralResources {
			// Capture variables for closure
			ephemeral := awsEphemeral
			svc := service

			tasks = append(tasks, func() error {
				ephemeralInfo := NewTerraformEphemeralFromAWS(ephemeral, svc)
				fileName := fmt.Sprintf("%s.json", ephemeral.TerraformType)
				filePath := filepath.Join(ephemeralDir, fileName)

				if err := index.WriteJSONFile(filePath, ephemeralInfo); err != nil {
					return fmt.Errorf("failed to write AWS ephemeral resource file %s: %w", fileName, err)
				}

				if progressTracker != nil {
					progressTracker.UpdateProgress(fmt.Sprintf("ephemeral %s", ephemeral.TerraformType))
				}
				return nil
			})
		}
	}

	return processCallbacksParallel(tasks)
}

// CreateDirectoryStructure creates the required directory structure for index files
func (index *TerraformProviderIndex) CreateDirectoryStructure(outputDir string) error {
	dirs := []string{
		outputDir,
		filepath.Join(outputDir, "resources"),
		filepath.Join(outputDir, "datasources"),
		filepath.Join(outputDir, "ephemeral"),
	}

	for _, dir := range dirs {
		if err := outputFs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// WriteJSONFile writes data as JSON to the specified file path
func (index *TerraformProviderIndex) WriteJSONFile(filePath string, data interface{}) error {
	// Ensure parent directory exists
	parentDir := filepath.Dir(filePath)
	if err := outputFs.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}

	// Marshal data to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Write to file
	if err := afero.WriteFile(outputFs, filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// =============================================================================
// Phase 3 Integration Functions: Annotation-based scanning
// =============================================================================

// parseAWSServiceFileWithAnnotations replaces parseAWSServiceFile with annotation-based scanning
// This is the new Phase 3 integration function that uses the annotation scanner
func parseAWSServiceFileWithAnnotations(packageInfo *gophon.PackageInfo, serviceReg *ServiceRegistration) error {
	// Use the annotation scanner to find all annotations in the package
	annotationResults, err := ScanPackageForAnnotations(packageInfo)
	if err != nil {
		return fmt.Errorf("failed to scan package for annotations: %w", err)
	}

	// Convert annotation results to service registration format
	convertAnnotationResultsToServiceRegistration(annotationResults, serviceReg)

	return nil
}

// convertAnnotationResultsToServiceRegistration converts AnnotationResults to ServiceRegistration format
// This function bridges the annotation scanner output with the existing ServiceRegistration structure
func convertAnnotationResultsToServiceRegistration(results *AnnotationResults, serviceReg *ServiceRegistration) {
	// Process SDK Resources
	for _, annotation := range results.SDKResources {
		resourceInfo := AWSResourceInfo{
			TerraformType:   annotation.TerraformType,
			FactoryFunction: extractFactoryFunctionNameFromTerraformType(annotation.TerraformType, "resource"),
			Name:            annotation.Name,
			SDKType:         "sdk",
			StructType:      "", // SDK resources don't have struct types
		}
		serviceReg.AWSSDKResources[annotation.TerraformType] = resourceInfo

		// Store CRUD methods if available from annotation scanner
		if len(annotation.CRUDMethods) > 0 {
			legacyCRUD := &LegacyResourceCRUDFunctions{
				CreateMethod: annotation.CRUDMethods["create"],
				ReadMethod:   annotation.CRUDMethods["read"],
				UpdateMethod: annotation.CRUDMethods["update"],
				DeleteMethod: annotation.CRUDMethods["delete"],
			}
			serviceReg.ResourceCRUDMethods[annotation.TerraformType] = legacyCRUD
		}
	}

	// Process SDK Data Sources
	for _, annotation := range results.SDKDataSources {
		resourceInfo := AWSResourceInfo{
			TerraformType:   annotation.TerraformType,
			FactoryFunction: extractFactoryFunctionNameFromTerraformType(annotation.TerraformType, "dataSource"),
			Name:            annotation.Name,
			SDKType:         "sdk",
			StructType:      "", // SDK data sources don't have struct types
		}
		serviceReg.AWSSDKDataSources[annotation.TerraformType] = resourceInfo

		// Store read method if available from annotation scanner
		if readMethod, exists := annotation.CRUDMethods["read"]; exists && readMethod != "" {
			legacyDataSource := &LegacyDataSourceMethods{
				ReadMethod: readMethod,
			}
			serviceReg.DataSourceMethods[annotation.TerraformType] = legacyDataSource
		}
	}

	// Process Framework Resources
	for _, annotation := range results.FrameworkResources {
		resourceInfo := AWSResourceInfo{
			TerraformType:   annotation.TerraformType,
			FactoryFunction: extractFactoryFunctionNameFromTerraformType(annotation.TerraformType, "frameworkResource"),
			Name:            annotation.Name,
			SDKType:         "framework",
			StructType:      annotation.StructType,
		}
		serviceReg.AWSFrameworkResources[annotation.TerraformType] = resourceInfo

		// Store struct type to terraform type mapping for framework resources
		if annotation.StructType != "" {
			serviceReg.ResourceTerraformTypes[annotation.StructType] = annotation.TerraformType
		}
	}

	// Process Framework Data Sources
	for _, annotation := range results.FrameworkDataSources {
		resourceInfo := AWSResourceInfo{
			TerraformType:   annotation.TerraformType,
			FactoryFunction: extractFactoryFunctionNameFromTerraformType(annotation.TerraformType, "frameworkDataSource"),
			Name:            annotation.Name,
			SDKType:         "framework",
			StructType:      annotation.StructType,
		}
		serviceReg.AWSFrameworkDataSources[annotation.TerraformType] = resourceInfo

		// Store struct type to terraform type mapping for framework data sources
		if annotation.StructType != "" {
			serviceReg.DataSourceTerraformTypes[annotation.StructType] = annotation.TerraformType
		}
	}

	// Process Ephemeral Resources
	for _, annotation := range results.EphemeralResources {
		resourceInfo := AWSResourceInfo{
			TerraformType:   annotation.TerraformType,
			FactoryFunction: extractFactoryFunctionNameFromTerraformType(annotation.TerraformType, "ephemeral"),
			Name:            annotation.Name,
			SDKType:         "ephemeral",
			StructType:      annotation.StructType,
		}
		serviceReg.AWSEphemeralResources[annotation.TerraformType] = resourceInfo

		// Store struct type to terraform type mapping for ephemeral resources
		if annotation.StructType != "" {
			serviceReg.EphemeralTerraformTypes[annotation.StructType] = annotation.TerraformType
		}
	}
}

// extractFactoryFunctionNameFromTerraformType extracts the likely factory function name from terraform type
// This function tries to infer the factory function name based on AWS provider naming conventions
func extractFactoryFunctionNameFromTerraformType(terraformType, functionType string) string {
	// Convert terraform type to likely function name
	// e.g., "aws_lambda_function" -> "resourceFunction" or "dataSourceFunction"
	
	// Remove "aws_" prefix
	suffix := terraformType
	if strings.HasPrefix(terraformType, "aws_") {
		suffix = terraformType[4:]
	}
	
	// Convert underscores to camelCase
	parts := strings.Split(suffix, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	
	functionBase := strings.Join(parts, "")
	
	// Return factory function name based on type following AWS provider conventions
	switch functionType {
	case "resource":
		return "resource" + functionBase
	case "dataSource":
		return "dataSource" + functionBase
	case "frameworkResource":
		return "new" + functionBase + "Resource"
	case "frameworkDataSource":
		return "new" + functionBase + "DataSource"
	case "ephemeral":
		return "new" + functionBase + "EphemeralResource"
	default:
		return "resource" + functionBase // Default fallback
	}
}

// =============================================================================
// End Phase 3 Integration Functions
// =============================================================================

// convertFunctionNamesToStructNames converts ephemeral resource function names to struct names
// by looking up the function declarations in PackageInfo and parsing their return statements
// For example: "NewKeyVaultSecretEphemeralResource" -> "KeyVaultSecretEphemeralResource"
func convertFunctionNamesToStructNames(functionNames []string, packageInfo *gophon.PackageInfo) []string {
	if packageInfo == nil || packageInfo.Functions == nil {
		return functionNames // Return as-is if no package info available
	}

	structNames := make([]string, 0, len(functionNames))

	for _, funcName := range functionNames {
		// Find the function in the gophon function data
		structName := ""
		for _, funcInfo := range packageInfo.Functions {
			if funcInfo.Name == funcName && funcInfo.FuncDecl != nil {
				// Extract struct type from the function's return statement
				if extracted := extractStructTypeFromEphemeralFunction(funcInfo.FuncDecl); extracted != "" {
					structName = extracted
					break
				}
			}
		}

		// If we couldn't extract from AST, fall back to string manipulation
		if structName == "" {
			if len(funcName) > 3 && funcName[:3] == "New" {
				structName = funcName[3:] // Remove "New" prefix
			} else {
				structName = funcName // Use as-is
			}
		}

		structNames = append(structNames, structName)
	}

	return structNames
}

// extractStructTypeFromEphemeralFunction extracts the struct type name from an ephemeral resource function
// For example, from: func NewKeyVaultSecretEphemeralResource() ephemeral.EphemeralResource { return &KeyVaultSecretEphemeralResource{} }
// It extracts: "KeyVaultSecretEphemeralResource"
func extractStructTypeFromEphemeralFunction(funcDecl *ast.FuncDecl) string {
	if funcDecl == nil || funcDecl.Body == nil {
		return ""
	}

	// Look for return statements in the function body
	for _, stmt := range funcDecl.Body.List {
		if returnStmt, ok := stmt.(*ast.ReturnStmt); ok {
			for _, result := range returnStmt.Results {
				// Handle &StructName{} pattern
				if unaryExpr, ok := result.(*ast.UnaryExpr); ok && unaryExpr.Op == token.AND {
					if compLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
						if ident, ok := compLit.Type.(*ast.Ident); ok {
							return ident.Name
						}
					}
				}
				// Handle StructName{} pattern (without &)
				if compLit, ok := result.(*ast.CompositeLit); ok {
					if ident, ok := compLit.Type.(*ast.Ident); ok {
						return ident.Name
					}
				}
			}
		}
	}

	return ""
}

// parseAWSServiceFile extracts all AWS registration methods from a single identified service file
func parseAWSServiceFile(fileInfo *gophon.FileInfo, serviceReg *ServiceRegistration) {
	if fileInfo.File == nil {
		return
	}

	// Extract all registration methods from this file using AWS-specific functions
	awsSDKResources := extractAWSSDKResources(fileInfo.File)
	awsSDKDataSources := extractAWSSDKDataSources(fileInfo.File)
	awsFrameworkResources := extractAWSFrameworkResources(fileInfo.File)
	awsFrameworkDataSources := extractAWSFrameworkDataSources(fileInfo.File)
	awsEphemeralResources := extractAWSEphemeralResources(fileInfo.File)

	// Merge AWS results into service registration
	for terraformType, resourceInfo := range awsSDKResources {
		serviceReg.AWSSDKResources[terraformType] = resourceInfo
	}
	for terraformType, resourceInfo := range awsSDKDataSources {
		serviceReg.AWSSDKDataSources[terraformType] = resourceInfo
	}
	for terraformType, resourceInfo := range awsFrameworkResources {
		serviceReg.AWSFrameworkResources[terraformType] = resourceInfo
	}
	for terraformType, resourceInfo := range awsFrameworkDataSources {
		serviceReg.AWSFrameworkDataSources[terraformType] = resourceInfo
	}
	for terraformType, resourceInfo := range awsEphemeralResources {
		serviceReg.AWSEphemeralResources[terraformType] = resourceInfo
	}
}

// extractAndStoreSDKCRUDMethodsForLegacyPlugin extracts CRUD methods from AWS SDK resources and data sources
// and stores them in the service registration for backward compatibility.
// Framework and Ephemeral resources use different patterns (struct methods) and don't need CRUD extraction.
func extractAndStoreSDKCRUDMethodsForLegacyPlugin(packageInfo *gophon.PackageInfo, serviceReg *ServiceRegistration) {
	// Framework and Ephemeral resources use different patterns (struct methods) and don't need CRUD extraction
	awsSDKItems := make(map[string]AWSResourceInfo)

	// Only merge SDK resources and data sources for CRUD analysis
	for terraformType, resourceInfo := range serviceReg.AWSSDKResources {
		awsSDKItems[terraformType] = resourceInfo
	}
	for terraformType, resourceInfo := range serviceReg.AWSSDKDataSources {
		awsSDKItems[terraformType] = resourceInfo
	}

	// Extract CRUD methods for SDK resources and data sources only
	for terraformType, resourceInfo := range awsSDKItems {
		if resourceInfo.FactoryFunction != "" && resourceInfo.SDKType == "sdk" {
			funcDecl := serviceReg.functions[resourceInfo.FactoryFunction]
			if funcDecl != nil {
				if crudMethods := extractFactoryFunctionDetails(funcDecl.FuncDecl); crudMethods != nil {
					// Check if this is a resource (has CRUD operations) or data source (only read)
					isDataSource := false

					// Check if this terraform type is in SDK data sources
					if _, exists := serviceReg.AWSSDKDataSources[terraformType]; exists {
						isDataSource = true
					}

					if isDataSource && crudMethods.ReadMethod != "" {
						// Store data source methods
						legacyDataSource := &LegacyDataSourceMethods{
							ReadMethod: crudMethods.ReadMethod,
						}
						serviceReg.DataSourceMethods[terraformType] = legacyDataSource
					} else if !isDataSource && (crudMethods.CreateMethod != "" || crudMethods.ReadMethod != "") {
						// Store resource CRUD methods
						legacyCRUD := &LegacyResourceCRUDFunctions{
							CreateMethod: crudMethods.CreateMethod,
							ReadMethod:   crudMethods.ReadMethod,
							UpdateMethod: crudMethods.UpdateMethod,
							DeleteMethod: crudMethods.DeleteMethod,
						}
						serviceReg.ResourceCRUDMethods[terraformType] = legacyCRUD
					}
				}
			}
		}
	}
}

package pkg

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/spf13/afero"
)

var outputFs = afero.NewOsFs()

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
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read services directory: %w", err)
	}

	// Filter entries to only include directories
	var dirEntries []os.DirEntry
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
	entryChan := make(chan os.DirEntry, len(dirEntries))
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

				// NEW: Use dynamic file detection to find ALL service files
				serviceFiles, err := identifyServicePackageFiles(packageInfo)
				if err != nil || len(serviceFiles) == 0 {
					// Skip packages without AWS service methods
					continue
				}

				// Process all identified service files and merge results
				for _, serviceFile := range serviceFiles {
					processAWSServiceFile(serviceFile, &serviceReg)
				}

				// Extract CRUD methods for AWS factory functions
				allAWSResources := make(map[string]AWSResourceInfo)
				
				// Merge all AWS resource types for CRUD analysis
				for terraformType, resourceInfo := range serviceReg.AWSSDKResources {
					allAWSResources[terraformType] = resourceInfo
				}
				for terraformType, resourceInfo := range serviceReg.AWSSDKDataSources {
					allAWSResources[terraformType] = resourceInfo
				}
				for terraformType, resourceInfo := range serviceReg.AWSFrameworkResources {
					allAWSResources[terraformType] = resourceInfo
				}
				for terraformType, resourceInfo := range serviceReg.AWSFrameworkDataSources {
					allAWSResources[terraformType] = resourceInfo
				}
				for terraformType, resourceInfo := range serviceReg.AWSEphemeralResources {
					allAWSResources[terraformType] = resourceInfo
				}

				// Extract CRUD methods for each AWS resource/data source/ephemeral
				for terraformType, resourceInfo := range allAWSResources {
					if resourceInfo.FactoryFunction != "" {
						// Find the factory function and extract CRUD details
						for _, fileInfo := range packageInfo.Files {
							if fileInfo.File != nil {
								if crudMethods := extractFactoryFunctionDetails(fileInfo.File, resourceInfo.FactoryFunction); crudMethods != nil {
									// Store CRUD methods in the resource info or service registration
									// For now, we can add this to legacy structure for compatibility
									if resourceInfo.SDKType == "sdk" {
										if crudMethods.CreateMethod != "" || crudMethods.ReadMethod != "" {
											legacyCRUD := &LegacyResourceCRUDFunctions{
												CreateMethod: crudMethods.CreateMethod,
												ReadMethod:   crudMethods.ReadMethod,
												UpdateMethod: crudMethods.UpdateMethod,
												DeleteMethod: crudMethods.DeleteMethod,
											}
											serviceReg.ResourceCRUDMethods[terraformType] = legacyCRUD
										}
									}
									break
								}
							}
						}
					}
				}

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
	globalResources := make(map[string]string)
	globalDataSources := make(map[string]string)
	stats := ProviderStatistics{}

	for serviceReg := range resultChan {
		services = append(services, serviceReg)
		stats.ServiceCount++

		// Add to global maps
		globalResources = mergeMap(globalResources, serviceReg.SupportedResources)
		globalDataSources = mergeMap(globalDataSources, serviceReg.SupportedDataSources)

		// Update statistics
		stats.LegacyResources += len(serviceReg.SupportedResources)
		stats.TotalDataSources += len(serviceReg.SupportedDataSources)
		stats.ModernResources += len(serviceReg.Resources)
		stats.TotalDataSources += len(serviceReg.DataSources)
		stats.EphemeralResources += len(serviceReg.EphemeralFunctions)
	}

	stats.TotalResources = stats.LegacyResources + stats.ModernResources + stats.EphemeralResources

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
		totalFiles += len(service.SupportedResources)   // legacy resources
		totalFiles += len(service.Resources)            // modern resources
		totalFiles += len(service.SupportedDataSources) // legacy data sources
		totalFiles += len(service.DataSources)          // modern data sources
		totalFiles += len(service.EphemeralFunctions)   // ephemeral resources
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
		// Process legacy resources
		for terraformType, registrationMethod := range service.SupportedResources {
			// Capture variables for closure
			tfType := terraformType
			regMethod := registrationMethod
			svc := service

			tasks = append(tasks, func() error {
				resourceInfo := NewTerraformResourceInfo(tfType, "", regMethod, "legacy_pluginsdk", svc)
				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(resourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, resourceInfo); err != nil {
					return fmt.Errorf("failed to write legacy resource file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("resource %s", tfType))
				return nil
			})
		}

		// Process modern resources
		for _, structType := range service.Resources {
			// Capture variables for closure
			structT := structType
			svc := service

			tasks = append(tasks, func() error {
				// Get the actual Terraform type from the mapping
				terraformType, exists := svc.ResourceTerraformTypes[structT]
				if !exists {
					// Fallback to struct type if mapping doesn't exist
					terraformType = structT
				}

				resourceInfo := NewTerraformResourceInfo(terraformType, structT, "", "modern_sdk", svc)
				fileName := fmt.Sprintf("%s.json", terraformType)
				filePath := filepath.Join(resourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, resourceInfo); err != nil {
					return fmt.Errorf("failed to write modern resource file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("resource %s", terraformType))
				return nil
			})
		}

		// Process AWS SDK resources (NEW)
		for terraformType, awsResourceInfo := range service.AWSSDKResources {
			// Capture variables for closure
			tfType := terraformType
			awsResource := awsResourceInfo
			svc := service

			tasks = append(tasks, func() error {
				// Create AWS-specific resource info that includes only core TerraformResource fields and essential AWS metadata
				awsResourceData := struct {
					TerraformResource
					FactoryFunction string `json:"factory_function"`
					Name            string `json:"name"`
				}{
					TerraformResource: NewTerraformResourceFromAWSSDK(awsResource, svc),
					FactoryFunction:   awsResource.FactoryFunction,
					Name:              awsResource.Name,
				}

				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(resourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, awsResourceData); err != nil {
					return fmt.Errorf("failed to write AWS SDK resource file %s: %w", fileName, err)
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
		// Process legacy data sources
		for terraformType, registrationMethod := range service.SupportedDataSources {
			// Capture variables for closure
			tfType := terraformType
			regMethod := registrationMethod
			svc := service

			tasks = append(tasks, func() error {
				dataSourceInfo := NewTerraformDataSourceInfo(tfType, "", regMethod, "legacy_pluginsdk", svc)
				fileName := fmt.Sprintf("%s.json", tfType)
				filePath := filepath.Join(dataSourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, dataSourceInfo); err != nil {
					return fmt.Errorf("failed to write legacy data source file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("data source %s", tfType))
				return nil
			})
		}

		// Process modern data sources
		for _, structType := range service.DataSources {
			// Capture variables for closure
			structT := structType
			svc := service

			tasks = append(tasks, func() error {
				// Get the actual Terraform type from the mapping
				terraformType, exists := svc.DataSourceTerraformTypes[structT]
				if !exists {
					// Fallback to struct type if mapping doesn't exist
					terraformType = structT
				}

				dataSourceInfo := NewTerraformDataSourceInfo(terraformType, structT, "", "modern_sdk", svc)
				fileName := fmt.Sprintf("%s.json", terraformType)
				filePath := filepath.Join(dataSourcesDir, fileName)

				if err := index.WriteJSONFile(filePath, dataSourceInfo); err != nil {
					return fmt.Errorf("failed to write modern data source file %s: %w", fileName, err)
				}

				progressTracker.UpdateProgress(fmt.Sprintf("data source %s", terraformType))
				return nil
			})
		}
	}

	return processCallbacksParallel(tasks)
}

// WriteEphemeralFiles writes individual JSON files for each ephemeral resource
func (index *TerraformProviderIndex) WriteEphemeralFiles(outputDir string, progressTracker *ProgressTracker) error {
	ephemeralDir := filepath.Join(outputDir, "ephemeral")
	var tasks []func() error

	for _, service := range index.Services {
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

// processAWSServiceFile extracts all AWS registration methods from a single identified service file
func processAWSServiceFile(fileInfo *gophon.FileInfo, serviceReg *ServiceRegistration) {
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

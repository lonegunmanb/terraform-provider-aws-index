package pkg

import (
	"fmt"
	"strings"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

// identifyServicePackageFile scans a PackageInfo and finds the file containing AWS service registration methods
// Returns the FileInfo containing AWS service methods, or an error if none found
func identifyServicePackageFile(packageInfo *gophon.PackageInfo) (*gophon.FileInfo, error) {
	var candidateFiles []*gophon.FileInfo
	
	// Find all files that contain AWS service methods
	for _, fileInfo := range packageInfo.Files {
		if hasAWSServiceMethods(fileInfo) {
			candidateFiles = append(candidateFiles, fileInfo)
		}
	}
	
	switch len(candidateFiles) {
	case 0:
		return nil, fmt.Errorf("no AWS service methods found in package")
	case 1:
		return candidateFiles[0], nil
	default:
		// Multiple files with AWS methods - use selection logic
		primary := selectPrimaryServiceFile(candidateFiles)
		return primary, nil
	}
}

// hasAWSServiceMethods checks if a FileInfo contains any of the 5 AWS service registration methods
func hasAWSServiceMethods(fileInfo *gophon.FileInfo) bool {
	if fileInfo == nil || fileInfo.File == nil {
		return false
	}
	
	// Use existing AWS extraction functions as detection logic
	// If any extraction function returns non-empty results, this file contains AWS methods
	awsSDKResources := extractAWSSDKResources(fileInfo.File)
	awsSDKDataSources := extractAWSSDKDataSources(fileInfo.File)
	awsFrameworkResources := extractAWSFrameworkResources(fileInfo.File)
	awsFrameworkDataSources := extractAWSFrameworkDataSources(fileInfo.File)
	awsEphemeralResources := extractAWSEphemeralResources(fileInfo.File)
	
	return len(awsSDKResources) > 0 || len(awsSDKDataSources) > 0 || 
		   len(awsFrameworkResources) > 0 || len(awsFrameworkDataSources) > 0 || 
		   len(awsEphemeralResources) > 0
}

// selectPrimaryServiceFile selects the best candidate from multiple files with AWS methods
// Priority logic:
// 1. Files named *service_package* (current AWS convention)
// 2. Files with most AWS methods  
// 3. Alphabetically first file
func selectPrimaryServiceFile(candidates []*gophon.FileInfo) *gophon.FileInfo {
	// Priority 1: Files named *service_package*
	for _, file := range candidates {
		if strings.Contains(file.FileName, "service_package") {
			return file
		}
	}
	
	// Priority 2: File with most AWS methods
	bestFile := candidates[0]
	maxMethods := countAWSMethods(bestFile)
	
	for _, file := range candidates[1:] {
		if count := countAWSMethods(file); count > maxMethods {
			bestFile = file
			maxMethods = count
		}
	}
	
	return bestFile
}

// countAWSMethods counts how many of the 5 AWS service registration methods are present in a file
func countAWSMethods(fileInfo *gophon.FileInfo) int {
	if fileInfo == nil || fileInfo.File == nil {
		return 0
	}
	
	count := 0
	
	// Count each AWS method type
	awsSDKResources := extractAWSSDKResources(fileInfo.File)
	if len(awsSDKResources) > 0 {
		count++
	}
	
	awsSDKDataSources := extractAWSSDKDataSources(fileInfo.File)
	if len(awsSDKDataSources) > 0 {
		count++
	}
	
	awsFrameworkResources := extractAWSFrameworkResources(fileInfo.File)
	if len(awsFrameworkResources) > 0 {
		count++
	}
	
	awsFrameworkDataSources := extractAWSFrameworkDataSources(fileInfo.File)
	if len(awsFrameworkDataSources) > 0 {
		count++
	}
	
	awsEphemeralResources := extractAWSEphemeralResources(fileInfo.File)
	if len(awsEphemeralResources) > 0 {
		count++
	}
	
	return count
}

package pkg

import (
	"fmt"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

// identifyServicePackageFile scans a PackageInfo and finds the file containing AWS service registration methods
// Returns the FileInfo containing AWS service methods, or an error if none found
// Uses simplified single-file assumption approach
func identifyServicePackageFile(packageInfo *gophon.PackageInfo) (*gophon.FileInfo, error) {
	var serviceFile *gophon.FileInfo
	
	// Find the first file that contains AWS service methods
	for _, fileInfo := range packageInfo.Files {
		if hasAWSServiceMethods(fileInfo) {
			if serviceFile != nil {
				// Multiple service files found - this should not happen with simplified assumption
				fmt.Printf("Warning: Multiple AWS service files found in package, using first: %s\n", serviceFile.FileName)
				break
			}
			serviceFile = fileInfo
		}
	}
	
	if serviceFile == nil {
		return nil, fmt.Errorf("no AWS service methods found in package")
	}
	
	fmt.Printf("Found AWS service file: %s\n", serviceFile.FileName)
	return serviceFile, nil
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

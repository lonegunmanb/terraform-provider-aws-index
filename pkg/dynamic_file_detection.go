package pkg

import (
	"fmt"

	gophon "github.com/lonegunmanb/gophon/pkg"
)

// identifyServicePackageFiles scans a PackageInfo and finds ALL files containing AWS service registration methods
// Returns a slice of FileInfo containing AWS service methods, or an error if none found
func identifyServicePackageFiles(packageInfo *gophon.PackageInfo) ([]*gophon.FileInfo, error) {
	var serviceFiles []*gophon.FileInfo
	
	// Find all files that contain AWS service methods
	for _, fileInfo := range packageInfo.Files {
		if hasAWSServiceMethods(fileInfo) {
			serviceFiles = append(serviceFiles, fileInfo)
		}
	}
	
	if len(serviceFiles) == 0 {
		return nil, fmt.Errorf("no AWS service methods found in package")
	}
	
	fmt.Printf("Found %d AWS service files\n", len(serviceFiles))
	return serviceFiles, nil
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

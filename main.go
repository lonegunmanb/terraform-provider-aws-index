package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lonegunmanb/terraform-provider-aws-index/pkg"
)

func main() {
	var (
		scanPath    = flag.String("scan-path", "", "Path to scan for Terraform provider services (required)")
		packagePath = flag.String("package-path", "", "Base package path for the provider (required)")
		version     = flag.String("version", "", "Version of the provider (required)")
		outputDir   = flag.String("output", "./index", "Output directory for index files")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		helpMessage := fmt.Sprintf(`Usage of %s:

Terraform Provider Index Generator
Scans a Terraform provider source directory and generates JSON index files.

Required flags:
  -scan-path string
        Path to scan for Terraform provider services (e.g., ./tmp/terraform-provider-aws/internal/services)
  -package-path string
        Base package path for the provider (e.g., github.com/hashicorp/terraform-provider-aws)
  -version string
        Version of the provider (e.g., v3.116.0)

Optional flags:
  -output string
        Output directory for index files (default "./index")
  -help
        Show this help message

Example:
  %s -scan-path ./tmp/terraform-provider-aws/internal/services \
    -package-path github.com/hashicorp/terraform-provider-aws \
    -version v3.116.0 \
    -output ./output/index
`, os.Args[0], os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "%s", helpMessage)
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate required arguments
	if *scanPath == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: -scan-path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *packagePath == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: -package-path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *version == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Error: -version is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if scan path exists
	if _, err := os.Stat(*scanPath); os.IsNotExist(err) {
		log.Fatalf("Error: scan path does not exist: %s", *scanPath)
	}

	fmt.Printf("🚀 Starting Terraform Provider Indexing...\n")
	fmt.Printf("  📁 Scan Path: %s\n", *scanPath)
	fmt.Printf("  📦 Package Path: %s\n", *packagePath)
	fmt.Printf("  🏷️  Version: %s\n", *version)
	fmt.Printf("  📂 Output Directory: %s\n", *outputDir)
	fmt.Printf("\n")

	// Create progress callback for rich visual feedback
	progressCallback := pkg.CreateRichProgressCallback()

	// Scan the Terraform provider services
	index, err := pkg.ScanTerraformProviderServices(*scanPath, *packagePath, *version, progressCallback)
	if err != nil {
		log.Fatalf("Error scanning Terraform provider services: %v", err)
	}

	fmt.Printf("\n📊 Scan Results:\n")
	fmt.Printf("  🏢 Services Found: %d\n", index.Statistics.ServiceCount)
	fmt.Printf("  📋 Total Resources: %d\n", index.Statistics.TotalResources)
	fmt.Printf("  📄 Total Data Sources: %d\n", index.Statistics.TotalDataSources)
	fmt.Printf("  🔗 Legacy Resources: %d\n", index.Statistics.LegacyResources)
	fmt.Printf("  ⚡ Modern Resources: %d\n", index.Statistics.ModernResources)
	fmt.Printf("  🔄 Ephemeral Resources: %d\n", index.Statistics.EphemeralResources)
	fmt.Printf("\n")

	// Generate JSON output
	err = index.WriteIndexFiles(*outputDir, progressCallback)
	if err != nil {
		log.Fatalf("Error generating JSON output: %v", err)
	}

	fmt.Printf("\n🎉 Index files generated successfully!\n")
	fmt.Printf("  📋 Main index: %s/terraform-provider-aws-index.json\n", *outputDir)
	fmt.Printf("  🔧 Resources: %s/resources/\n", *outputDir)
	fmt.Printf("  📊 Data Sources: %s/datasources/\n", *outputDir)
	fmt.Printf("  ⚡ Ephemeral Resources: %s/ephemeral/\n", *outputDir)
}

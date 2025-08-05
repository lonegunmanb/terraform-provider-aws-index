package pkg

// AnnotationType represents the type of annotation found
type AnnotationType string

const (
	AnnotationSDKResource         AnnotationType = "SDKResource"
	AnnotationSDKDataSource       AnnotationType = "SDKDataSource"
	AnnotationFrameworkResource   AnnotationType = "FrameworkResource"
	AnnotationFrameworkDataSource AnnotationType = "FrameworkDataSource"
	AnnotationEphemeralResource   AnnotationType = "EphemeralResource"
)

// AnnotationResult represents a parsed annotation with its context and extracted info
type AnnotationResult struct {
	Type           AnnotationType `json:"type"`            // The annotation type
	TerraformType  string         `json:"terraform_type"`  // e.g., "aws_key_pair"
	Name           string         `json:"name"`            // e.g., "Key Pair"
	FilePath       string         `json:"file_path"`       // Source file path
	RawAnnotation  string         `json:"raw_annotation"`  // The raw annotation text for debugging
	
	// Extracted information from the file
	StructType     string            `json:"struct_type,omitempty"`     // For framework resources: "guardrailResource"
	CRUDMethods    map[string]string `json:"crud_methods,omitempty"`    // For SDK resources: "create" -> "resourceFunctionCreate"
	FrameworkMethods []string        `json:"framework_methods,omitempty"` // For framework: ["Create", "Read", "Update", "Delete"]
}

// AnnotationResults contains all annotation results found in a package
type AnnotationResults struct {
	SDKResources         []AnnotationResult `json:"sdk_resources"`
	SDKDataSources       []AnnotationResult `json:"sdk_data_sources"`
	FrameworkResources   []AnnotationResult `json:"framework_resources"`
	FrameworkDataSources []AnnotationResult `json:"framework_data_sources"`
	EphemeralResources   []AnnotationResult `json:"ephemeral_resources"`
	
	// Summary statistics
	TotalAnnotations int `json:"total_annotations"`
}

// GetAll returns all annotation results as a single slice
func (ar *AnnotationResults) GetAll() []AnnotationResult {
	var all []AnnotationResult
	all = append(all, ar.SDKResources...)
	all = append(all, ar.SDKDataSources...)
	all = append(all, ar.FrameworkResources...)
	all = append(all, ar.FrameworkDataSources...)
	all = append(all, ar.EphemeralResources...)
	return all
}

// Add adds an annotation result to the appropriate collection
func (ar *AnnotationResults) Add(result AnnotationResult) {
	switch result.Type {
	case AnnotationSDKResource:
		ar.SDKResources = append(ar.SDKResources, result)
	case AnnotationSDKDataSource:
		ar.SDKDataSources = append(ar.SDKDataSources, result)
	case AnnotationFrameworkResource:
		ar.FrameworkResources = append(ar.FrameworkResources, result)
	case AnnotationFrameworkDataSource:
		ar.FrameworkDataSources = append(ar.FrameworkDataSources, result)
	case AnnotationEphemeralResource:
		ar.EphemeralResources = append(ar.EphemeralResources, result)
	}
	ar.TotalAnnotations++
}

// NewAnnotationResults creates a new empty AnnotationResults
func NewAnnotationResults() *AnnotationResults {
	return &AnnotationResults{
		SDKResources:         make([]AnnotationResult, 0),
		SDKDataSources:       make([]AnnotationResult, 0),
		FrameworkResources:   make([]AnnotationResult, 0),
		FrameworkDataSources: make([]AnnotationResult, 0),
		EphemeralResources:   make([]AnnotationResult, 0),
		TotalAnnotations:     0,
	}
}

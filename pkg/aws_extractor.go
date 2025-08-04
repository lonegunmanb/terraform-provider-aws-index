package pkg

import (
	"go/ast"
	"go/token"
	"strings"
)

// AWSResourceInfo represents information about an AWS resource extracted from service package
type AWSResourceInfo struct {
	TerraformType   string             `json:"terraform_type"`
	FactoryFunction string             `json:"factory_function"`
	Name            string             `json:"name"`
	SDKType         string             `json:"sdk_type"` // "sdk", "framework", "ephemeral"
	HasTags         bool               `json:"has_tags"`
	TagsConfig      *AWSTagsConfig     `json:"tags_config,omitempty"`
	Region          *AWSRegionConfig   `json:"region,omitempty"`
	Identity        *AWSIdentityConfig `json:"identity,omitempty"`
	Import          *AWSImportConfig   `json:"import,omitempty"`
}

// AWSTagsConfig represents AWS-specific tags configuration
type AWSTagsConfig struct {
	IdentifierAttribute string `json:"identifier_attribute"`
	ResourceType        string `json:"resource_type"`
}

// AWSRegionConfig represents AWS-specific region configuration
type AWSRegionConfig struct {
	IsOverrideEnabled             bool `json:"is_override_enabled"`
	IsValidateOverrideInPartition bool `json:"is_validate_override_in_partition"`
}

// AWSIdentityConfig represents AWS-specific identity configuration
type AWSIdentityConfig struct {
	IsGlobalResource bool     `json:"is_global_resource"`
	IsSingleton      bool     `json:"is_singleton"`
	IsARN            bool     `json:"is_arn"`
	Attributes       []string `json:"attributes,omitempty"`
}

// AWSImportConfig represents AWS-specific import configuration
type AWSImportConfig struct {
	CustomImport    bool   `json:"custom_import"`
	WrappedImport   bool   `json:"wrapped_import"`
	ImportID        string `json:"import_id,omitempty"`
	ImportStateFunc string `json:"import_state_func,omitempty"`
}

// AWSFactoryCRUDMethods represents CRUD methods extracted from AWS factory functions
type AWSFactoryCRUDMethods struct {
	CreateMethod string `json:"create_method,omitempty"` // "resourceBucketCreate"
	ReadMethod   string `json:"read_method,omitempty"`   // "resourceBucketRead"
	UpdateMethod string `json:"update_method,omitempty"` // "resourceBucketUpdate"
	DeleteMethod string `json:"delete_method,omitempty"` // "resourceBucketDelete"

	// Framework resource methods (for struct-based resources)
	SchemaMethod string `json:"schema_method,omitempty"` // "Schema"

	// Data source specific methods
	ConfigureMethod string `json:"configure_method,omitempty"` // "Configure"

	// Ephemeral resource specific methods
	OpenMethod  string `json:"open_method,omitempty"`  // "Open"
	RenewMethod string `json:"renew_method,omitempty"` // "Renew"
	CloseMethod string `json:"close_method,omitempty"` // "Close"
}

// extractAWSSDKResources extracts SDK resources from the SDKResources method in AWS service packages
func extractAWSSDKResources(node *ast.File) map[string]AWSResourceInfo {
	resources := make(map[string]AWSResourceInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "SDKResources" {
			return true
		}

		// Look for return statements in the function body
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			// Process each return expression
			for _, result := range returnStmt.Results {
				// Handle direct slice literal return
				if sliceLit, ok := result.(*ast.CompositeLit); ok {
					extractedResources := extractAWSSDKResourcesFromSlice(sliceLit)
					for k, v := range extractedResources {
						resources[k] = v
					}
				}

				// Handle variable reference (like "resources" variable)
				ident, ok := result.(*ast.Ident)
				if !ok {
					continue
				}
				// Find the variable definition in the function
				ast.Inspect(fn.Body, func(varNode ast.Node) bool {
					assignStmt, ok := varNode.(*ast.AssignStmt)
					if !ok {
						return true
					}
					for i, lhs := range assignStmt.Lhs {
						lhsIdent, ok := lhs.(*ast.Ident)
						if !ok || lhsIdent.Name != ident.Name {
							return true
						}
						if i >= len(assignStmt.Rhs) {
							return true
						}
						if sliceLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit); ok {
							extractedResources := extractAWSSDKResourcesFromSlice(sliceLit)
							for k, v := range extractedResources {
								resources[k] = v
							}
						}
					}
					return true
				})
			}
			return true
		})

		return true
	})

	return resources
}

// extractAWSSDKResourcesFromSlice extracts AWS SDK resources from a slice literal
func extractAWSSDKResourcesFromSlice(sliceLit *ast.CompositeLit) map[string]AWSResourceInfo {
	resources := make(map[string]AWSResourceInfo)

	for _, elt := range sliceLit.Elts {
		// Handle struct literals like &ServicePackageSDKResource{...}
		var compLit *ast.CompositeLit

		// Check if it's a pointer to struct (&StructName{...})
		if unaryExpr, ok := elt.(*ast.UnaryExpr); ok && unaryExpr.Op == token.AND {
			if cl, ok := unaryExpr.X.(*ast.CompositeLit); ok {
				compLit = cl
			}
		} else if cl, ok := elt.(*ast.CompositeLit); ok {
			// Direct struct literal (StructName{...})
			compLit = cl
		}

		if compLit == nil {
			continue
		}

		resourceInfo := extractAWSResourceInfoFromStruct(compLit)
		if resourceInfo.TerraformType != "" {
			resources[resourceInfo.TerraformType] = resourceInfo
		}
	}

	return resources
}

// extractAWSResourceInfoFromStruct extracts resource information from a ServicePackageSDKResource struct literal
func extractAWSResourceInfoFromStruct(compLit *ast.CompositeLit) AWSResourceInfo {
	resourceInfo := AWSResourceInfo{
		SDKType: "sdk",
	}

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Get the field name
		fieldName := ""
		if ident, ok := kv.Key.(*ast.Ident); ok {
			fieldName = ident.Name
		}

		switch fieldName {
		case "Factory":
			if ident, ok := kv.Value.(*ast.Ident); ok {
				resourceInfo.FactoryFunction = ident.Name
			}
		case "TypeName":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
				resourceInfo.TerraformType = strings.Trim(basicLit.Value, `"`)
			}
		case "Name":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
				resourceInfo.Name = strings.Trim(basicLit.Value, `"`)
			}
		case "Tags":
			tagsConfig := extractAWSTagsConfig(kv.Value)
			if tagsConfig != nil {
				resourceInfo.HasTags = true
				resourceInfo.TagsConfig = tagsConfig
			}
		case "Region":
			regionConfig := extractAWSRegionConfig(kv.Value)
			if regionConfig != nil {
				resourceInfo.Region = regionConfig
			}
		case "Identity":
			identityConfig := extractAWSIdentityConfig(kv.Value)
			if identityConfig != nil {
				resourceInfo.Identity = identityConfig
			}
		case "Import":
			importConfig := extractAWSImportConfig(kv.Value)
			if importConfig != nil {
				resourceInfo.Import = importConfig
			}
		}
	}

	return resourceInfo
}

// extractAWSTagsConfig extracts tags configuration from unique.Make call
func extractAWSTagsConfig(expr ast.Expr) *AWSTagsConfig {
	// Handle unique.Make(inttypes.ServicePackageResourceTags{...})
	if callExpr, ok := expr.(*ast.CallExpr); ok {
		if len(callExpr.Args) > 0 {
			if compLit, ok := callExpr.Args[0].(*ast.CompositeLit); ok {
				tagsConfig := &AWSTagsConfig{}
				for _, elt := range compLit.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						fieldName := ""
						if ident, ok := kv.Key.(*ast.Ident); ok {
							fieldName = ident.Name
						}

						switch fieldName {
						case "IdentifierAttribute":
							// Handle both string literals and names.AttrBucket references
							switch v := kv.Value.(type) {
							case *ast.BasicLit:
								if v.Kind == token.STRING {
									tagsConfig.IdentifierAttribute = strings.Trim(v.Value, `"`)
								}
							case *ast.SelectorExpr:
								// Handle names.AttrBucket -> "bucket"
								if x, ok := v.X.(*ast.Ident); ok && x.Name == "names" {
									attrName := v.Sel.Name
									// Convert AttrBucket to "bucket"
									if strings.HasPrefix(attrName, "Attr") {
										tagsConfig.IdentifierAttribute = strings.ToLower(attrName[4:])
									}
								}
							}
						case "ResourceType":
							if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
								tagsConfig.ResourceType = strings.Trim(basicLit.Value, `"`)
							}
						}
					}
				}
				return tagsConfig
			}
		}
	}
	return nil
}

// extractAWSRegionConfig extracts region configuration from unique.Make call
func extractAWSRegionConfig(expr ast.Expr) *AWSRegionConfig {
	// Handle unique.Make(inttypes.ResourceRegionDefault()) or similar
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok || len(callExpr.Args) == 0 {
		return nil
	}

	// Check for ResourceRegionDefault() call
	innerCall, ok := callExpr.Args[0].(*ast.CallExpr)
	if !ok {
		return nil
	}

	selectorExpr, ok := innerCall.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	x, ok := selectorExpr.X.(*ast.Ident)
	if !ok || x.Name != "inttypes" {
		return nil
	}

	switch selectorExpr.Sel.Name {
	case "ResourceRegionDefault":
		return &AWSRegionConfig{
			IsOverrideEnabled:             true,
			IsValidateOverrideInPartition: true,
		}
	case "ResourceRegionDisabled":
		return &AWSRegionConfig{
			IsOverrideEnabled:             false,
			IsValidateOverrideInPartition: false,
		}
	}

	return nil
}

// extractAWSIdentityConfig extracts identity configuration (placeholder for future implementation)
func extractAWSIdentityConfig(expr ast.Expr) *AWSIdentityConfig {
	// TODO: Implement when we encounter actual identity configurations in tests
	return nil
}

// extractAWSImportConfig extracts import configuration (placeholder for future implementation)
func extractAWSImportConfig(expr ast.Expr) *AWSImportConfig {
	// TODO: Implement when we encounter actual import configurations in tests
	return nil
}

// extractAWSSDKDataSources extracts SDK data sources from the SDKDataSources method in AWS service packages
func extractAWSSDKDataSources(node *ast.File) map[string]AWSResourceInfo {
	dataSources := make(map[string]AWSResourceInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "SDKDataSources" {
			return true
		}

		// Look for return statements in the function body
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			returnStmt, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			// Process each return expression
			for _, result := range returnStmt.Results {
				// Handle direct slice literal return
				if sliceLit, ok := result.(*ast.CompositeLit); ok {
					extractedDataSources := extractAWSSDKDataSourcesFromSlice(sliceLit)
					for k, v := range extractedDataSources {
						dataSources[k] = v
					}
					continue
				}

				// Handle variable reference (like "dataSources" variable)
				if ident, ok := result.(*ast.Ident); ok {
					extractedFromVariable := extractFromVariableReference(fn.Body, ident.Name, extractAWSSDKDataSourcesFromSlice)
					for k, v := range extractedFromVariable {
						dataSources[k] = v
					}
				}
			}
			return true
		})

		return true
	})

	return dataSources
}

// extractAWSSDKDataSourcesFromSlice extracts AWS SDK data sources from a slice literal
func extractAWSSDKDataSourcesFromSlice(sliceLit *ast.CompositeLit) map[string]AWSResourceInfo {
	dataSources := make(map[string]AWSResourceInfo)

	for _, elt := range sliceLit.Elts {
		// Handle struct literals like &ServicePackageSDKDataSource{...}
		var compLit *ast.CompositeLit

		// Extract composite literal from different patterns
		switch e := elt.(type) {
		case *ast.UnaryExpr:
			// Check if it's a pointer to struct (&StructName{...})
			cl, ok := e.X.(*ast.CompositeLit)
			if e.Op == token.AND && ok {
				compLit = cl
			}
		case *ast.CompositeLit:
			// Direct struct literal (StructName{...})
			compLit = e
		}

		if compLit == nil {
			continue
		}

		dataSourceInfo := extractAWSDataSourceInfoFromStruct(compLit)
		if dataSourceInfo.TerraformType != "" {
			dataSources[dataSourceInfo.TerraformType] = dataSourceInfo
		}
	}

	return dataSources
}

// extractAWSDataSourceInfoFromStruct extracts data source information from a ServicePackageSDKDataSource struct literal
func extractAWSDataSourceInfoFromStruct(compLit *ast.CompositeLit) AWSResourceInfo {
	dataSourceInfo := AWSResourceInfo{
		SDKType: "sdk",
	}

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Get the field name
		fieldName := ""
		if ident, ok := kv.Key.(*ast.Ident); ok {
			fieldName = ident.Name
		}

		switch fieldName {
		case "Factory":
			if ident, ok := kv.Value.(*ast.Ident); ok {
				dataSourceInfo.FactoryFunction = ident.Name
			}
		case "TypeName":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
				dataSourceInfo.TerraformType = strings.Trim(basicLit.Value, `"`)
			}
		case "Name":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
				dataSourceInfo.Name = strings.Trim(basicLit.Value, `"`)
			}
		case "Tags":
			tagsConfig := extractAWSTagsConfig(kv.Value)
			if tagsConfig != nil {
				dataSourceInfo.HasTags = true
				dataSourceInfo.TagsConfig = tagsConfig
			}
		case "Region":
			regionConfig := extractAWSRegionConfig(kv.Value)
			if regionConfig != nil {
				dataSourceInfo.Region = regionConfig
			}
		case "Identity":
			identityConfig := extractAWSIdentityConfig(kv.Value)
			if identityConfig != nil {
				dataSourceInfo.Identity = identityConfig
			}
		case "Import":
			importConfig := extractAWSImportConfig(kv.Value)
			if importConfig != nil {
				dataSourceInfo.Import = importConfig
			}
		}
	}

	return dataSourceInfo
}

// extractAWSFrameworkResources extracts Framework resources from the FrameworkResources method in AWS service packages
func extractAWSFrameworkResources(node *ast.File) map[string]AWSResourceInfo {
	resources := make(map[string]AWSResourceInfo)

	// Find the FrameworkResources method
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "FrameworkResources" {
			continue
		}

		// Look for return statements in the function body
		for _, stmt := range funcDecl.Body.List {
			switch s := stmt.(type) {
			case *ast.ReturnStmt:
				// Handle direct return
				if len(s.Results) == 0 {
					continue
				}
				if sliceLit, ok := s.Results[0].(*ast.CompositeLit); ok {
					extractedResources := extractAWSFrameworkResourcesFromSlice(sliceLit)
					for k, v := range extractedResources {
						resources[k] = v
					}
				}
			case *ast.AssignStmt:
				// Handle variable assignment pattern: resources := []*inttypes.ServicePackageFrameworkResource{...}
				if len(s.Rhs) == 0 {
					continue
				}
				if sliceLit, ok := s.Rhs[0].(*ast.CompositeLit); ok {
					extractedResources := extractAWSFrameworkResourcesFromSlice(sliceLit)
					for k, v := range extractedResources {
						resources[k] = v
					}
				}
			case *ast.DeclStmt:
				// Handle variable declaration: var resources = []*inttypes.ServicePackageFrameworkResource{...}
				genDecl, ok := s.Decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.VAR {
					continue
				}
				for _, spec := range genDecl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok && len(valueSpec.Values) == 0 {
						continue
					}
					sliceLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
					if !ok {
						continue
					}
					extractedResources := extractAWSFrameworkResourcesFromSlice(sliceLit)
					for k, v := range extractedResources {
						resources[k] = v
					}
				}
			}
		}

		break
	}

	return resources
}

// extractAWSFrameworkResourcesFromSlice extracts AWS Framework resources from a slice literal
func extractAWSFrameworkResourcesFromSlice(sliceLit *ast.CompositeLit) map[string]AWSResourceInfo {
	resources := make(map[string]AWSResourceInfo)

	for _, elt := range sliceLit.Elts {
		structLit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		resourceInfo := extractAWSFrameworkResourceInfo(structLit)
		if resourceInfo.TerraformType != "" {
			resources[resourceInfo.TerraformType] = resourceInfo
		}
	}

	return resources
}

// extractAWSFrameworkResourceInfo extracts individual AWS Framework resource information from a struct literal
func extractAWSFrameworkResourceInfo(structLit *ast.CompositeLit) AWSResourceInfo {
	resourceInfo := AWSResourceInfo{
		SDKType: "framework",
		HasTags: false,
	}

	for _, elt := range structLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch key.Name {
		case "Factory":
			if ident, ok := kv.Value.(*ast.Ident); ok {
				resourceInfo.FactoryFunction = ident.Name
			}
		case "TypeName":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok {
				terraformType := strings.Trim(basicLit.Value, `"`)
				resourceInfo.TerraformType = terraformType
			}
		case "Name":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok {
				name := strings.Trim(basicLit.Value, `"`)
				resourceInfo.Name = name
			}
		case "Tags":
			// Check if Tags field exists and is not nil
			if kv.Value == nil {
				continue
			}
			resourceInfo.HasTags = true
			tagsConfig := extractAWSTagsConfig(kv.Value)
			if tagsConfig != nil {
				resourceInfo.TagsConfig = tagsConfig
			}
		case "Region":
			regionConfig := extractAWSRegionConfig(kv.Value)
			if regionConfig != nil {
				resourceInfo.Region = regionConfig
			}
		case "Identity":
			identityConfig := extractAWSIdentityConfig(kv.Value)
			if identityConfig != nil {
				resourceInfo.Identity = identityConfig
			}
		case "Import":
			importConfig := extractAWSImportConfig(kv.Value)
			if importConfig != nil {
				resourceInfo.Import = importConfig
			}
		}
	}

	return resourceInfo
}

// extractAWSFrameworkDataSources extracts Framework data sources from the FrameworkDataSources method in AWS service packages
func extractAWSFrameworkDataSources(node *ast.File) map[string]AWSResourceInfo {
	dataSources := make(map[string]AWSResourceInfo)

	// Find the FrameworkDataSources method
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "FrameworkDataSources" {
			continue
		}

		// Look for return statements in the function body
		for _, stmt := range funcDecl.Body.List {
			switch s := stmt.(type) {
			case *ast.ReturnStmt:
				// Handle direct return
				if len(s.Results) == 0 {
					continue
				}
				if sliceLit, ok := s.Results[0].(*ast.CompositeLit); ok {
					extractedDataSources := extractAWSFrameworkDataSourcesFromSlice(sliceLit)
					for k, v := range extractedDataSources {
						dataSources[k] = v
					}
				}
			case *ast.AssignStmt:
				// Handle variable assignment pattern: dataSources := []*inttypes.ServicePackageFrameworkDataSource{...}
				if len(s.Rhs) == 0 {
					continue
				}
				if sliceLit, ok := s.Rhs[0].(*ast.CompositeLit); ok {
					extractedDataSources := extractAWSFrameworkDataSourcesFromSlice(sliceLit)
					for k, v := range extractedDataSources {
						dataSources[k] = v
					}
				}
			case *ast.DeclStmt:
				// Handle variable declaration: var dataSources = []*inttypes.ServicePackageFrameworkDataSource{...}
				genDecl, ok := s.Decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.VAR {
					continue
				}
				for _, spec := range genDecl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok || len(valueSpec.Values) == 0 {
						continue
					}
					sliceLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
					if !ok {
						continue
					}
					extractedDataSources := extractAWSFrameworkDataSourcesFromSlice(sliceLit)
					for k, v := range extractedDataSources {
						dataSources[k] = v
					}
				}
			}
		}
		break
	}

	return dataSources
}

// extractAWSFrameworkDataSourcesFromSlice extracts AWS Framework data sources from a slice literal
func extractAWSFrameworkDataSourcesFromSlice(sliceLit *ast.CompositeLit) map[string]AWSResourceInfo {
	dataSources := make(map[string]AWSResourceInfo)

	for _, elt := range sliceLit.Elts {
		structLit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		dataSourceInfo := extractAWSFrameworkDataSourceInfo(structLit)
		if dataSourceInfo.TerraformType != "" {
			dataSources[dataSourceInfo.TerraformType] = dataSourceInfo
		}
	}

	return dataSources
}

// extractAWSFrameworkDataSourceInfo extracts individual AWS Framework data source information from a struct literal
func extractAWSFrameworkDataSourceInfo(structLit *ast.CompositeLit) AWSResourceInfo {
	dataSourceInfo := AWSResourceInfo{
		SDKType: "framework",
		HasTags: false,
	}

	for _, elt := range structLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch key.Name {
		case "Factory":
			if ident, ok := kv.Value.(*ast.Ident); ok {
				dataSourceInfo.FactoryFunction = ident.Name
			}
		case "TypeName":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok {
				terraformType := strings.Trim(basicLit.Value, `"`)
				dataSourceInfo.TerraformType = terraformType
			}
		case "Name":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok {
				name := strings.Trim(basicLit.Value, `"`)
				dataSourceInfo.Name = name
			}
		case "Tags":
			// Check if Tags field exists and is not nil
			if kv.Value != nil {
				dataSourceInfo.HasTags = true
				tagsConfig := extractAWSTagsConfig(kv.Value)
				if tagsConfig != nil {
					dataSourceInfo.TagsConfig = tagsConfig
				}
			}
		case "Region":
			regionConfig := extractAWSRegionConfig(kv.Value)
			if regionConfig != nil {
				dataSourceInfo.Region = regionConfig
			}
		case "Identity":
			identityConfig := extractAWSIdentityConfig(kv.Value)
			if identityConfig != nil {
				dataSourceInfo.Identity = identityConfig
			}
		case "Import":
			importConfig := extractAWSImportConfig(kv.Value)
			if importConfig != nil {
				dataSourceInfo.Import = importConfig
			}
		}
	}

	return dataSourceInfo
}

// extractAWSEphemeralResources extracts ephemeral resources from the EphemeralResources method in AWS service packages
func extractAWSEphemeralResources(node *ast.File) map[string]AWSResourceInfo {
	resources := make(map[string]AWSResourceInfo)

	// Find the EphemeralResources method
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "EphemeralResources" {
			continue
		}

		// Look for return statements in the function body
		for _, stmt := range funcDecl.Body.List {
			switch s := stmt.(type) {
			case *ast.ReturnStmt:
				// Handle direct return
				if len(s.Results) == 0 {
					continue
				}
				sliceLit, ok := s.Results[0].(*ast.CompositeLit)
				if !ok {
					continue
				}
				extractedResources := extractAWSEphemeralResourcesFromSlice(sliceLit)
				for k, v := range extractedResources {
					resources[k] = v

				}
			case *ast.AssignStmt:
				// Handle variable assignment pattern: resources := []*inttypes.ServicePackageEphemeralResource{...}
				if len(s.Rhs) == 0 {
					continue
				}
				sliceLit, ok := s.Rhs[0].(*ast.CompositeLit)
				if !ok {
					continue
				}
				extractedResources := extractAWSEphemeralResourcesFromSlice(sliceLit)
				for k, v := range extractedResources {
					resources[k] = v
				}

			case *ast.DeclStmt:
				// Handle variable declaration: var resources = []*inttypes.ServicePackageEphemeralResource{...}
				genDecl, ok := s.Decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.VAR {
					continue
				}
				for _, spec := range genDecl.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok || len(valueSpec.Values) == 0 {
						continue
					}
					sliceLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
					if !ok {
						continue
					}
					extractedResources := extractAWSEphemeralResourcesFromSlice(sliceLit)
					for k, v := range extractedResources {
						resources[k] = v
					}
				}
			}
		}
		break
	}

	return resources
}

// extractAWSEphemeralResourcesFromSlice extracts AWS ephemeral resources from a slice literal
func extractAWSEphemeralResourcesFromSlice(sliceLit *ast.CompositeLit) map[string]AWSResourceInfo {
	resources := make(map[string]AWSResourceInfo)

	for _, elt := range sliceLit.Elts {
		structLit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		resourceInfo := extractAWSEphemeralResourceInfo(structLit)
		if resourceInfo.TerraformType != "" {
			resources[resourceInfo.TerraformType] = resourceInfo
		}
	}

	return resources
}

// extractAWSEphemeralResourceInfo extracts individual AWS ephemeral resource information from a struct literal
func extractAWSEphemeralResourceInfo(structLit *ast.CompositeLit) AWSResourceInfo {
	resourceInfo := AWSResourceInfo{
		SDKType: "ephemeral",
		HasTags: false,
	}

	for _, elt := range structLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch key.Name {
		case "Factory":
			if ident, ok := kv.Value.(*ast.Ident); ok {
				resourceInfo.FactoryFunction = ident.Name
			}
		case "TypeName":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok {
				terraformType := strings.Trim(basicLit.Value, `"`)
				resourceInfo.TerraformType = terraformType
			}
		case "Name":
			if basicLit, ok := kv.Value.(*ast.BasicLit); ok {
				name := strings.Trim(basicLit.Value, `"`)
				resourceInfo.Name = name
			}
		case "Tags":
			// Check if Tags field exists and is not nil
			if kv.Value != nil {
				resourceInfo.HasTags = true
				tagsConfig := extractAWSTagsConfig(kv.Value)
				if tagsConfig != nil {
					resourceInfo.TagsConfig = tagsConfig
				}
			}
		case "Region":
			regionConfig := extractAWSRegionConfig(kv.Value)
			if regionConfig != nil {
				resourceInfo.Region = regionConfig
			}
		case "Identity":
			identityConfig := extractAWSIdentityConfig(kv.Value)
			if identityConfig != nil {
				resourceInfo.Identity = identityConfig
			}
		case "Import":
			importConfig := extractAWSImportConfig(kv.Value)
			if importConfig != nil {
				resourceInfo.Import = importConfig
			}
		}
	}

	return resourceInfo
}

// Helper function placeholders - these will be implemented in aws_extractor.go
func extractFactoryFunctionDetails(node *ast.File, factoryFunctionName string) *AWSFactoryCRUDMethods {
	// Find the factory function by name
	factoryFunc := findFactoryFunction(node, factoryFunctionName)
	if factoryFunc == nil {
		// Function not found - return empty struct
		return &AWSFactoryCRUDMethods{}
	}

	// Initialize result
	result := &AWSFactoryCRUDMethods{}

	// Try to extract from direct return statements first (SDK pattern)
	returnStmts := findReturnStatements(factoryFunc)
	for _, returnStmt := range returnStmts {
		for _, expr := range returnStmt.Results {
			switch e := expr.(type) {
			case *ast.UnaryExpr:
				// Handle &schema.Resource{...} pattern (UnaryExpr with & operator)
				compositeLit, ok := e.X.(*ast.CompositeLit)
				if ok && e.Op == token.AND {
					extractSDKCRUDFromCompositeLit(compositeLit, result)
				}
			case *ast.CompositeLit:
				// Direct return pattern: return schema.Resource{...} (without &)
				extractSDKCRUDFromCompositeLit(e, result)
			case *ast.Ident:
				// Variable reference pattern: return resource
				extractFromVariableAssignment(factoryFunc, e.Name, result)
			}
		}
	}

	// Try Framework pattern extraction if no SDK methods found
	if isEmptyResult(result) {
		structType := findFrameworkStructType(factoryFunc)
		if structType != "" {
			extractFrameworkMethods(node, structType, result)
		}
	}

	return result
}

// Helper function to find a factory function by name in the AST
func findFactoryFunction(node *ast.File, functionName string) *ast.FuncDecl {
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if funcDecl.Name.Name == functionName {
			return funcDecl
		}
	}
	return nil
}

// Helper function to navigate AST and find return statements in a function
func findReturnStatements(funcDecl *ast.FuncDecl) []*ast.ReturnStmt {
	var returns []*ast.ReturnStmt

	if funcDecl.Body == nil {
		return returns
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			returns = append(returns, ret)
		}
		return true
	})

	return returns
}

// Helper function to find variable assignments in a function body
func findVariableAssignments(funcDecl *ast.FuncDecl) []*ast.AssignStmt {
	var assignments []*ast.AssignStmt

	if funcDecl.Body == nil {
		return assignments
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			assignments = append(assignments, assign)
		}
		return true
	})

	return assignments
}

// extractFromVariableAssignment extracts CRUD methods from variable assignments
func extractFromVariableAssignment(factoryFunc *ast.FuncDecl, variableName string, result *AWSFactoryCRUDMethods) {
	assignments := findVariableAssignments(factoryFunc)
	for _, assign := range assignments {
		// Look for assignments where the left side matches our identifier
		for i, lhs := range assign.Lhs {
			lhsIdent, ok := lhs.(*ast.Ident)
			if !ok || lhsIdent.Name != variableName {
				continue
			}

			// Found the assignment, check the right hand side
			if i >= len(assign.Rhs) {
				continue
			}

			switch rhs := assign.Rhs[i].(type) {
			case *ast.UnaryExpr:
				// Handle &schema.Resource{...} pattern in assignment
				if compositeLit, ok := rhs.X.(*ast.CompositeLit); ok && rhs.Op == token.AND {
					extractSDKCRUDFromCompositeLit(compositeLit, result)
				}
			case *ast.CompositeLit:
				extractSDKCRUDFromCompositeLit(rhs, result)
			}
		}
	}
}

// extractSDKCRUDFromCompositeLit extracts CRUD method names from a schema.Resource composite literal
func extractSDKCRUDFromCompositeLit(compositeLit *ast.CompositeLit, result *AWSFactoryCRUDMethods) {
	// Look through the elements of the composite literal
	for _, elt := range compositeLit.Elts {
		if keyValue, ok := elt.(*ast.KeyValueExpr); ok {
			// Get the field name
			var fieldName string
			if ident, ok := keyValue.Key.(*ast.Ident); ok {
				fieldName = ident.Name
			}

			// Get the function name from the value
			var functionName string
			if ident, ok := keyValue.Value.(*ast.Ident); ok {
				functionName = ident.Name
			}

			// Map SDK field names to CRUD methods
			switch fieldName {
			// Create method variants
			case "Create", "CreateWithoutTimeout", "CreateContext":
				result.CreateMethod = functionName
			// Read method variants
			case "Read", "ReadWithoutTimeout", "ReadContext":
				result.ReadMethod = functionName
			// Update method variants
			case "Update", "UpdateWithoutTimeout", "UpdateContext":
				result.UpdateMethod = functionName
			// Delete method variants
			case "Delete", "DeleteWithoutTimeout", "DeleteContext":
				result.DeleteMethod = functionName
			}
		}
	}
}

// isEmptyResult checks if the CRUD methods result is empty (no SDK methods found)
func isEmptyResult(result *AWSFactoryCRUDMethods) bool {
	return result.CreateMethod == "" && result.ReadMethod == "" &&
		result.UpdateMethod == "" && result.DeleteMethod == ""
}

// findFrameworkStructType extracts the struct type name from Framework factory function
func findFrameworkStructType(factoryFunc *ast.FuncDecl) string {
	if factoryFunc.Body == nil {
		return ""
	}

	var structType string

	// Look for assignment patterns like: r := &structName{}
	ast.Inspect(factoryFunc.Body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, rhs := range assign.Rhs {
				// Look for &structName{} pattern
				unaryExpr, ok := rhs.(*ast.UnaryExpr)
				if !ok || unaryExpr.Op != token.AND {
					continue
				}
				compositeLit, ok := unaryExpr.X.(*ast.CompositeLit)
				if !ok {
					continue
				}
				if ident, ok := compositeLit.Type.(*ast.Ident); ok {
					structType = ident.Name
					return false // Stop inspection, we found it
				}
			}
		}
		return true
	})

	return structType
}

// extractFrameworkMethods finds method implementations on a struct type
func extractFrameworkMethods(node *ast.File, structTypeName string, result *AWSFactoryCRUDMethods) {
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for method declarations
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			return true
		}
		// Check if this method belongs to our struct type
		for _, field := range funcDecl.Recv.List {
			// Handle receiver patterns like (r *structName)
			starExpr, ok := field.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			if ident, ok := starExpr.X.(*ast.Ident); ok && ident.Name == structTypeName {
				methodName := funcDecl.Name.Name
				mapFrameworkMethod(methodName, result)
			}
			// Handle receiver patterns like (r structName)
			if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == structTypeName {
				methodName := funcDecl.Name.Name
				mapFrameworkMethod(methodName, result)
			}
		}
		return true
	})
}

// mapFrameworkMethod maps Framework method names to CRUD result fields
func mapFrameworkMethod(methodName string, result *AWSFactoryCRUDMethods) {
	switch methodName {
	case "Schema":
		result.SchemaMethod = methodName
	case "Create":
		result.CreateMethod = methodName
	case "Read":
		result.ReadMethod = methodName
	case "Update":
		result.UpdateMethod = methodName
	case "Delete":
		result.DeleteMethod = methodName
	case "Configure":
		result.ConfigureMethod = methodName
	case "Open":
		result.OpenMethod = methodName
	case "Renew":
		result.RenewMethod = methodName
	case "Close":
		result.CloseMethod = methodName
	}
}

// extractFromVariableReference is a generic helper function that extracts resources/data sources
// from variable references in return statements. It finds the variable definition in the function
// body and applies the provided extraction function to any slice literals found.
func extractFromVariableReference(fnBody *ast.BlockStmt, variableName string, extractFunc func(*ast.CompositeLit) map[string]AWSResourceInfo) map[string]AWSResourceInfo {
	result := make(map[string]AWSResourceInfo)

	// Find the variable definition in the function
	ast.Inspect(fnBody, func(varNode ast.Node) bool {
		assignStmt, ok := varNode.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for i, lhs := range assignStmt.Lhs {
			lhsIdent, ok := lhs.(*ast.Ident)
			if !ok || lhsIdent.Name != variableName {
				continue
			}

			if i >= len(assignStmt.Rhs) {
				continue
			}
			sliceLit, ok := assignStmt.Rhs[i].(*ast.CompositeLit)
			if !ok {
				continue
			}
			extracted := extractFunc(sliceLit)
			for k, v := range extracted {
				result[k] = v
			}
		}
		return true
	})

	return result
}

package pkg

import (
	gophon "github.com/lonegunmanb/gophon/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestExtractCRUDFromPackage_ManagedApplicationDefinition(t *testing.T) {
	// Test case with the exact function you provided from the Azure provider
	source := `package managedapplications

func resourceManagedApplicationDefinition() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceManagedApplicationDefinitionCreate,
		Read:   resourceManagedApplicationDefinitionRead,
		Update: resourceManagedApplicationDefinitionUpdate,
		Delete: resourceManagedApplicationDefinitionDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := applicationdefinitions.ParseApplicationDefinitionID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApplicationDefinitionName,
			},

			"resource_group_name": commonschema.ResourceGroupName(),

			"location": commonschema.Location(),

			"display_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.ApplicationDefinitionDisplayName,
			},

			"lock_level": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(applicationdefinitions.ApplicationLockLevelCanNotDelete),
					string(applicationdefinitions.ApplicationLockLevelNone),
					string(applicationdefinitions.ApplicationLockLevelReadOnly),
				}, false),
			},

			"authorization": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"role_definition_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
						"service_principal_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},
					},
				},
			},

			"create_ui_definition": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
				ConflictsWith:    []string{"package_file_uri"},
				RequiredWith:     []string{"main_template"},
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validate.ApplicationDefinitionDescription,
			},

			"main_template": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: pluginsdk.SuppressJsonDiff,
				ConflictsWith:    []string{"package_file_uri"},
				RequiredWith:     []string{"create_ui_definition"},
			},

			"package_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  true,
			},

			"package_file_uri": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},

			"tags": commonschema.Tags(),
		},
	}
}`

	// Expected result - should extract the CRUD method names
	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "resourceManagedApplicationDefinitionCreate",
		ReadMethod:   "resourceManagedApplicationDefinitionRead",
		UpdateMethod: "resourceManagedApplicationDefinitionUpdate",
		DeleteMethod: "resourceManagedApplicationDefinitionDelete",
	}

	// Parse the source code into AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	// Find the function declaration from the parsed file
	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "resourceManagedApplicationDefinition" {
			funcDecl = fn
			return false // Stop searching
		}
		return true
	})
	require.NotNil(t, funcDecl, "Function declaration should be found")

	// Create mock package info with Functions field populated
	packageInfo := &gophon.PackageInfo{
		Functions: []*gophon.FunctionInfo{
			{
				Name:     "resourceManagedApplicationDefinition",
				FuncDecl: funcDecl,
			},
		},
	}

	// Test the extractCRUDFromPackage function
	result := extractCRUDFromPackage("resourceManagedApplicationDefinition", packageInfo)

	// Verify the result
	require.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, expected, result)
}

func TestExtractCRUDFromPackage_NotFound(t *testing.T) {
	// Test case where the registration method is not found
	source := `package test

func someOtherFunction() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: someCreateFunc,
		Read:   someReadFunc,
	}
}`

	// Parse the source code into AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	// Find the function declaration
	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "someOtherFunction" {
			funcDecl = fn
			return false
		}
		return true
	})
	require.NotNil(t, funcDecl)

	// Create mock package info with different function name
	packageInfo := &gophon.PackageInfo{
		Functions: []*gophon.FunctionInfo{
			{
				Name:     "someOtherFunction",
				FuncDecl: funcDecl,
			},
		},
	}

	// Test with non-matching registration method
	result := extractCRUDFromPackage("resourceManagedApplicationDefinition", packageInfo)

	// Should return nil since the function name doesn't match
	assert.Nil(t, result)
}

func TestExtractCRUDFromPackage_NilPackageInfo(t *testing.T) {
	// Test case with nil package info
	result := extractCRUDFromPackage("resourceManagedApplicationDefinition", nil)
	assert.Nil(t, result)
}

func TestExtractCRUDFromPackage_EmptyFunctions(t *testing.T) {
	// Test case with empty Functions slice
	packageInfo := &gophon.PackageInfo{
		Functions: []*gophon.FunctionInfo{},
	}

	result := extractCRUDFromPackage("resourceManagedApplicationDefinition", packageInfo)
	assert.Nil(t, result)
}

func TestExtractCRUDFromPackage_NilFuncDecl(t *testing.T) {
	// Test case where FuncDecl is nil
	packageInfo := &gophon.PackageInfo{
		Functions: []*gophon.FunctionInfo{
			{
				Name:     "resourceManagedApplicationDefinition",
				FuncDecl: nil, // This should cause the function to return nil
			},
		},
	}

	result := extractCRUDFromPackage("resourceManagedApplicationDefinition", packageInfo)
	assert.Nil(t, result)
}

func TestExtractCRUDFromPackage_ContextVariants(t *testing.T) {
	// Test case with CreateContext, ReadContext, etc. variants
	source := `package test

func resourceWithContextMethods() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		CreateContext: resourceWithContextMethodsCreateContext,
		ReadContext:   resourceWithContextMethodsReadContext,
		UpdateContext: resourceWithContextMethodsUpdateContext,
		DeleteContext: resourceWithContextMethodsDeleteContext,
		
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "resourceWithContextMethodsCreateContext",
		ReadMethod:   "resourceWithContextMethodsReadContext",
		UpdateMethod: "resourceWithContextMethodsUpdateContext",
		DeleteMethod: "resourceWithContextMethodsDeleteContext",
	}

	// Parse and create package info
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "resourceWithContextMethods" {
			funcDecl = fn
			return false
		}
		return true
	})
	require.NotNil(t, funcDecl)

	packageInfo := &gophon.PackageInfo{
		Functions: []*gophon.FunctionInfo{
			{
				Name:     "resourceWithContextMethods",
				FuncDecl: funcDecl,
			},
		},
	}

	result := extractCRUDFromPackage("resourceWithContextMethods", packageInfo)
	require.NotNil(t, result)
	assert.Equal(t, expected, result)
}

func TestExtractCRUDFromPackage_PartialMethods(t *testing.T) {
	// Test case where only some CRUD methods are present
	source := `package test

func resourcePartialMethods() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePartialMethodsCreate,
		Read:   resourcePartialMethodsRead,
		// No Update or Delete methods
		
		Schema: map[string]*pluginsdk.Schema{
			"name": {Type: pluginsdk.TypeString, Required: true},
		},
	}
}`

	expected := &LegacyResourceCRUDFunctions{
		CreateMethod: "resourcePartialMethodsCreate",
		ReadMethod:   "resourcePartialMethodsRead",
		UpdateMethod: "", // Should be empty
		DeleteMethod: "", // Should be empty
	}

	// Parse and create package info
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "resourcePartialMethods" {
			funcDecl = fn
			return false
		}
		return true
	})
	require.NotNil(t, funcDecl)

	packageInfo := &gophon.PackageInfo{
		Functions: []*gophon.FunctionInfo{
			{
				Name:     "resourcePartialMethods",
				FuncDecl: funcDecl,
			},
		},
	}

	result := extractCRUDFromPackage("resourcePartialMethods", packageInfo)
	require.NotNil(t, result)
	assert.Equal(t, expected, result)
}

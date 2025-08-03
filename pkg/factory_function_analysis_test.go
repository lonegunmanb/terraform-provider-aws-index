package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExtractFactoryFunctionDetails_SDKResource(t *testing.T) {
	t.Run("extract CRUD methods from SDK resource factory", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"time"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,
		
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation here
	return nil
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation here
	return nil
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation here
	return nil
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation here
	return nil
}`

		expected := &AWSFactoryCRUDMethods{
			CreateMethod: "resourceBucketCreate",
			ReadMethod:   "resourceBucketRead",
			UpdateMethod: "resourceBucketUpdate",
			DeleteMethod: "resourceBucketDelete",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceBucket")
		assert.Equal(t, expected, result)
	})

	t.Run("extract CRUD methods with legacy field names", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceExample() *schema.Resource {
	return &schema.Resource{
		Create: resourceExampleCreate,
		Read:   resourceExampleRead,
		Update: resourceExampleUpdate,
		Delete: resourceExampleDelete,
		
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
		},
	}
}`

		expected := &AWSFactoryCRUDMethods{
			CreateMethod: "resourceExampleCreate",
			ReadMethod:   "resourceExampleRead",
			UpdateMethod: "resourceExampleUpdate",
			DeleteMethod: "resourceExampleDelete",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceExample")
		assert.Equal(t, expected, result)
	})

	t.Run("extract partial CRUD methods", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceReadOnly() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReadOnlyCreate,
		ReadWithoutTimeout:   resourceReadOnlyRead,
		// No Update or Delete methods
		
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Computed: true},
		},
	}
}`

		expected := &AWSFactoryCRUDMethods{
			CreateMethod: "resourceReadOnlyCreate",
			ReadMethod:   "resourceReadOnlyRead",
			UpdateMethod: "", // Should be empty
			DeleteMethod: "", // Should be empty
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceReadOnly")
		assert.Equal(t, expected, result)
	})

	t.Run("factory function not found", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func someOtherFunction() string {
	return "not a factory"
}`

		expected := &AWSFactoryCRUDMethods{}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceNonExistent")
		assert.Equal(t, expected, result)
	})
}

func TestExtractFactoryFunctionDetails_SDKDataSource(t *testing.T) {
	t.Run("extract read method from SDK data source factory", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBucket() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBucketRead,
		
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Implementation here
	return nil
}`

		expected := &AWSFactoryCRUDMethods{
			ReadMethod: "dataSourceBucketRead",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "dataSourceBucket")
		assert.Equal(t, expected, result)
	})

	t.Run("extract read method with legacy field name", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExample() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceExampleRead,
		
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
		},
	}
}`

		expected := &AWSFactoryCRUDMethods{
			ReadMethod: "dataSourceExampleRead",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "dataSourceExample")
		assert.Equal(t, expected, result)
	})
}

func TestExtractFactoryFunctionDetails_FrameworkResource(t *testing.T) {
	t.Run("extract methods from Framework resource factory", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

func newDirectoryBucketResource(ctx context.Context) (resource.ResourceWithConfigure, error) {
	r := &directoryBucketResource{}
	return r, nil
}

type directoryBucketResource struct {
	framework.ResourceWithModel[directoryBucketResourceModel]
	framework.WithImportByID
}

func (r *directoryBucketResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3_directory_bucket"
}

func (r *directoryBucketResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	// Schema definition here
}

func (r *directoryBucketResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Create implementation here
}

func (r *directoryBucketResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Read implementation here
}

func (r *directoryBucketResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// Update implementation here
}

func (r *directoryBucketResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Delete implementation here
}

type directoryBucketResourceModel struct {
	Bucket string `+"`"+`tfsdk:"bucket"`+"`"+`
}`

		expected := &AWSFactoryCRUDMethods{
			SchemaMethod: "Schema",
			CreateMethod: "Create",
			ReadMethod:   "Read",
			UpdateMethod: "Update",
			DeleteMethod: "Delete",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "newDirectoryBucketResource")
		assert.Equal(t, expected, result)
	})

	t.Run("extract methods with partial implementation", func(t *testing.T) {
		source := `package test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func newExampleResource(ctx context.Context) (resource.ResourceWithConfigure, error) {
	r := &exampleResource{}
	return r, nil
}

type exampleResource struct{}

func (r *exampleResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_example"
}

func (r *exampleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	// Schema definition here
}

func (r *exampleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Create implementation here
}

func (r *exampleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Read implementation here
}

// No Update or Delete methods`

		expected := &AWSFactoryCRUDMethods{
			SchemaMethod: "Schema",
			CreateMethod: "Create",
			ReadMethod:   "Read",
			UpdateMethod: "", // Should be empty
			DeleteMethod: "", // Should be empty
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "newExampleResource")
		assert.Equal(t, expected, result)
	})
}

func TestExtractFactoryFunctionDetails_FrameworkDataSource(t *testing.T) {
	t.Run("extract methods from Framework data source factory", func(t *testing.T) {
		source := `package s3

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

func newBucketDataSource(ctx context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &bucketDataSource{}
	return d, nil
}

type bucketDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *bucketDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_s3_bucket"
}

func (d *bucketDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	// Schema definition here
}

func (d *bucketDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Read implementation here
}

func (d *bucketDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Configure implementation here
}`

		expected := &AWSFactoryCRUDMethods{
			SchemaMethod:    "Schema",
			ReadMethod:      "Read",
			ConfigureMethod: "Configure",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "newBucketDataSource")
		assert.Equal(t, expected, result)
	})

	t.Run("extract methods with minimal implementation", func(t *testing.T) {
		source := `package test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func newMinimalDataSource(ctx context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &minimalDataSource{}
	return d, nil
}

type minimalDataSource struct{}

func (d *minimalDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_minimal"
}

func (d *minimalDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	// Schema definition here
}

func (d *minimalDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// Read implementation here
}

// No Configure method`

		expected := &AWSFactoryCRUDMethods{
			SchemaMethod:    "Schema",
			ReadMethod:      "Read",
			ConfigureMethod: "", // Should be empty
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "newMinimalDataSource")
		assert.Equal(t, expected, result)
	})
}

func TestExtractFactoryFunctionDetails_EphemeralResource(t *testing.T) {
	t.Run("extract methods from Ephemeral resource factory", func(t *testing.T) {
		source := `package keyvault

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

func NewKeyVaultCertificateEphemeralResource(ctx context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	e := &keyVaultCertificateEphemeralResource{}
	return e, nil
}

type keyVaultCertificateEphemeralResource struct {
	framework.EphemeralResourceWithConfigure
}

func (e *keyVaultCertificateEphemeralResource) Metadata(ctx context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	response.TypeName = "aws_key_vault_certificate"
}

func (e *keyVaultCertificateEphemeralResource) Schema(ctx context.Context, request ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	// Schema definition here
}

func (e *keyVaultCertificateEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	// Open implementation here
}

func (e *keyVaultCertificateEphemeralResource) Renew(ctx context.Context, request ephemeral.RenewRequest, response *ephemeral.RenewResponse) {
	// Renew implementation here
}

func (e *keyVaultCertificateEphemeralResource) Close(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) {
	// Close implementation here
}

func (e *keyVaultCertificateEphemeralResource) Configure(ctx context.Context, request ephemeral.ConfigureRequest, response *ephemeral.ConfigureResponse) {
	// Configure implementation here
}`

		expected := &AWSFactoryCRUDMethods{
			SchemaMethod:    "Schema",
			OpenMethod:      "Open",
			RenewMethod:     "Renew",
			CloseMethod:     "Close",
			ConfigureMethod: "Configure",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "NewKeyVaultCertificateEphemeralResource")
		assert.Equal(t, expected, result)
	})

	t.Run("extract methods with partial lifecycle", func(t *testing.T) {
		source := `package test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

func NewBasicEphemeralResource(ctx context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	e := &basicEphemeralResource{}
	return e, nil
}

type basicEphemeralResource struct{}

func (e *basicEphemeralResource) Metadata(ctx context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	response.TypeName = "aws_basic_ephemeral"
}

func (e *basicEphemeralResource) Schema(ctx context.Context, request ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	// Schema definition here
}

func (e *basicEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	// Open implementation here
}

func (e *basicEphemeralResource) Close(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) {
	// Close implementation here
}

// No Renew or Configure methods`

		expected := &AWSFactoryCRUDMethods{
			SchemaMethod:    "Schema",
			OpenMethod:      "Open",
			RenewMethod:     "", // Should be empty
			CloseMethod:     "Close",
			ConfigureMethod: "", // Should be empty
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "NewBasicEphemeralResource")
		assert.Equal(t, expected, result)
	})
}

func TestExtractFactoryFunctionDetails_ComplexPatterns(t *testing.T) {
	t.Run("extract from factory with variable assignment", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWithVariableAssignment() *schema.Resource {
	resource := &schema.Resource{
		CreateWithoutTimeout: resourceWithVariableAssignmentCreate,
		ReadWithoutTimeout:   resourceWithVariableAssignmentRead,
		UpdateWithoutTimeout: resourceWithVariableAssignmentUpdate,
		DeleteWithoutTimeout: resourceWithVariableAssignmentDelete,
		
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
		},
	}
	return resource
}`

		expected := &AWSFactoryCRUDMethods{
			CreateMethod: "resourceWithVariableAssignmentCreate",
			ReadMethod:   "resourceWithVariableAssignmentRead",
			UpdateMethod: "resourceWithVariableAssignmentUpdate",
			DeleteMethod: "resourceWithVariableAssignmentDelete",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceWithVariableAssignment")
		assert.Equal(t, expected, result)
	})

	t.Run("extract with mixed field variants", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMixedFields() *schema.Resource {
	return &schema.Resource{
		CreateContext:        resourceMixedFieldsCreateContext,
		Read:                 resourceMixedFieldsRead,
		UpdateWithoutTimeout: resourceMixedFieldsUpdate,
		DeleteContext:        resourceMixedFieldsDeleteContext,
		
		Schema: map[string]*schema.Schema{
			"id": {Type: schema.TypeString, Required: true},
		},
	}
}`

		expected := &AWSFactoryCRUDMethods{
			CreateMethod: "resourceMixedFieldsCreateContext",
			ReadMethod:   "resourceMixedFieldsRead",
			UpdateMethod: "resourceMixedFieldsUpdate",
			DeleteMethod: "resourceMixedFieldsDeleteContext",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceMixedFields")
		assert.Equal(t, expected, result)
	})

	t.Run("factory with no CRUD methods", func(t *testing.T) {
		source := `package test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNoCRUD() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"read_only": {Type: schema.TypeString, Computed: true},
		},
	}
}`

		expected := &AWSFactoryCRUDMethods{
			CreateMethod: "",
			ReadMethod:   "",
			UpdateMethod: "",
			DeleteMethod: "",
		}

		node, err := parseSource(source)
		require.NoError(t, err)

		result := extractFactoryFunctionDetails(node, "resourceNoCRUD")
		assert.Equal(t, expected, result)
	})
}

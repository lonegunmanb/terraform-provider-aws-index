package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
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
}`

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Find the resourceBucket function
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			fmt.Printf("Found function: %s\n", funcDecl.Name.Name)
			if funcDecl.Name.Name == "resourceBucket" {
				fmt.Println("Found resourceBucket function!")
				
				// Look for return statements
				ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
					if ret, ok := n.(*ast.ReturnStmt); ok {
						fmt.Printf("Found return statement with %d results\n", len(ret.Results))
						for i, result := range ret.Results {
							fmt.Printf("Result %d type: %T\n", i, result)
							
							// Handle &schema.Resource{...} pattern
							if unaryExpr, ok := result.(*ast.UnaryExpr); ok {
								fmt.Printf("Found unary expression with op: %s\n", unaryExpr.Op.String())
								if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
									fmt.Printf("Found composite literal with %d elements\n", len(compositeLit.Elts))
									for j, elt := range compositeLit.Elts {
										fmt.Printf("Element %d type: %T\n", j, elt)
										if keyValue, ok := elt.(*ast.KeyValueExpr); ok {
											if ident, ok := keyValue.Key.(*ast.Ident); ok {
												fmt.Printf("  Key: %s\n", ident.Name)
											}
											if ident, ok := keyValue.Value.(*ast.Ident); ok {
												fmt.Printf("  Value: %s\n", ident.Name)
											}
										}
									}
								}
							}
							
							// Original composite literal check
							if compositeLit, ok := result.(*ast.CompositeLit); ok {
								fmt.Printf("Found composite literal with %d elements\n", len(compositeLit.Elts))
								for j, elt := range compositeLit.Elts {
									fmt.Printf("Element %d type: %T\n", j, elt)
									if keyValue, ok := elt.(*ast.KeyValueExpr); ok {
										if ident, ok := keyValue.Key.(*ast.Ident); ok {
											fmt.Printf("  Key: %s\n", ident.Name)
										}
										if ident, ok := keyValue.Value.(*ast.Ident); ok {
											fmt.Printf("  Value: %s\n", ident.Name)
										}
									}
								}
							}
						}
					}
					return true
				})
			}
		}
	}
}

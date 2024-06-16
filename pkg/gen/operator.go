package gen

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceGen struct {
	template string

	schema *schema.Schema
}

func NewResourceGen(template string, schema *schema.Schema) *ResourceGen {
	return &ResourceGen{
		template: template,
		schema:   schema,
	}
}

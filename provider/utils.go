package provider

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func validateNotBlank(i interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	value := i.(string)

	if value != "" {
		return diags
	}

	return append(diags, diag.Diagnostic{
		Severity:      diag.Error,
		Summary:       "Empty string",
		Detail:        "Empty string",
		AttributePath: path,
	})
}

package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// stringPtrToValue maps a nullable API string to a framework value, treating
// both nil and "" as null so optional attributes don't perpetually diff.
func stringPtrToValue(p *string) types.String {
	if p == nil || *p == "" {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// stringValueOrNull maps a non-pointer API string to null when empty.
func stringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func boolPtrToValue(p *bool) types.Bool {
	if p == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*p)
}

func int64PtrToValue(p *int64) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*p)
}
